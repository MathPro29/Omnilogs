package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"omnilogs-api/dto"
	"omnilogs-api/internal/environment/repository"
	"omnilogs-api/models"

	"gorm.io/gorm"
)

var (
	ErrNotFound  = errors.New("environment resource not found")
	ErrForbidden = errors.New("environment access denied")
	ErrConflict  = errors.New("environment resource already exists")
	ErrInvalid   = errors.New("invalid environment relationship")
)

type Actor struct {
	UserID        int
	PlatformAdmin bool
}

type AccessTarget struct {
	ProductID  int
	ProjectID  *int
	CategoryID *int
}

type Usecase interface {
	CreateEnvironment(Actor, int, dto.CreateEnvironmentRequest) (*models.ProductEnvironment, error)
	ListEnvironments(Actor, int) ([]models.ProductEnvironment, error)
	UpdateEnvironment(Actor, int, int, dto.UpdateEnvironmentRequest) (*models.ProductEnvironment, error)
}

type usecase struct {
	repo repository.Repository
}

func NewUsecase(repo repository.Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) CreateEnvironment(actor Actor, productID int, req dto.CreateEnvironmentRequest) (*models.ProductEnvironment, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ENVIRONMENT", "CREATE"); err != nil {
		return nil, err
	}
	if req.ProductID != nil && *req.ProductID != productID {
		return nil, ErrInvalid
	}
	value := &models.ProductEnvironment{
		ProductID:       productID,
		EnvironmentCode: normalizeCode(req.EnvironmentCode),
		EnvironmentName: strings.TrimSpace(req.EnvironmentName),
	}
	if value.EnvironmentCode == "" || value.EnvironmentName == "" {
		return nil, ErrInvalid
	}
	if err := u.repo.DB().Create(value).Error; err != nil {
		return nil, classifyDBError(err)
	}
	return value, nil
}

func (u *usecase) ListEnvironments(actor Actor, productID int) ([]models.ProductEnvironment, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ENVIRONMENT", "READ"); err != nil {
		return nil, err
	}
	var values []models.ProductEnvironment
	return values, u.repo.DB().Where("product_id = ?", productID).Order("environment_id").Find(&values).Error
}

func (u *usecase) UpdateEnvironment(actor Actor, productID, id int, req dto.UpdateEnvironmentRequest) (*models.ProductEnvironment, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ENVIRONMENT", "UPDATE"); err != nil {
		return nil, err
	}
	if req.EnvironmentName != nil {
		name := strings.TrimSpace(*req.EnvironmentName)
		if name == "" {
			return nil, ErrInvalid
		}
		res := u.repo.DB().Model(&models.ProductEnvironment{}).Where("environment_id = ? AND product_id = ?", id, productID).Update("environment_name", name)
		if res.Error != nil {
			return nil, classifyDBError(res.Error)
		}
		if res.RowsAffected == 0 {
			return nil, ErrNotFound
		}
	}
	var value models.ProductEnvironment
	if err := u.repo.DB().Where("environment_id = ? AND product_id = ?", id, productID).First(&value).Error; err != nil {
		return nil, classifyDBError(err)
	}
	return &value, nil
}

func (u *usecase) authorize(actor Actor, target AccessTarget, resource, action string) error {
	if actor.PlatformAdmin {
		return nil
	}
	if target.ProductID <= 0 || !u.exists(&models.Product{}, "product_id = ?", target.ProductID) {
		return ErrNotFound
	}
	var memberships []models.ProductMembership
	if err := u.repo.DB().Where("user_id = ? AND product_id = ? AND is_active = TRUE AND (expires_at IS NULL OR expires_at > ?)", actor.UserID, target.ProductID, time.Now()).Find(&memberships).Error; err != nil {
		return err
	}
	if len(memberships) == 0 {
		return ErrForbidden
	}
	var rules []models.UserRolePermissionRule
	if err := u.repo.DB().Where("user_id = ? AND resource_type = ? AND action = ? AND is_active = TRUE AND (expires_at IS NULL OR expires_at > ?) AND (product_id IS NULL OR product_id = ?)", actor.UserID, strings.ToUpper(resource), strings.ToUpper(action), time.Now(), target.ProductID).Find(&rules).Error; err != nil {
		return err
	}
	allowedByRule := false
	for _, membership := range memberships {
		if !u.scopeAllows(membership.MembershipID, target, strings.EqualFold(action, "READ")) {
			continue
		}
		for _, rule := range rules {
			if !ruleMatches(rule, membership.RoleID, target) {
				continue
			}
			if rule.Effect == "DENY" {
				return ErrForbidden
			}
			if rule.Effect == "ALLOW" {
				allowedByRule = true
			}
		}
		if allowedByRule || (strings.EqualFold(action, "READ") && membershipReadAllowed(resource)) {
			return nil
		}
		var role models.ProductRole
		if err := u.repo.DB().Where("role_id = ? AND product_id = ?", membership.RoleID, target.ProductID).First(&role).Error; err == nil && permissionJSONAllows(role.Permissions, resource, action) {
			return nil
		}
	}
	return ErrForbidden
}

func (u *usecase) scopeAllows(membershipID int, target AccessTarget, allowParentRead bool) bool {
	var scopes []models.ProductMembershipScope
	if err := u.repo.DB().Where("membership_id = ? AND product_id = ? AND is_active = TRUE", membershipID, target.ProductID).Find(&scopes).Error; err != nil {
		return false
	}
	for _, scope := range scopes {
		if allowParentRead && target.ProjectID == nil && target.CategoryID == nil {
			return true
		}
		switch scope.ScopeLevel {
		case "PRODUCT":
			return true
		case "PROJECT":
			if target.ProjectID != nil && scope.ProjectID != nil && *target.ProjectID == *scope.ProjectID {
				return true
			}
		case "CATEGORY":
			if target.CategoryID != nil && scope.CategoryID != nil && *target.CategoryID == *scope.CategoryID {
				return true
			}
			if target.CategoryID != nil && scope.CategoryID != nil {
				var feature models.ProjectFeature
				if err := u.repo.DB().Select("path_ids").First(&feature, *target.CategoryID).Error; err == nil && pathContains(feature.PathIDs, *scope.CategoryID) {
					return true
				}
			}
			if allowParentRead && target.ProjectID != nil && scope.ProjectID != nil && *target.ProjectID == *scope.ProjectID {
				return true
			}
		}
	}
	return false
}

func (u *usecase) exists(model any, query string, args ...any) bool {
	return existsDB(u.repo.DB(), model, query, args...)
}

func existsDB(db *gorm.DB, model any, query string, args ...any) bool {
	var count int64
	return db.Model(model).Where(query, args...).Count(&count).Error == nil && count > 0
}

func ruleMatches(rule models.UserRolePermissionRule, roleID int, target AccessTarget) bool {
	if rule.RoleID != nil && *rule.RoleID != roleID {
		return false
	}
	if rule.ProjectID != nil && (target.ProjectID == nil || *rule.ProjectID != *target.ProjectID) {
		return false
	}
	if rule.CategoryID != nil && (target.CategoryID == nil || *rule.CategoryID != *target.CategoryID) {
		return false
	}
	return true
}

func permissionJSONAllows(raw json.RawMessage, resource, action string) bool {
	var value map[string]any
	if json.Unmarshal(raw, &value) != nil {
		return false
	}
	if all, ok := value["all"].(bool); ok && all {
		return true
	}
	resource, action = strings.ToUpper(resource), strings.ToUpper(action)
	for _, key := range []string{resource, strings.ToLower(resource)} {
		entry, ok := value[key]
		if !ok {
			continue
		}
		switch typed := entry.(type) {
		case bool:
			if typed {
				return true
			}
		case []any:
			for _, item := range typed {
				if strings.EqualFold(fmt.Sprint(item), action) {
					return true
				}
			}
		case map[string]any:
			for k, item := range typed {
				if strings.EqualFold(k, action) {
					allowed, _ := item.(bool)
					return allowed
				}
			}
		}
	}
	for key, item := range value {
		if strings.EqualFold(key, resource+"."+action) {
			allowed, _ := item.(bool)
			return allowed
		}
	}
	return false
}

func membershipReadAllowed(resource string) bool {
	switch strings.ToUpper(resource) {
	case "PRODUCT", "PROJECT", "FEATURE", "CATEGORY", "ENVIRONMENT", "LOG":
		return true
	default:
		return false
	}
}

func pathContains(path *string, id int) bool {
	if path == nil {
		return false
	}
	needle := "," + fmt.Sprint(id) + ","
	return strings.Contains(","+*path+",", needle)
}

func normalizeCode(value string) string { return strings.ToLower(strings.TrimSpace(value)) }

func classifyDBError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "duplicate key") || strings.Contains(message, "unique constraint") {
		return ErrConflict
	}
	return err
}

package usecase

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"omnilogs-api/dto"
	"omnilogs-api/internal/scopes/repository"
	"omnilogs-api/models"
	"omnilogs-api/responses"

	"gorm.io/gorm"
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
	CreateScope(Actor, int, int, dto.CreateMembershipScopeRequest) (*models.ProductMembershipScope, error)
	ListScopes(Actor, int, int) ([]models.ProductMembershipScope, error)
	UpdateScope(Actor, int, int, int, dto.UpdateMembershipScopeRequest) (*models.ProductMembershipScope, error)
	DeleteScope(Actor, int, int, int) (*models.ProductMembershipScope, error)
}

type usecase struct {
	repo repository.Repository
	db   *gorm.DB
}

func NewUsecase(repo repository.Repository, db *gorm.DB) Usecase {
	return &usecase{repo: repo, db: db}
}

func (u *usecase) CreateScope(actor Actor, productID, membershipID int, req dto.CreateMembershipScopeRequest) (*models.ProductMembershipScope, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ACCESS", "GRANT"); err != nil {
		return nil, err
	}
	if !u.exists(&models.ProductMembership{}, "membership_id = ? AND product_id = ?", membershipID, productID) || !validScope(u.db, productID, req.ScopeLevel, req.ProjectID, req.CategoryID) {
		return nil, responses.ErrorScopeCode["INVALID_SCOPE"]
	}
	var existing models.ProductMembershipScope
	err := u.db.Where(
		"membership_id = ? AND product_id = ? AND scope_level = ? AND ((project_id IS NULL AND ?::int IS NULL) OR project_id = ?) AND ((category_id IS NULL AND ?::int IS NULL) OR category_id = ?)",
		membershipID,
		productID,
		req.ScopeLevel,
		req.ProjectID,
		req.ProjectID,
		req.CategoryID,
		req.CategoryID,
	).Limit(1).Find(&existing).Error
	if err != nil {
		return nil, classifyDBError(err)
	}

	if existing.ScopeID != 0 {
		if !existing.IsActive {
			existing.IsActive = true
			if err := u.repo.UpdateScope(&existing); err != nil {
				return nil, classifyDBError(err)
			}
		}
		return &existing, nil
	}
	value := &models.ProductMembershipScope{
		MembershipID: membershipID,
		ProductID:    productID,
		ProjectID:    req.ProjectID,
		CategoryID:   req.CategoryID,
		ScopeLevel:   req.ScopeLevel,
		IsActive:     true,
	}
	if err := u.repo.CreateScope(value); err != nil {
		return nil, classifyDBError(err)
	}
	return value, nil
}

func (u *usecase) ListScopes(actor Actor, productID, membershipID int) ([]models.ProductMembershipScope, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ACCESS", "READ"); err != nil {
		return nil, err
	}
	scopes, err := u.repo.ListScopes(productID, membershipID)
	if err != nil {
		return nil, classifyDBError(err)
	}
	return scopes, nil
}

func (u *usecase) UpdateScope(actor Actor, productID, membershipID, id int, req dto.UpdateMembershipScopeRequest) (*models.ProductMembershipScope, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ACCESS", "GRANT"); err != nil {
		return nil, err
	}
	if !validScope(u.db, productID, req.ScopeLevel, req.ProjectID, req.CategoryID) {
		return nil, responses.ErrorScopeCode["INVALID_SCOPE"]
	}
	scope, err := u.repo.GetScopeByID(id)
	if err != nil {
		return nil, classifyDBError(err)
	}
	if scope.ProductID != productID || scope.MembershipID != membershipID {
		return nil, responses.ErrorScopeCode["FORBIDDEN"]
	}
	scope.ScopeLevel = req.ScopeLevel
	scope.ProjectID = req.ProjectID
	scope.CategoryID = req.CategoryID
	if err := u.repo.UpdateScope(scope); err != nil {
		return nil, classifyDBError(err)
	}
	return scope, nil
}

func (u *usecase) DeleteScope(actor Actor, productID, membershipID, id int) (*models.ProductMembershipScope, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ACCESS", "GRANT"); err != nil {
		return nil, err
	}
	scope, err := u.repo.GetScopeByID(id)
	if err != nil {
		return nil, classifyDBError(err)
	}
	if scope.ProductID != productID || scope.MembershipID != membershipID {
		return nil, responses.ErrorScopeCode["FORBIDDEN"]
	}
	if err := u.repo.DeleteScope(scope); err != nil {
		return nil, classifyDBError(err)
	}
	return scope, nil
}

func (u *usecase) authorize(actor Actor, target AccessTarget, resource, action string) error {
	var isGlobalAdmin bool
	var pm models.PlatformMembership
	if err := u.db.Where("user_id = ? AND is_active = TRUE AND platform_role_id IN (1, 3)", actor.UserID).Limit(1).Find(&pm).Error; err == nil && pm.PlatformMembershipID != 0 {
		isGlobalAdmin = true
	}
	if isGlobalAdmin {
		return nil
	}
	if target.ProductID <= 0 || !u.exists(&models.Product{}, "product_id = ?", target.ProductID) {
		return responses.ErrorScopeCode["NOT_FOUND"]
	}
	var memberships []models.ProductMembership
	if err := u.db.Where("user_id = ? AND product_id = ? AND is_active = TRUE AND (expires_at IS NULL OR expires_at > ?)", actor.UserID, target.ProductID, time.Now()).Find(&memberships).Error; err != nil {
		return err
	}
	if len(memberships) == 0 {
		return responses.ErrorScopeCode["FORBIDDEN"]
	}
	var rules []models.UserRolePermissionRule
	if err := u.db.Where("user_id = ? AND resource_type = ? AND action = ? AND is_active = TRUE AND (expires_at IS NULL OR expires_at > ?) AND (product_id IS NULL OR product_id = ?)", actor.UserID, strings.ToUpper(resource), strings.ToUpper(action), time.Now(), target.ProductID).Find(&rules).Error; err != nil {
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
				return responses.ErrorScopeCode["FORBIDDEN"]
			}
			if rule.Effect == "ALLOW" {
				allowedByRule = true
			}
		}
		if allowedByRule || (strings.EqualFold(action, "READ") && membershipReadAllowed(resource)) {
			return nil
		}
		var role models.ProductRole
		if err := u.db.Preload("Permissions").Where("role_id = ? AND product_id = ?", membership.RoleID, target.ProductID).First(&role).Error; err == nil {
			if strings.ToUpper(resource) == "FEATURE" || strings.ToUpper(resource) == "CATEGORY" {
				if role.RoleCode != "owner" {
					continue
				}
			}
			if permissionListAllows(role.Permissions, resource, action) {
				return nil
			}
		}
	}
	return responses.ErrorScopeCode["FORBIDDEN"]
}

func (u *usecase) scopeAllows(membershipID int, target AccessTarget, allowParentRead bool) bool {
	var scopes []models.ProductMembershipScope
	if err := u.db.Where("membership_id = ? AND product_id = ? AND is_active = TRUE", membershipID, target.ProductID).Find(&scopes).Error; err != nil {
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
				if err := u.db.Select("path_ids").First(&feature, *target.CategoryID).Error; err == nil && pathContains(feature.PathIDs, *scope.CategoryID) {
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
	return existsDB(u.db, model, query, args...)
}

func existsDB(db *gorm.DB, model any, query string, args ...any) bool {
	var count int64
	return db.Model(model).Where(query, args...).Count(&count).Error == nil && count > 0
}

func validScope(db *gorm.DB, productID int, level string, projectID, categoryID *int) bool {
	switch level {
	case "PRODUCT":
		return projectID == nil && categoryID == nil
	case "PROJECT":
		return projectID != nil && categoryID == nil && existsDB(db, &models.Project{}, "project_id = ? AND product_id = ?", *projectID, productID)
	case "CATEGORY":
		if projectID == nil || categoryID == nil {
			return false
		}
		return existsDB(db, &models.ProjectFeature{}, "category_id = ? AND project_id = ? AND product_id = ?", *categoryID, *projectID, productID)
	default:
		return false
	}
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

func permissionListAllows(values []models.ProductRolePermission, resource, action string) bool {
	resource, action = strings.ToUpper(resource), strings.ToUpper(action)
	for _, value := range values {
		if strings.EqualFold(value.ResourceType, resource) && strings.EqualFold(value.Action, action) {
			return true
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

func classifyDBError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return responses.ErrorScopeCode["NOT_FOUND"]
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "duplicate key") || strings.Contains(message, "unique constraint") {
		return responses.ErrorScopeCode["CONFLICT"]
	}
	return err
}

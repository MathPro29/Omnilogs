package usecase

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"omnilogs-api/dto"
	"omnilogs-api/internal/product/repository"
	"omnilogs-api/models"

	"gorm.io/gorm"
)

var (
	ErrNotFound  = errors.New("product resource not found")
	ErrForbidden = errors.New("product access denied")
	ErrConflict  = errors.New("product resource already exists")
	ErrInvalid   = errors.New("invalid product relationship")
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
	CreateProduct(Actor, dto.CreateProductRequest) (*models.Product, error)
	ListProducts(Actor) ([]models.Product, error)
	GetProduct(Actor, int) (*models.Product, error)
	UpdateProduct(Actor, int, dto.UpdateProductRequest) (*models.Product, error)
	CreateEnvironment(Actor, int, dto.CreateEnvironmentRequest) (*models.ProductEnvironment, error)
	ListEnvironments(Actor, int) ([]models.ProductEnvironment, error)
	UpdateEnvironment(Actor, int, int, dto.UpdateEnvironmentRequest) (*models.ProductEnvironment, error)
	CreateAPIKey(Actor, int, dto.CreateAPIKeyRequest) (*dto.CreateAPIKeyResponse, error)
	ListAPIKeys(Actor, int) ([]dto.APIKeyResponse, error)
	UpdateAPIKey(Actor, int, int, dto.UpdateAPIKeyRequest) (*dto.APIKeyResponse, error)
	RevokeAPIKey(Actor, int, int) error
	CreateProject(Actor, int, dto.CreateProjectRequest) (*models.Project, error)
	ListProjects(Actor, int) ([]models.Project, error)
	GetProject(Actor, int, int) (*models.Project, error)
	UpdateProject(Actor, int, int, dto.UpdateProjectRequest) (*models.Project, error)
	CreateFeature(Actor, int, int, dto.CreateProjectFeatureRequest) (*models.ProjectFeature, error)
	ListFeatures(Actor, int, int) ([]models.ProjectFeature, error)
	UpdateFeature(Actor, int, int, int, dto.UpdateProjectFeatureRequest) (*models.ProjectFeature, error)
	CreateRole(Actor, int, dto.CreateProductRoleRequest) (*models.ProductRole, error)
	ListRoles(Actor, int) ([]models.ProductRole, error)
	UpdateRole(Actor, int, int, dto.UpdateRoleRequest) (*models.ProductRole, error)
	CreateMembership(Actor, int, dto.CreateProductMembershipRequest) (*models.ProductMembership, error)
	ListMemberships(Actor, int) ([]models.ProductMembership, error)
	UpdateMembership(Actor, int, int, dto.UpdateProductMembershipRequest) (*models.ProductMembership, error)
	CreateScope(Actor, int, int, dto.CreateMembershipScopeRequest) (*models.ProductMembershipScope, error)
	ListScopes(Actor, int, int) ([]models.ProductMembershipScope, error)
	UpdateScope(Actor, int, int, int, dto.UpdateMembershipScopeRequest) (*models.ProductMembershipScope, error)
	CreatePermissionRule(Actor, dto.CreatePermissionRuleRequest) (*models.UserRolePermissionRule, error)
	ListPermissionRules(Actor, int) ([]models.UserRolePermissionRule, error)
	UpdatePermissionRule(Actor, int, dto.UpdatePermissionRuleRequest) (*models.UserRolePermissionRule, error)
	CheckPermission(Actor, dto.PermissionCheckRequest) (*dto.PermissionCheckResponse, error)
}

type usecase struct{ repository repository.Repository }

func NewUsecase(repository repository.Repository) Usecase { return &usecase{repository: repository} }

func (u *usecase) CreateProduct(actor Actor, req dto.CreateProductRequest) (*models.Product, error) {
	if !actor.PlatformAdmin {
		return nil, ErrForbidden
	}
	product := &models.Product{ProductName: strings.TrimSpace(req.ProductName), ProductCode: normalizeCode(req.ProductCode), IsActive: true}
	if product.ProductName == "" || product.ProductCode == "" {
		return nil, ErrInvalid
	}
	err := u.repository.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(product).Error; err != nil {
			return classifyDBError(err)
		}
		for _, environmentRequest := range req.Environments {
			if environmentRequest.ProductID != nil && *environmentRequest.ProductID != product.ProductID {
				return ErrInvalid
			}
			environment := models.ProductEnvironment{
				ProductID:       product.ProductID,
				EnvironmentCode: normalizeCode(environmentRequest.EnvironmentCode),
				EnvironmentName: strings.TrimSpace(environmentRequest.EnvironmentName),
			}
			if environment.EnvironmentCode == "" || environment.EnvironmentName == "" {
				return ErrInvalid
			}
			if err := tx.Create(&environment).Error; err != nil {
				return classifyDBError(err)
			}
			product.ProductEnvironments = append(product.ProductEnvironments, environment)
		}
		role := models.ProductRole{ProductID: product.ProductID, RoleCode: "product_owner", RoleName: "Product Owner", Permissions: json.RawMessage(`{"all":true}`)}
		if err := tx.Create(&role).Error; err != nil {
			return classifyDBError(err)
		}
		membership := models.ProductMembership{UserID: actor.UserID, ProductID: product.ProductID, RoleID: role.RoleID, IsActive: true}
		if err := tx.Create(&membership).Error; err != nil {
			return classifyDBError(err)
		}
		scope := models.ProductMembershipScope{MembershipID: membership.MembershipID, ProductID: product.ProductID, ScopeLevel: "PRODUCT", IsActive: true}
		return tx.Create(&scope).Error
	})
	return product, err
}

func (u *usecase) ListProducts(actor Actor) ([]models.Product, error) {
	var products []models.Product
	q := u.repository.DB().Model(&models.Product{}).Preload("ProductEnvironments").Order("product_id")
	if !actor.PlatformAdmin {
		q = q.Where(`EXISTS (SELECT 1 FROM product_memberships pm JOIN product_membership_scopes pms ON pms.membership_id = pm.membership_id WHERE pm.product_id = products.product_id AND pm.user_id = ? AND pm.is_active = TRUE AND pms.is_active = TRUE AND (pm.expires_at IS NULL OR pm.expires_at > ?))`, actor.UserID, time.Now())
	}
	return products, q.Find(&products).Error
}

func (u *usecase) GetProduct(actor Actor, productID int) (*models.Product, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "PRODUCT", "READ"); err != nil {
		return nil, err
	}
	var value models.Product
	if err := u.repository.DB().Preload("ProductEnvironments").First(&value, productID).Error; err != nil {
		return nil, classifyDBError(err)
	}
	return &value, nil
}

func (u *usecase) UpdateProduct(actor Actor, productID int, req dto.UpdateProductRequest) (*models.Product, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "PRODUCT", "UPDATE"); err != nil {
		return nil, err
	}
	updates := map[string]any{}
	if req.ProductName != nil {
		name := strings.TrimSpace(*req.ProductName)
		if name == "" {
			return nil, ErrInvalid
		}
		updates["product_name"] = name
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if len(updates) > 0 {
		if err := u.repository.DB().Model(&models.Product{}).Where("product_id = ?", productID).Updates(updates).Error; err != nil {
			return nil, classifyDBError(err)
		}
	}
	return u.GetProduct(actor, productID)
}

func (u *usecase) CreateEnvironment(actor Actor, productID int, req dto.CreateEnvironmentRequest) (*models.ProductEnvironment, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ENVIRONMENT", "CREATE"); err != nil {
		return nil, err
	}
	if req.ProductID != nil && *req.ProductID != productID {
		return nil, ErrInvalid
	}
	value := &models.ProductEnvironment{ProductID: productID, EnvironmentCode: normalizeCode(req.EnvironmentCode), EnvironmentName: strings.TrimSpace(req.EnvironmentName)}
	if value.EnvironmentCode == "" || value.EnvironmentName == "" {
		return nil, ErrInvalid
	}
	if err := u.repository.DB().Create(value).Error; err != nil {
		return nil, classifyDBError(err)
	}
	return value, nil
}

func (u *usecase) ListEnvironments(actor Actor, productID int) ([]models.ProductEnvironment, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ENVIRONMENT", "READ"); err != nil {
		return nil, err
	}
	var values []models.ProductEnvironment
	return values, u.repository.DB().Where("product_id = ?", productID).Order("environment_id").Find(&values).Error
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
		res := u.repository.DB().Model(&models.ProductEnvironment{}).Where("environment_id = ? AND product_id = ?", id, productID).Update("environment_name", name)
		if res.Error != nil {
			return nil, classifyDBError(res.Error)
		}
		if res.RowsAffected == 0 {
			return nil, ErrNotFound
		}
	}
	var value models.ProductEnvironment
	if err := u.repository.DB().Where("environment_id = ? AND product_id = ?", id, productID).First(&value).Error; err != nil {
		return nil, classifyDBError(err)
	}
	return &value, nil
}

func (u *usecase) CreateAPIKey(actor Actor, productID int, req dto.CreateAPIKeyRequest) (*dto.CreateAPIKeyResponse, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "API_KEY", "CREATE"); err != nil {
		return nil, err
	}
	if req.ProductID != 0 && req.ProductID != productID {
		return nil, ErrInvalid
	}
	if req.EnvironmentID != nil && !u.exists(&models.ProductEnvironment{}, "environment_id = ? AND product_id = ?", *req.EnvironmentID, productID) {
		return nil, ErrInvalid
	}
	if !validJSON(req.Permissions) {
		return nil, ErrInvalid
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	secret := "omni_" + hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(secret))
	prefix := secret[:13]
	value := &models.ProductAPIKey{ProductID: productID, EnvironmentID: req.EnvironmentID, KeyName: req.KeyName, KeyPrefix: prefix, KeyHash: hex.EncodeToString(sum[:]), Permissions: req.Permissions, IsActive: true, ExpiresAt: req.ExpiresAt}
	if err := u.repository.DB().Create(value).Error; err != nil {
		return nil, classifyDBError(err)
	}
	response := toAPIKeyResponse(value)
	return &dto.CreateAPIKeyResponse{APIKeyResponse: response, APIKey: secret}, nil
}

func (u *usecase) ListAPIKeys(actor Actor, productID int) ([]dto.APIKeyResponse, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "API_KEY", "READ"); err != nil {
		return nil, err
	}
	var values []models.ProductAPIKey
	if err := u.repository.DB().Where("product_id = ?", productID).Order("key_id").Find(&values).Error; err != nil {
		return nil, err
	}
	result := make([]dto.APIKeyResponse, 0, len(values))
	for i := range values {
		result = append(result, toAPIKeyResponse(&values[i]))
	}
	return result, nil
}

func (u *usecase) UpdateAPIKey(actor Actor, productID, id int, req dto.UpdateAPIKeyRequest) (*dto.APIKeyResponse, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "API_KEY", "UPDATE"); err != nil {
		return nil, err
	}
	updates := map[string]any{}
	if req.KeyName != nil {
		updates["key_name"] = req.KeyName
	}
	if len(req.Permissions) > 0 {
		if !validJSON(req.Permissions) {
			return nil, ErrInvalid
		}
		updates["permissions"] = req.Permissions
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.ExpiresAt != nil {
		updates["expires_at"] = req.ExpiresAt
	}
	res := u.repository.DB().Model(&models.ProductAPIKey{}).Where("key_id = ? AND product_id = ?", id, productID).Updates(updates)
	if res.Error != nil {
		return nil, classifyDBError(res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	var value models.ProductAPIKey
	if err := u.repository.DB().Where("key_id = ? AND product_id = ?", id, productID).First(&value).Error; err != nil {
		return nil, classifyDBError(err)
	}
	response := toAPIKeyResponse(&value)
	return &response, nil
}

func (u *usecase) RevokeAPIKey(actor Actor, productID, id int) error {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "API_KEY", "DELETE"); err != nil {
		return err
	}
	now := time.Now()
	res := u.repository.DB().Model(&models.ProductAPIKey{}).Where("key_id = ? AND product_id = ?", id, productID).Updates(map[string]any{"is_active": false, "revoked_at": &now})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (u *usecase) CreateProject(actor Actor, productID int, req dto.CreateProjectRequest) (*models.Project, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "PROJECT", "CREATE"); err != nil {
		return nil, err
	}
	if req.ProductID != 0 && req.ProductID != productID {
		return nil, ErrInvalid
	}
	value := &models.Project{ProductID: productID, ProjectCode: normalizeCode(req.ProjectCode), ProjectName: strings.TrimSpace(req.ProjectName), IsActive: true}
	if value.ProjectCode == "" || value.ProjectName == "" {
		return nil, ErrInvalid
	}
	if err := u.repository.DB().Create(value).Error; err != nil {
		return nil, classifyDBError(err)
	}
	return value, nil
}

func (u *usecase) ListProjects(actor Actor, productID int) ([]models.Project, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "PROJECT", "READ"); err != nil {
		return nil, err
	}
	var values []models.Project
	q := u.repository.DB().Where("product_id = ?", productID)
	if !actor.PlatformAdmin {
		q = q.Where(`EXISTS (SELECT 1 FROM product_memberships pm JOIN product_membership_scopes pms ON pms.membership_id = pm.membership_id WHERE pm.user_id = ? AND pm.product_id = projects.product_id AND pm.is_active = TRUE AND pms.is_active = TRUE AND (pm.expires_at IS NULL OR pm.expires_at > ?) AND (pms.scope_level = 'PRODUCT' OR (pms.scope_level IN ('PROJECT', 'CATEGORY') AND pms.project_id = projects.project_id)))`, actor.UserID, time.Now())
	}
	return values, q.Order("project_id").Find(&values).Error
}

func (u *usecase) GetProject(actor Actor, productID, id int) (*models.Project, error) {
	target := AccessTarget{ProductID: productID, ProjectID: &id}
	if err := u.authorize(actor, target, "PROJECT", "READ"); err != nil {
		return nil, err
	}
	var value models.Project
	if err := u.repository.DB().Where("project_id = ? AND product_id = ?", id, productID).First(&value).Error; err != nil {
		return nil, classifyDBError(err)
	}
	return &value, nil
}

func (u *usecase) UpdateProject(actor Actor, productID, id int, req dto.UpdateProjectRequest) (*models.Project, error) {
	target := AccessTarget{ProductID: productID, ProjectID: &id}
	if err := u.authorize(actor, target, "PROJECT", "UPDATE"); err != nil {
		return nil, err
	}
	updates := map[string]any{}
	if req.ProjectName != nil {
		name := strings.TrimSpace(*req.ProjectName)
		if name == "" {
			return nil, ErrInvalid
		}
		updates["project_name"] = name
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	res := u.repository.DB().Model(&models.Project{}).Where("project_id = ? AND product_id = ?", id, productID).Updates(updates)
	if res.Error != nil {
		return nil, classifyDBError(res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return u.GetProject(actor, productID, id)
}

func (u *usecase) CreateFeature(actor Actor, productID, projectID int, req dto.CreateProjectFeatureRequest) (*models.ProjectFeature, error) {
	target := AccessTarget{ProductID: productID, ProjectID: &projectID}
	if err := u.authorize(actor, target, "FEATURE", "CREATE"); err != nil {
		return nil, err
	}
	if (req.ProductID != 0 && req.ProductID != productID) || (req.ProjectID != 0 && req.ProjectID != projectID) || !u.exists(&models.Project{}, "project_id = ? AND product_id = ?", projectID, productID) {
		return nil, ErrInvalid
	}
	value := &models.ProjectFeature{ProductID: productID, ProjectID: projectID, ParentID: req.ParentID, CategoryType: req.CategoryType, CategoryCode: normalizeCode(req.CategoryCode), CategoryName: strings.TrimSpace(req.CategoryName), IsActive: true}
	if value.CategoryCode == "" || value.CategoryName == "" {
		return nil, ErrInvalid
	}
	err := u.repository.Transaction(func(tx *gorm.DB) error {
		if value.ParentID != nil && !existsDB(tx, &models.ProjectFeature{}, "category_id = ? AND product_id = ? AND project_id = ?", *value.ParentID, productID, projectID) {
			return ErrInvalid
		}
		if err := tx.Create(value).Error; err != nil {
			return classifyDBError(err)
		}
		return rebuildFeaturePaths(tx, projectID)
	})
	if err != nil {
		return nil, err
	}
	return u.getFeature(productID, projectID, value.CategoryID)
}

func (u *usecase) ListFeatures(actor Actor, productID, projectID int) ([]models.ProjectFeature, error) {
	target := AccessTarget{ProductID: productID, ProjectID: &projectID}
	if err := u.authorize(actor, target, "FEATURE", "READ"); err != nil {
		return nil, err
	}
	var values []models.ProjectFeature
	q := u.repository.DB().Where("product_id = ? AND project_id = ?", productID, projectID)
	if !actor.PlatformAdmin {
		q = q.Where(`EXISTS (SELECT 1 FROM product_memberships pm JOIN product_membership_scopes pms ON pms.membership_id = pm.membership_id WHERE pm.user_id = ? AND pm.product_id = project_features.product_id AND pm.is_active = TRUE AND pms.is_active = TRUE AND (pm.expires_at IS NULL OR pm.expires_at > ?) AND (pms.scope_level = 'PRODUCT' OR (pms.scope_level = 'PROJECT' AND pms.project_id = project_features.project_id) OR (pms.scope_level = 'CATEGORY' AND pms.project_id = project_features.project_id AND (',' || project_features.path_ids || ',') LIKE ('%,' || pms.category_id::text || ',%'))))`, actor.UserID, time.Now())
	}
	return values, q.Order("level, category_id").Find(&values).Error
}

func (u *usecase) UpdateFeature(actor Actor, productID, projectID, id int, req dto.UpdateProjectFeatureRequest) (*models.ProjectFeature, error) {
	target := AccessTarget{ProductID: productID, ProjectID: &projectID, CategoryID: &id}
	if err := u.authorize(actor, target, "FEATURE", "UPDATE"); err != nil {
		return nil, err
	}
	err := u.repository.Transaction(func(tx *gorm.DB) error {
		var value models.ProjectFeature
		if err := tx.Where("category_id = ? AND product_id = ? AND project_id = ?", id, productID, projectID).First(&value).Error; err != nil {
			return classifyDBError(err)
		}
		if req.ParentID != nil {
			if *req.ParentID == id || !existsDB(tx, &models.ProjectFeature{}, "category_id = ? AND product_id = ? AND project_id = ?", *req.ParentID, productID, projectID) {
				return ErrInvalid
			}
			value.ParentID = req.ParentID
		}
		if req.CategoryType != nil {
			value.CategoryType = req.CategoryType
		}
		if req.CategoryName != nil {
			value.CategoryName = strings.TrimSpace(*req.CategoryName)
			if value.CategoryName == "" {
				return ErrInvalid
			}
		}
		if req.IsActive != nil {
			value.IsActive = *req.IsActive
		}
		if err := tx.Save(&value).Error; err != nil {
			return classifyDBError(err)
		}
		return rebuildFeaturePaths(tx, projectID)
	})
	if err != nil {
		return nil, err
	}
	return u.getFeature(productID, projectID, id)
}

func (u *usecase) CreateRole(actor Actor, productID int, req dto.CreateProductRoleRequest) (*models.ProductRole, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ROLE", "CREATE"); err != nil {
		return nil, err
	}
	if req.ProductID != 0 && req.ProductID != productID || !validJSON(req.Permissions) {
		return nil, ErrInvalid
	}
	if req.TemplateID != nil && !u.exists(&models.RoleTemplate{}, "template_id = ? AND is_active = TRUE", *req.TemplateID) {
		return nil, ErrInvalid
	}
	value := &models.ProductRole{ProductID: productID, TemplateID: req.TemplateID, RoleCode: normalizeCode(req.RoleCode), RoleName: strings.TrimSpace(req.RoleName), Permissions: req.Permissions}
	if err := u.repository.DB().Create(value).Error; err != nil {
		return nil, classifyDBError(err)
	}
	return value, nil
}

func (u *usecase) ListRoles(actor Actor, productID int) ([]models.ProductRole, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ROLE", "READ"); err != nil {
		return nil, err
	}
	var values []models.ProductRole
	return values, u.repository.DB().Where("product_id = ?", productID).Order("role_id").Find(&values).Error
}

func (u *usecase) UpdateRole(actor Actor, productID, id int, req dto.UpdateRoleRequest) (*models.ProductRole, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ROLE", "UPDATE"); err != nil {
		return nil, err
	}
	updates := map[string]any{}
	if req.RoleName != nil {
		updates["role_name"] = strings.TrimSpace(*req.RoleName)
	}
	if len(req.Permissions) > 0 {
		if !validJSON(req.Permissions) {
			return nil, ErrInvalid
		}
		updates["permissions"] = req.Permissions
	}
	res := u.repository.DB().Model(&models.ProductRole{}).Where("role_id = ? AND product_id = ?", id, productID).Updates(updates)
	if res.Error != nil {
		return nil, classifyDBError(res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	var value models.ProductRole
	if err := u.repository.DB().Where("role_id = ? AND product_id = ?", id, productID).First(&value).Error; err != nil {
		return nil, classifyDBError(err)
	}
	return &value, nil
}

func (u *usecase) CreateMembership(actor Actor, productID int, req dto.CreateProductMembershipRequest) (*models.ProductMembership, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ACCESS", "GRANT"); err != nil {
		return nil, err
	}
	if req.ProductID != 0 && req.ProductID != productID || !u.exists(&models.User{}, "user_id = ? AND deleted_at IS NULL", req.UserID) || !u.exists(&models.ProductRole{}, "role_id = ? AND product_id = ?", req.RoleID, productID) {
		return nil, ErrInvalid
	}
	value := &models.ProductMembership{UserID: req.UserID, ProductID: productID, RoleID: req.RoleID, ExpiresAt: req.ExpiresAt, IsActive: true}
	if err := u.repository.DB().Create(value).Error; err != nil {
		return nil, classifyDBError(err)
	}
	return value, nil
}

func (u *usecase) ListMemberships(actor Actor, productID int) ([]models.ProductMembership, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ACCESS", "READ"); err != nil {
		return nil, err
	}
	var values []models.ProductMembership
	return values, u.repository.DB().Where("product_id = ?", productID).Order("membership_id").Find(&values).Error
}

func (u *usecase) UpdateMembership(actor Actor, productID, id int, req dto.UpdateProductMembershipRequest) (*models.ProductMembership, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ACCESS", "GRANT"); err != nil {
		return nil, err
	}
	updates := map[string]any{}
	if req.RoleID != nil {
		if !u.exists(&models.ProductRole{}, "role_id = ? AND product_id = ?", *req.RoleID, productID) {
			return nil, ErrInvalid
		}
		updates["role_id"] = *req.RoleID
	}
	if req.ExpiresAt != nil {
		updates["expires_at"] = req.ExpiresAt
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	res := u.repository.DB().Model(&models.ProductMembership{}).Where("membership_id = ? AND product_id = ?", id, productID).Updates(updates)
	if res.Error != nil {
		return nil, classifyDBError(res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	var value models.ProductMembership
	if err := u.repository.DB().Where("membership_id = ? AND product_id = ?", id, productID).First(&value).Error; err != nil {
		return nil, classifyDBError(err)
	}
	return &value, nil
}

func (u *usecase) CreateScope(actor Actor, productID, membershipID int, req dto.CreateMembershipScopeRequest) (*models.ProductMembershipScope, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ACCESS", "GRANT"); err != nil {
		return nil, err
	}
	if !u.exists(&models.ProductMembership{}, "membership_id = ? AND product_id = ?", membershipID, productID) || !validScope(u.repository.DB(), productID, req.ScopeLevel, req.ProjectID, req.CategoryID) {
		return nil, ErrInvalid
	}
	value := &models.ProductMembershipScope{MembershipID: membershipID, ProductID: productID, ProjectID: req.ProjectID, CategoryID: req.CategoryID, ScopeLevel: req.ScopeLevel, IsActive: true}
	if err := u.repository.DB().Create(value).Error; err != nil {
		return nil, classifyDBError(err)
	}
	return value, nil
}

func (u *usecase) ListScopes(actor Actor, productID, membershipID int) ([]models.ProductMembershipScope, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ACCESS", "READ"); err != nil {
		return nil, err
	}
	var values []models.ProductMembershipScope
	return values, u.repository.DB().Where("product_id = ? AND membership_id = ?", productID, membershipID).Order("scope_id").Find(&values).Error
}

func (u *usecase) UpdateScope(actor Actor, productID, membershipID, id int, req dto.UpdateMembershipScopeRequest) (*models.ProductMembershipScope, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ACCESS", "GRANT"); err != nil {
		return nil, err
	}
	if !validScope(u.repository.DB(), productID, req.ScopeLevel, req.ProjectID, req.CategoryID) {
		return nil, ErrInvalid
	}
	updates := map[string]any{"scope_level": req.ScopeLevel, "project_id": req.ProjectID, "category_id": req.CategoryID}
	res := u.repository.DB().Model(&models.ProductMembershipScope{}).Where("scope_id = ? AND membership_id = ? AND product_id = ?", id, membershipID, productID).Updates(updates)
	if res.Error != nil {
		return nil, classifyDBError(res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	var value models.ProductMembershipScope
	if err := u.repository.DB().First(&value, id).Error; err != nil {
		return nil, classifyDBError(err)
	}
	return &value, nil
}

func (u *usecase) CreatePermissionRule(actor Actor, req dto.CreatePermissionRuleRequest) (*models.UserRolePermissionRule, error) {
	if req.ProductID == nil {
		if !actor.PlatformAdmin {
			return nil, ErrForbidden
		}
	} else if err := u.authorize(actor, AccessTarget{ProductID: *req.ProductID, ProjectID: req.ProjectID, CategoryID: req.CategoryID}, "ACCESS", "GRANT"); err != nil {
		return nil, err
	}
	if !u.exists(&models.User{}, "user_id = ?", req.UserID) {
		return nil, ErrInvalid
	}
	value := &models.UserRolePermissionRule{UserID: req.UserID, ProductID: req.ProductID, RoleID: req.RoleID, ProjectID: req.ProjectID, CategoryID: req.CategoryID, ResourceType: req.ResourceType, Action: req.Action, Effect: req.Effect, ScopeLevel: req.ScopeLevel, GrantedBy: &actor.UserID, IsActive: true, ExpiresAt: req.ExpiresAt}
	if err := u.repository.DB().Create(value).Error; err != nil {
		return nil, classifyDBError(err)
	}
	return value, nil
}

func (u *usecase) ListPermissionRules(actor Actor, productID int) ([]models.UserRolePermissionRule, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ACCESS", "READ"); err != nil {
		return nil, err
	}
	var values []models.UserRolePermissionRule
	return values, u.repository.DB().Where("product_id = ?", productID).Order("permission_rule_id").Find(&values).Error
}

func (u *usecase) UpdatePermissionRule(actor Actor, id int, req dto.UpdatePermissionRuleRequest) (*models.UserRolePermissionRule, error) {
	var value models.UserRolePermissionRule
	if err := u.repository.DB().First(&value, id).Error; err != nil {
		return nil, classifyDBError(err)
	}
	if value.ProductID == nil {
		if !actor.PlatformAdmin {
			return nil, ErrForbidden
		}
	} else if err := u.authorize(actor, AccessTarget{ProductID: *value.ProductID}, "ACCESS", "GRANT"); err != nil {
		return nil, err
	}
	updates := map[string]any{}
	if req.Effect != nil {
		updates["effect"] = *req.Effect
	}
	if req.ExpiresAt != nil {
		updates["expires_at"] = req.ExpiresAt
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if err := u.repository.DB().Model(&value).Updates(updates).Error; err != nil {
		return nil, classifyDBError(err)
	}
	if err := u.repository.DB().First(&value, id).Error; err != nil {
		return nil, err
	}
	return &value, nil
}

func (u *usecase) CheckPermission(actor Actor, req dto.PermissionCheckRequest) (*dto.PermissionCheckResponse, error) {
	if req.ProductID == nil {
		allowed := actor.PlatformAdmin
		return &dto.PermissionCheckResponse{Allowed: allowed, Reason: map[bool]string{true: "platform role override", false: "product_id is required"}[allowed]}, nil
	}
	err := u.authorize(actor, AccessTarget{ProductID: *req.ProductID, ProjectID: req.ProjectID, CategoryID: req.CategoryID}, req.ResourceType, req.Action)
	if err == nil {
		return &dto.PermissionCheckResponse{Allowed: true, Reason: "platform role, product role, or explicit rule allowed access"}, nil
	}
	if errors.Is(err, ErrForbidden) {
		return &dto.PermissionCheckResponse{Allowed: false, Reason: "no matching permission"}, nil
	}
	return nil, err
}

func (u *usecase) authorize(actor Actor, target AccessTarget, resource, action string) error {
	if actor.PlatformAdmin {
		return nil
	}
	if target.ProductID <= 0 || !u.exists(&models.Product{}, "product_id = ?", target.ProductID) {
		return ErrNotFound
	}
	var memberships []models.ProductMembership
	if err := u.repository.DB().Where("user_id = ? AND product_id = ? AND is_active = TRUE AND (expires_at IS NULL OR expires_at > ?)", actor.UserID, target.ProductID, time.Now()).Find(&memberships).Error; err != nil {
		return err
	}
	if len(memberships) == 0 {
		return ErrForbidden
	}
	var rules []models.UserRolePermissionRule
	if err := u.repository.DB().Where("user_id = ? AND resource_type = ? AND action = ? AND is_active = TRUE AND (expires_at IS NULL OR expires_at > ?) AND (product_id IS NULL OR product_id = ?)", actor.UserID, strings.ToUpper(resource), strings.ToUpper(action), time.Now(), target.ProductID).Find(&rules).Error; err != nil {
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
		if err := u.repository.DB().Where("role_id = ? AND product_id = ?", membership.RoleID, target.ProductID).First(&role).Error; err == nil && permissionJSONAllows(role.Permissions, resource, action) {
			return nil
		}
	}
	return ErrForbidden
}

func (u *usecase) scopeAllows(membershipID int, target AccessTarget, allowParentRead bool) bool {
	var scopes []models.ProductMembershipScope
	if err := u.repository.DB().Where("membership_id = ? AND product_id = ? AND is_active = TRUE", membershipID, target.ProductID).Find(&scopes).Error; err != nil {
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
				if err := u.repository.DB().Select("path_ids").First(&feature, *target.CategoryID).Error; err == nil && pathContains(feature.PathIDs, *scope.CategoryID) {
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
	return existsDB(u.repository.DB(), model, query, args...)
}
func existsDB(db *gorm.DB, model any, query string, args ...any) bool {
	var count int64
	return db.Model(model).Where(query, args...).Count(&count).Error == nil && count > 0
}
func (u *usecase) getFeature(productID, projectID, id int) (*models.ProjectFeature, error) {
	var value models.ProjectFeature
	if err := u.repository.DB().Where("category_id = ? AND product_id = ? AND project_id = ?", id, productID, projectID).First(&value).Error; err != nil {
		return nil, classifyDBError(err)
	}
	return &value, nil
}

func rebuildFeaturePaths(db *gorm.DB, projectID int) error {
	var values []models.ProjectFeature
	if err := db.Where("project_id = ?", projectID).Find(&values).Error; err != nil {
		return err
	}
	byID := map[int]*models.ProjectFeature{}
	for i := range values {
		byID[values[i].CategoryID] = &values[i]
	}
	state := map[int]int{}
	var visit func(*models.ProjectFeature) error
	visit = func(value *models.ProjectFeature) error {
		if state[value.CategoryID] == 2 {
			return nil
		}
		if state[value.CategoryID] == 1 {
			return ErrInvalid
		}
		state[value.CategoryID] = 1
		if value.ParentID == nil {
			path, ids := value.CategoryCode, fmt.Sprint(value.CategoryID)
			value.FullPath, value.PathIDs, value.Level = &path, &ids, 1
		} else {
			parent := byID[*value.ParentID]
			if parent == nil {
				return ErrInvalid
			}
			if err := visit(parent); err != nil {
				return err
			}
			path, ids := *parent.FullPath+"/"+value.CategoryCode, *parent.PathIDs+","+fmt.Sprint(value.CategoryID)
			value.FullPath, value.PathIDs, value.Level = &path, &ids, parent.Level+1
		}
		state[value.CategoryID] = 2
		return db.Model(&models.ProjectFeature{}).Where("category_id = ?", value.CategoryID).Updates(map[string]any{"full_path": value.FullPath, "path_ids": value.PathIDs, "level": value.Level}).Error
	}
	for i := range values {
		if err := visit(&values[i]); err != nil {
			return err
		}
	}
	return nil
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
func validJSON(value json.RawMessage) bool { return len(value) > 0 && json.Valid(value) }
func normalizeCode(value string) string    { return strings.ToLower(strings.TrimSpace(value)) }
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
func toAPIKeyResponse(value *models.ProductAPIKey) dto.APIKeyResponse {
	return dto.APIKeyResponse{KeyID: value.KeyID, ProductID: value.ProductID, EnvironmentID: value.EnvironmentID, KeyName: value.KeyName, KeyPrefix: value.KeyPrefix, Permissions: value.Permissions, IsActive: value.IsActive, ExpiresAt: value.ExpiresAt, TimestampResponse: dto.TimestampResponse{CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt}}
}

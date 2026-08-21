package usecase

import (
	"errors"
	"strings"

	"omnilogs-api/dto"
	"omnilogs-api/models"
	"omnilogs-api/responses"
)

func (u *usecase) CreatePermissionRule(actor Actor, req dto.CreatePermissionRuleRequest) (*models.UserRolePermissionRule, error) {
	if req.ProductID == nil {
		if !actor.PlatformAdmin {
			return nil, responses.ErrInvalid
		}
	} else if err := u.authorize(actor, AccessTarget{ProductID: *req.ProductID, ProjectID: req.ProjectID, CategoryID: req.CategoryID}, "ACCESS", "GRANT"); err != nil {
		return nil, err
	}
	if !u.exists(&models.User{}, "user_id = ?", req.UserID) {
		return nil, responses.ErrInvalid
	}
	resourceType := strings.ToUpper(strings.TrimSpace(req.ResourceType))
	action := strings.ToUpper(strings.TrimSpace(req.Action))
	lookup := map[string]any{"user_id": req.UserID, "product_id": req.ProductID, "role_id": req.RoleID, "project_id": req.ProjectID, "category_id": req.CategoryID, "resource_type": resourceType, "action": action}
	var value models.UserRolePermissionRule
	if err := u.repository.DB().Where(lookup).Limit(1).Find(&value).Error; err != nil {
		return nil, classifyDBError(err)
	}
	if value.PermissionRuleID != 0 {
		if err := u.repository.DB().Model(&value).Updates(map[string]any{"effect": req.Effect, "scope_level": req.ScopeLevel, "granted_by": actor.UserID, "is_active": true, "expires_at": req.ExpiresAt}).Error; err != nil {
			return nil, classifyDBError(err)
		}
		if err := u.repository.DB().First(&value, value.PermissionRuleID).Error; err != nil {
			return nil, classifyDBError(err)
		}
		return &value, nil
	}
	value = models.UserRolePermissionRule{UserID: req.UserID, ProductID: req.ProductID, RoleID: req.RoleID, ProjectID: req.ProjectID, CategoryID: req.CategoryID, ResourceType: resourceType, Action: action, Effect: req.Effect, ScopeLevel: req.ScopeLevel, GrantedBy: &actor.UserID, IsActive: true, ExpiresAt: req.ExpiresAt}
	if err := u.repository.DB().Create(&value).Error; err != nil {
		return nil, classifyDBError(err)
	}
	return &value, nil
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
			return nil, responses.ErrForbidden
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
	err := u.authorize(actor, AccessTarget{ProductID: *req.ProductID, EnvironmentID: req.EnvironmentID, ProjectID: req.ProjectID, CategoryID: req.CategoryID}, req.ResourceType, req.Action)
	if err == nil {
		return &dto.PermissionCheckResponse{Allowed: true, Reason: "platform role, product role, or explicit rule allowed access"}, nil
	}
	if errors.Is(err, responses.ErrForbidden) {
		return &dto.PermissionCheckResponse{Allowed: false, Reason: "no matching permission"}, nil
	}
	return nil, err
}

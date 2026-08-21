package usecase

import (
	"fmt"
	"strings"
	"time"

	"omnilogs-api/dto"
	"omnilogs-api/models"
	"omnilogs-api/responses"

	"gorm.io/gorm"
)

func (u *usecase) CreateRole(actor Actor, productID int, req dto.CreateProductRoleRequest) (*models.ProductRole, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ROLE", "CREATE"); err != nil {
		return nil, err
	}
	if req.ProductID != 0 && req.ProductID != productID || strings.TrimSpace(req.RoleName) == "" {
		return nil, responses.ErrInvalid
	}
	if req.TemplateID != nil && !u.exists(&models.RoleTemplate{}, "template_id = ? AND is_active = TRUE", *req.TemplateID) {
		return nil, responses.ErrInvalid
	}
	permissions := req.Permissions
	if len(permissions) == 0 {
		permissions = []dto.RolePermissionAssignment{{ResourceType: "LOG", Action: "READ"}}
	}
	if !validRolePermissions(permissions) {
		return nil, responses.ErrInvalid
	}

	value := &models.ProductRole{
		ProductID:         productID,
		TemplateID:        req.TemplateID,
		RoleCode:          fmt.Sprintf("role-%d-%d", productID, time.Now().UnixNano()),
		RoleName:          strings.TrimSpace(req.RoleName),
		LegacyPermissions: []byte(`[]`),
		IsActive:          true,
	}
	if err := u.repository.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(value).Error; err != nil {
			return classifyDBError(err)
		}
		permissions := toRolePermissions(value.RoleID, permissions)
		if err := tx.Create(&permissions).Error; err != nil {
			return classifyDBError(err)
		}
		value.Permissions = permissions
		return nil
	}); err != nil {
		return nil, err
	}
	return value, nil
}

func (u *usecase) ListRoles(actor Actor, productID int) ([]models.ProductRole, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ROLE", "READ"); err != nil {
		return nil, err
	}
	var values []models.ProductRole
	return values, u.repository.DB().Preload("Permissions").Where("product_id = ?", productID).Order("role_id").Find(&values).Error
}

func (u *usecase) UpdateRole(actor Actor, productID, id int, req dto.UpdateRoleRequest) (*models.ProductRole, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ROLE", "UPDATE"); err != nil {
		return nil, err
	}
	if req.Description != nil {
		return nil, responses.ErrInvalid
	}
	updates := map[string]any{}
	if req.RoleName != nil {
		updates["role_name"] = strings.TrimSpace(*req.RoleName)
	}
	replacePermissions := len(req.Permissions) > 0
	if replacePermissions && !validRolePermissions(req.Permissions) {
		return nil, responses.ErrInvalid
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	var value models.ProductRole
	if err := u.repository.Transaction(func(tx *gorm.DB) error {
		if len(updates) > 0 {
			res := tx.Model(&models.ProductRole{}).Where("role_id = ? AND product_id = ?", id, productID).Updates(updates)
			if res.Error != nil {
				return classifyDBError(res.Error)
			}
			if res.RowsAffected == 0 {
				return responses.ErrNotFound
			}
		} else if err := tx.Where("role_id = ? AND product_id = ?", id, productID).First(&models.ProductRole{}).Error; err != nil {
			return classifyDBError(err)
		}
		if replacePermissions {
			if err := tx.Where("role_id = ?", id).Delete(&models.ProductRolePermission{}).Error; err != nil {
				return classifyDBError(err)
			}
			permissions := toRolePermissions(id, req.Permissions)
			if err := tx.Create(&permissions).Error; err != nil {
				return classifyDBError(err)
			}
		}
		if err := tx.Preload("Permissions").Where("role_id = ? AND product_id = ?", id, productID).First(&value).Error; err != nil {
			return classifyDBError(err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &value, nil
}

func (u *usecase) DeleteRole(actor Actor, productID, id int) (*models.ProductRole, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ROLE", "DELETE"); err != nil {
		return nil, err
	}
	var role models.ProductRole
	err := u.repository.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ? AND product_id = ?", id, productID).First(&role).Error; err != nil {
			return classifyDBError(err)
		}
		var memberships int64
		if err := tx.Model(&models.ProductMembership{}).Where("role_id = ? AND product_id = ?", id, productID).Count(&memberships).Error; err != nil {
			return err
		}
		if memberships > 0 {
			return responses.ErrConflict
		}
		if err := tx.Where("role_id = ?", id).Delete(&models.ProductRolePermission{}).Error; err != nil {
			return classifyDBError(err)
		}
		if err := tx.Delete(&role).Error; err != nil {
			return classifyDBError(err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &role, nil
}

package usecase

import (
	"strings"

	"omnilogs-api/dto"
	"omnilogs-api/models"

	"gorm.io/gorm"
)

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
	if req.Description != nil {
		return nil, ErrInvalid
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
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if len(updates) > 0 {
		res := u.repository.DB().Model(&models.ProductRole{}).Where("role_id = ? AND product_id = ?", id, productID).Updates(updates)
		if res.Error != nil {
			return nil, classifyDBError(res.Error)
		}
		if res.RowsAffected == 0 {
			return nil, ErrNotFound
		}
	}
	var value models.ProductRole
	if err := u.repository.DB().Where("role_id = ? AND product_id = ?", id, productID).First(&value).Error; err != nil {
		return nil, classifyDBError(err)
	}
	return &value, nil
}

func (u *usecase) DeleteRole(actor Actor, productID, id int) error {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ROLE", "DELETE"); err != nil {
		return err
	}
	return u.repository.Transaction(func(tx *gorm.DB) error {
		var role models.ProductRole
		if err := tx.Where("role_id = ? AND product_id = ?", id, productID).First(&role).Error; err != nil {
			return classifyDBError(err)
		}
		var memberships int64
		if err := tx.Model(&models.ProductMembership{}).Where("role_id = ? AND product_id = ?", id, productID).Count(&memberships).Error; err != nil {
			return err
		}
		if memberships > 0 {
			return ErrConflict
		}
		if err := tx.Delete(&role).Error; err != nil {
			return classifyDBError(err)
		}
		return nil
	})
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

func (u *usecase) DeleteMembership(actor Actor, productID, id int) error {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ACCESS", "GRANT"); err != nil {
		return err
	}
	return u.repository.Transaction(func(tx *gorm.DB) error {
		var membership models.ProductMembership
		if err := tx.Where("membership_id = ? AND product_id = ?", id, productID).First(&membership).Error; err != nil {
			return classifyDBError(err)
		}
		if err := tx.Where("membership_id = ? AND product_id = ?", id, productID).Delete(&models.ProductMembershipScope{}).Error; err != nil {
			return classifyDBError(err)
		}
		if err := tx.Delete(&membership).Error; err != nil {
			return classifyDBError(err)
		}
		return nil
	})
}

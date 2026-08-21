package repository

import (
	"errors"

	"omnilogs-api/models"

	"gorm.io/gorm"
)

func (r *repository) populateUserRole(user *models.User) error {
	var membership models.PlatformMembership
	err := r.db.Where("user_id = ? AND is_active = true", user.UserID).First(&membership).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No role assigned, default to admin role
			user.RoleID = 5
			user.Role = &models.UserRole{RoleName: "admin"}
			return nil
		}
		return err
	}

	var role models.PlatformRole
	err = r.db.Where("platform_role_id = ?", membership.PlatformRoleID).First(&role).Error
	if err != nil {
		return err
	}

	user.RoleID = uint(role.PlatformRoleID)
	user.Role = &models.UserRole{RoleName: role.RoleCode}
	return nil
}

func (r *repository) getOrCreateRegistrationRoleID() (int, error) {
	var role models.PlatformRole
	err := r.db.Where("role_code = ?", "admin").First(&role).Error
	if err == nil {
		return role.PlatformRoleID, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		role = models.PlatformRole{
			RoleCode:     "admin",
			RoleName:     "Admin",
			Permissions:  []byte(`{"default_menu_access": true}`),
			IsSystemRole: true,
			IsActive:     true,
		}
		if err := r.db.Create(&role).Error; err != nil {
			return 0, err
		}
		return role.PlatformRoleID, nil
	}
	return 0, err
}

func (r *repository) UpdateRoleID(userID uint, roleID uint) error {
	var membership models.PlatformMembership
	err := r.db.Where("user_id = ?", userID).First(&membership).Error
	if err == nil {
		return r.db.Model(&membership).Update("platform_role_id", roleID).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		membership = models.PlatformMembership{
			UserID:         int(userID),
			PlatformRoleID: int(roleID),
			IsActive:       true,
		}
		return r.db.Create(&membership).Error
	}
	return err
}

func (r *repository) CanOwnerManageUser(ownerUserID uint, targetUserID uint) (bool, error) {
	var count int64
	err := r.db.Table("product_memberships AS owner_memberships").
		Joins("JOIN product_roles ON product_roles.role_id = owner_memberships.role_id").
		Joins("JOIN product_memberships AS target_memberships ON target_memberships.product_id = owner_memberships.product_id").
		Where("owner_memberships.user_id = ? AND owner_memberships.is_active = TRUE", ownerUserID).
		Where("target_memberships.user_id = ? AND target_memberships.is_active = TRUE", targetUserID).
		Where("product_roles.role_code = ?", "product_owner").
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *repository) FindPlatformRoleByID(roleID uint) (*models.PlatformRole, error) {
	var role models.PlatformRole
	if err := r.db.Where("platform_role_id = ?", roleID).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *repository) UpdatePlatformRole(userID uint, roleID uint) error {
	var membership models.PlatformMembership
	err := r.db.Where("user_id = ?", userID).First(&membership).Error
	if err == nil {
		return r.db.Model(&membership).Update("platform_role_id", roleID).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		membership = models.PlatformMembership{
			UserID:         int(userID),
			PlatformRoleID: int(roleID),
			IsActive:       true,
		}
		return r.db.Create(&membership).Error
	}
	return err
}

func (r *repository) UpdateUserRole(userID uint, roleID uint) error {
	var membership models.ProductMembership
	err := r.db.Where("user_id = ?", userID).First(&membership).Error
	if err == nil {
		return r.db.Model(&membership).Update("role_id", roleID).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		membership = models.ProductMembership{
			UserID: int(userID),
			RoleID: int(roleID),
		}
		return r.db.Create(&membership).Error
	}
	return err
}

func (r *repository) UpdateUserPermission(userID uint, productID uint, environmentID uint, permissions map[string]any) error {
	var perm models.UserPermission
	err := r.db.Where("user_id = ? AND product_id = ? AND environment_id = ?", int(userID), int(productID), int(environmentID)).First(&perm).Error
	if err == nil {
		perm.Permissions = permissions
		return r.db.Save(&perm).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		perm = models.UserPermission{
			UserID:        int(userID),
			ProductID:     int(productID),
			EnvironmentID: int(environmentID),
			Permissions:   permissions,
		}
		return r.db.Create(&perm).Error
	}
	return err
}

func (r *repository) CreateUserPermission(userID uint, productID uint, environmentID uint, permissions map[string]any) error {
	perm := models.UserPermission{
		UserID:        int(userID),
		ProductID:     int(productID),
		EnvironmentID: int(environmentID),
		Permissions:   permissions,
	}
	return r.db.Create(&perm).Error
}

func (r *repository) GetUserPermissionByID(userID uint, productID uint, environmentID uint) (*models.UserPermission, error) {
	var perm models.UserPermission
	err := r.db.Where("user_id = ? AND product_id = ? AND environment_id = ?", int(userID), int(productID), int(environmentID)).First(&perm).Error
	return &perm, err
}

func (r *repository) FindPlatformRoleByCode(code string) (*models.PlatformRole, error) {
	var role models.PlatformRole
	if err := r.db.Where("role_code = ?", code).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *repository) FindProductRoleByID(roleID uint) (*models.ProductRole, error) {
	var role models.ProductRole
	if err := r.db.Preload("Permissions").Where("role_id = ?", roleID).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

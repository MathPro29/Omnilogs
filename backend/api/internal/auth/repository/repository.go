package repository

import (
	"errors"
	"omnilogs-api/models"

	"gorm.io/gorm"
)

type Repository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(userID uint) (*models.User, error)
	ListAllUsers() ([]models.User, error)
	UpdateRoleID(userID uint, roleID uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) populateUserRole(user *models.User) error {
	var membership models.PlatformMembership
	err := r.db.Where("user_id = ? AND is_active = true", user.UserID).First(&membership).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No role assigned, default to user role
			user.RoleID = 1
			user.Role = &models.UserRole{RoleName: "user"}
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

func (r *repository) getOrCreateDefaultRoleID() (int, error) {
	var role models.PlatformRole
	err := r.db.Where("role_code = ?", "user").First(&role).Error
	if err == nil {
		return role.PlatformRoleID, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create the default user role
		role = models.PlatformRole{
			RoleCode:     "user",
			RoleName:     "User",
			Permissions:  []byte("{}"),
			IsSystemRole: true,
			IsActive:     true,
		}
		if err := r.db.Create(&role).Error; err != nil {
			return 0, err
		}

		// Also create admin and superadmin roles if they don't exist
		adminRole := models.PlatformRole{
			RoleCode:     "admin",
			RoleName:     "Admin",
			Permissions:  []byte("{}"),
			IsSystemRole: true,
			IsActive:     true,
		}
		r.db.FirstOrCreate(&adminRole, models.PlatformRole{RoleCode: "admin"})

		superadminRole := models.PlatformRole{
			RoleCode:     "superadmin",
			RoleName:     "Superadmin",
			Permissions:  []byte("{}"),
			IsSystemRole: true,
			IsActive:     true,
		}
		r.db.FirstOrCreate(&superadminRole, models.PlatformRole{RoleCode: "superadmin"})

		return role.PlatformRoleID, nil
	}
	return 0, err
}

func (r *repository) Create(user *models.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return err
	}

	defaultRoleID, err := r.getOrCreateDefaultRoleID()
	if err != nil {
		return err
	}

	membership := models.PlatformMembership{
		UserID:         user.UserID,
		PlatformRoleID: defaultRoleID,
		IsActive:       true,
	}
	if err := r.db.Create(&membership).Error; err != nil {
		return err
	}

	return r.populateUserRole(user)
}

func (r *repository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	if err := r.populateUserRole(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindByID(userID uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	if err := r.populateUserRole(&user); err != nil {
		return nil, err
	}
	return &user, nil
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

func (r *repository) ListAllUsers() ([]models.User, error) {
	var users []models.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	for i := range users {
		if err := r.populateUserRole(&users[i]); err != nil {
			return nil, err
		}
	}
	return users, nil
}

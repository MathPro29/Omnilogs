package repository

import (
	"errors"
	"time"

	"omnilogs-api/models"

	"gorm.io/gorm"
)

type Repository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByIdentifier(identifier string) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	FindByID(userID uint) (*models.User, error)
	ListAllUsers() ([]models.User, error)
	UpdateRoleID(userID uint, roleID uint) error
	CanOwnerManageUser(ownerUserID uint, targetUserID uint) (bool, error)
	FindPlatformRoleByID(roleID uint) (*models.PlatformRole, error)
	CreateSession(session *models.AuthSession) error
	FindActiveSessionByRefreshTokenHash(hash string) (*models.AuthSession, error)
	RevokeSessionByRefreshTokenHash(hash string, revokedAt time.Time) error
	TouchSession(sessionID int, lastUsedAt time.Time) error
	RevokeAllUserSessions(userID uint, revokedAt time.Time) error
	CreatePasswordResetToken(token *models.PasswordResetToken) error
	FindActivePasswordResetToken(hash string) (*models.PasswordResetToken, error)
	MarkPasswordResetTokenUsed(id int, usedAt time.Time) error
	UpdatePassword(userID uint, passwordHash string) error
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

func (r *repository) FindByIdentifier(identifier string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ? OR username = ?", identifier, identifier).First(&user).Error; err != nil {
		return nil, err
	}
	if err := r.populateUserRole(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
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

func (r *repository) CreateSession(session *models.AuthSession) error {
	return r.db.Create(session).Error
}

func (r *repository) FindActiveSessionByRefreshTokenHash(hash string) (*models.AuthSession, error) {
	var session models.AuthSession
	if err := r.db.Where("refresh_token_hash = ? AND revoked_at IS NULL", hash).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *repository) RevokeSessionByRefreshTokenHash(hash string, revokedAt time.Time) error {
	return r.db.Model(&models.AuthSession{}).
		Where("refresh_token_hash = ? AND revoked_at IS NULL", hash).
		Updates(map[string]any{"revoked_at": revokedAt, "last_used_at": revokedAt}).Error
}

func (r *repository) TouchSession(sessionID int, lastUsedAt time.Time) error {
	return r.db.Model(&models.AuthSession{}).
		Where("session_id = ?", sessionID).
		Update("last_used_at", lastUsedAt).Error
}

func (r *repository) RevokeAllUserSessions(userID uint, revokedAt time.Time) error {
	return r.db.Model(&models.AuthSession{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Updates(map[string]any{"revoked_at": revokedAt, "last_used_at": revokedAt}).Error
}

func (r *repository) CreatePasswordResetToken(token *models.PasswordResetToken) error {
	return r.db.Create(token).Error
}

func (r *repository) FindActivePasswordResetToken(hash string) (*models.PasswordResetToken, error) {
	var token models.PasswordResetToken
	if err := r.db.Where("token_hash = ? AND used_at IS NULL", hash).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *repository) MarkPasswordResetTokenUsed(id int, usedAt time.Time) error {
	return r.db.Model(&models.PasswordResetToken{}).
		Where("password_reset_token_id = ?", id).
		Update("used_at", usedAt).Error
}

func (r *repository) UpdatePassword(userID uint, passwordHash string) error {
	return r.db.Model(&models.User{}).
		Where("user_id = ?", userID).
		Update("password_hash", passwordHash).Error
}

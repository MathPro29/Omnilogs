package repository

import (
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
	UpdateUserField(userID uint, fieldName string, value interface{}) error
	UpdatePlatformRole(userID uint, roleID uint) error
	UpdateUserRole(userID uint, roleID uint) error
	UpdateUserPermission(userID uint, productID uint, environmentID uint, permissions map[string]interface{}) error
	CreateUserPermission(userID uint, productID uint, environmentID uint, permissions map[string]interface{}) error
	GetUserPermissionByID(userID uint, productID uint, environmentID uint) (*models.UserPermission, error)
	FindPlatformRoleByCode(code string) (*models.PlatformRole, error)
	FindProductRoleByID(roleID uint) (*models.ProductRole, error)
	DeleteUserByID(userID uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

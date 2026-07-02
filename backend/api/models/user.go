package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	UserID       int        `gorm:"primaryKey;autoIncrement" json:"user_id"`
	Username     *string    `json:"username,omitempty"`
	FirstName    *string    `json:"first_name,omitempty"`
	LastName     *string    `json:"last_name,omitempty"`
	Email        string     `gorm:"not null;uniqueIndex" json:"email"`
	PhoneNumber  *string    `json:"phone_number,omitempty"`
	PasswordHash string     `gorm:"not null" json:"-"`
	IsActive     bool       `gorm:"not null;default:true" json:"is_active"`
	LastLoginAt  *time.Time `gorm:"type:timestamptz" json:"last_login_at,omitempty"`
	Timestamps
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	RoleID    uint           `gorm:"-" json:"role_id,omitempty"`
	Role      *UserRole      `gorm:"-" json:"role,omitempty"`
}

type UserRole struct {
	RoleName string `json:"role_name"`
}

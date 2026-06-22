package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	FirstName    string         `json:"firstName" gorm:"not null"`
	LastName     string         `json:"lastName" gorm:"not null"`
	Email        string         `json:"email" gorm:"uniqueIndex;not null"`
	PhoneNumber  *string        `json:"phoneNumber" gorm:"type:varchar(20);index"`
	RoleID       uint           `json:"roleId" gorm:"not null;index;default:1"`
	Role         *Role          `json:"role" gorm:"foreignKey:RoleID;references:ID"`
	Provider     string         `json:"-"`
	ProviderID   string         `json:"-"`
	PasswordHash string         `json:"-" gorm:"not null"`
}

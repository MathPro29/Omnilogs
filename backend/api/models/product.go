package models

import "gorm.io/gorm"

type Product struct {
	ProductID      int            `gorm:"primaryKey;autoIncrement" json:"product_id"`
	ProductName    string         `gorm:"not null" json:"product_name"`
	ProductCode    string         `gorm:"not null" json:"product_code"`
	Description    string         `gorm:"type:text" json:"description,omitempty"`
	SetupStatus    string         `gorm:"type:varchar(50);not null;default:'DRAFT'" json:"setup_status"`
	IsActive       bool           `gorm:"not null;default:true" json:"is_active"`
	MembershipID   int            `gorm:"-" json:"membership_id,omitempty"`
	MembershipRole string         `gorm:"-" json:"membership_role,omitempty"`
	IsOwner        bool           `gorm:"-" json:"is_owner"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	Timestamps
	ProductEnvironments []ProductEnvironment `gorm:"foreignKey:ProductID;references:ProductID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"environments"`
}

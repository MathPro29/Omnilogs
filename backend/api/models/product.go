package models

import "gorm.io/gorm"

type Product struct {
	ProductID   int            `gorm:"primaryKey;autoIncrement" json:"product_id"`
	ProductName string         `gorm:"not null" json:"product_name"`
	ProductCode string         `gorm:"not null;unique" json:"product_code"`
	IsActive    bool           `gorm:"not null;default:true" json:"is_active"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Timestamps
	ProductEnvironments []ProductEnvironment `gorm:"foreignKey:ProductID;references:ProductID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"environments"`
}

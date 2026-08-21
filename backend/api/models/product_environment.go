package models

import "time"

type ProductEnvironment struct {
	EnvironmentID   int        `gorm:"primaryKey;autoIncrement" json:"environment_id"`
	ProductID       int        `gorm:"not null;uniqueIndex:uq_product_environment,priority:1" json:"product_id"`
	EnvironmentCode string     `gorm:"not null;uniqueIndex:uq_product_environment,priority:2" json:"environment_code"`
	EnvironmentName string     `gorm:"not null" json:"environment_name"`
	EnvironmentType string     `gorm:"type:varchar(50);not null;default:'DEVELOPMENT'" json:"environment_type"`
	IsDefault       bool       `gorm:"not null;default:false" json:"is_default"`
	IsActive        bool       `gorm:"not null;default:true" json:"is_active"`
	CreatedAt       *time.Time `gorm:"type:timestamptz" json:"created_at,omitempty"`
	DeletedAt       *time.Time `gorm:"type:timestamptz" json:"deleted_at,omitempty"`
	UpdatedAt       *time.Time `gorm:"type:timestamptz" json:"updated_at,omitempty"`
}

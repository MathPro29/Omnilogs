package models

import "time"

type SensitiveLogAccessHistory struct {
	AccessID     string     `gorm:"type:uuid;primaryKey" json:"access_id"`
	RequestID    *string    `gorm:"type:uuid" json:"request_id,omitempty"`
	ProductID    *int       `json:"product_id,omitempty"`
	UserID       int        `gorm:"not null" json:"user_id"`
	ProjectID    *int       `json:"project_id,omitempty"`
	CategoryID   *int       `json:"category_id,omitempty"`
	LogID        *string    `gorm:"type:uuid;index" json:"log_id,omitempty"`
	FieldPath    *string    `json:"field_path,omitempty"`
	AccessReason *string    `gorm:"type:text" json:"access_reason,omitempty"`
	AuthMethod   *string    `json:"auth_method,omitempty"`
	AccessStatus *string    `json:"access_status,omitempty"`
	CreatedAt    *time.Time `gorm:"type:timestamptz" json:"created_at,omitempty"`
}
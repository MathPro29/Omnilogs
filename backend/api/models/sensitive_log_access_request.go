package models

import "time"

type SensitiveLogAccessRequest struct {
	RequestID         string     `gorm:"type:uuid;primaryKey" json:"request_id"`
	UserID            int        `gorm:"not null" json:"user_id"`
	ProductID         int        `gorm:"not null" json:"product_id"`
	LogID             *string    `gorm:"type:uuid;index" json:"log_id,omitempty"`
	FieldDefinitionID *int       `gorm:"index" json:"field_definition_id,omitempty"`
	SecretID          *string    `gorm:"type:uuid;index" json:"secret_id,omitempty"`
	FieldPath         *string    `json:"field_path,omitempty"`
	Reason            *string    `gorm:"type:text" json:"reason,omitempty"`
	ApprovalStatus    string     `gorm:"not null" json:"approval_status"`
	ApprovedBy        *int       `json:"approved_by,omitempty"`
	ApprovedAt        *time.Time `gorm:"type:timestamptz" json:"approved_at,omitempty"`
	ExpiresAt         *time.Time `gorm:"type:timestamptz" json:"expires_at,omitempty"`
	Timestamps
}
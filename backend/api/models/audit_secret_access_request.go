package models

import "time"

type AuditSecretAccessRequest struct {
	RequestID      string     `gorm:"type:uuid;primaryKey" json:"request_id"`
	AuditID        string     `gorm:"type:uuid;not null;index" json:"audit_id"`
	SecretID       string     `gorm:"type:uuid;not null;index" json:"secret_id"`
	UserID         int        `gorm:"not null;index" json:"user_id"`
	Reason         *string    `gorm:"type:text" json:"reason,omitempty"`
	ApprovalStatus string     `gorm:"not null;size:20;index" json:"approval_status"`
	ApprovedBy     *int       `json:"approved_by,omitempty"`
	ApprovedAt     *time.Time `gorm:"type:timestamptz" json:"approved_at,omitempty"`
	ExpiresAt      *time.Time `gorm:"type:timestamptz" json:"expires_at,omitempty"`
	CreatedAt      *time.Time `gorm:"type:timestamptz" json:"created_at,omitempty"`
	UpdatedAt      *time.Time `gorm:"type:timestamptz" json:"updated_at,omitempty"`
}

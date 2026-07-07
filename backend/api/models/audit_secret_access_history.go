package models

import "time"

type AuditSecretAccessHistory struct {
	AccessID     string     `gorm:"type:uuid;primaryKey" json:"access_id"`
	RequestID    *string    `gorm:"type:uuid;index" json:"request_id,omitempty"`
	AuditID      string     `gorm:"type:uuid;not null;index" json:"audit_id"`
	SecretID     string     `gorm:"type:uuid;not null;index" json:"secret_id"`
	UserID       int        `gorm:"not null;index" json:"user_id"`
	AccessReason *string    `gorm:"type:text" json:"access_reason,omitempty"`
	AuthMethod   *string    `gorm:"size:100" json:"auth_method,omitempty"`
	AccessStatus *string    `gorm:"size:50;index" json:"access_status,omitempty"`
	CreatedAt    *time.Time `gorm:"type:timestamptz" json:"created_at,omitempty"`
}

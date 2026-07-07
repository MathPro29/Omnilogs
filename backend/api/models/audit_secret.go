package models

import "time"

type AuditSecret struct {
	SecretID            string     `gorm:"type:uuid;primaryKey" json:"secret_id"`
	AuditID             string     `gorm:"type:uuid;not null;index" json:"audit_id"`
	ActorUserID         *int64     `gorm:"index" json:"actor_user_id,omitempty"`
	FieldKey            string     `gorm:"not null" json:"field_key"`
	FieldPath           string     `gorm:"not null" json:"field_path"`
	SourceSection       string     `gorm:"not null;index" json:"source_section"`
	EncryptedValue      string     `gorm:"type:text;not null" json:"-"`
	EncryptionAlgorithm *string    `json:"encryption_algorithm,omitempty"`
	RequiresApproval    bool       `gorm:"not null;default:true" json:"requires_approval"`
	CreatedAt           *time.Time `gorm:"type:timestamptz" json:"created_at,omitempty"`
}

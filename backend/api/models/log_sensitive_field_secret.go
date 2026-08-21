package models

import "time"

type LogSensitiveFieldSecret struct {
	SecretID            string     `gorm:"type:uuid;primaryKey" json:"secret_id"`
	LogID               string     `gorm:"type:uuid;not null;index" json:"log_id"`
	ProductID           int        `gorm:"not null" json:"product_id"`
	ProjectID           *int       `json:"project_id,omitempty"`
	CategoryID          *int       `json:"category_id,omitempty"`
	EnvironmentID       *int       `json:"environment_id,omitempty"`
	FieldDefinitionID   *int       `gorm:"index" json:"field_definition_id,omitempty"`
	FieldKey            string     `gorm:"not null" json:"field_key"`
	FieldPath           string     `gorm:"not null" json:"field_path"`
	SourceSection       string     `gorm:"not null;index" json:"source_section"`
	ValueHash           *string    `gorm:"index" json:"value_hash,omitempty"`
	EncryptedValue      string     `gorm:"type:text;not null" json:"-"`
	EncryptionKeyRef    *string    `json:"encryption_key_ref,omitempty"`
	EncryptionAlgorithm *string    `json:"encryption_algorithm,omitempty"`
	RequiresApproval    bool       `gorm:"not null;default:true" json:"requires_approval"`
	RetentionUntil      *time.Time `gorm:"type:timestamptz;index" json:"retention_until,omitempty"`
	PurgedAt            *time.Time `gorm:"type:timestamptz;index" json:"purged_at,omitempty"`
	CreatedAt           *time.Time `gorm:"type:timestamptz" json:"created_at,omitempty"`
}

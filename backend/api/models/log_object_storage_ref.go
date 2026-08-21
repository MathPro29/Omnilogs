package models

import "time"

type LogObjectStorageRef struct {
	ObjectRefID         string     `gorm:"type:uuid;primaryKey" json:"object_ref_id"`
	LogID               string     `gorm:"type:uuid;not null;uniqueIndex:uq_log_object_type,priority:1" json:"log_id"`
	ProductID           int        `gorm:"not null;index:idx_object_product_environment,priority:1" json:"product_id"`
	EnvironmentID       *int       `gorm:"index:idx_object_product_environment,priority:2" json:"environment_id,omitempty"`
	StorageProvider     string     `gorm:"not null;default:GCS" json:"storage_provider"`
	BucketName          *string    `json:"bucket_name,omitempty"`
	ObjectPath          string     `gorm:"type:text;not null" json:"object_path"`
	ObjectType          string     `gorm:"not null;uniqueIndex:uq_log_object_type,priority:2" json:"object_type"`
	FileFormat          *string    `json:"file_format,omitempty"`
	SizeBytes           *int64     `json:"size_bytes,omitempty"`
	Checksum            *string    `json:"checksum,omitempty"`
	EncryptedPayload    *string    `gorm:"type:text" json:"-"`
	EncryptionKeyRef    *string    `json:"encryption_key_ref,omitempty"`
	EncryptionAlgorithm *string    `json:"encryption_algorithm,omitempty"`
	IsEncrypted         bool       `gorm:"not null;default:false" json:"is_encrypted"`
	RetentionUntil      *time.Time `gorm:"type:timestamptz;index" json:"retention_until,omitempty"`
	PurgedAt            *time.Time `gorm:"type:timestamptz;index" json:"purged_at,omitempty"`
	CreatedAt           *time.Time `gorm:"type:timestamptz" json:"created_at,omitempty"`
}

package models

import "time"

type LogArchive struct {
	ArchiveID           string     `gorm:"type:uuid;primaryKey" json:"archive_id"`
	ProductID           int        `gorm:"not null" json:"product_id"`
	EnvironmentID       *int       `json:"environment_id,omitempty"`
	ProjectID           *int       `json:"project_id,omitempty"`
	ArchiveYear         int        `gorm:"not null" json:"archive_year"`
	ArchiveMonth        int        `gorm:"not null" json:"archive_month"`
	DateFrom            *time.Time `gorm:"type:timestamptz" json:"date_from,omitempty"`
	DateTo              *time.Time `gorm:"type:timestamptz" json:"date_to,omitempty"`
	StorageProvider     string     `gorm:"not null;default:GCS" json:"storage_provider"`
	BucketName          *string    `json:"bucket_name,omitempty"`
	FilePath            *string    `gorm:"type:text" json:"file_path,omitempty"`
	FileFormat          *string    `json:"file_format,omitempty"`
	TotalLogs           *int       `json:"total_logs,omitempty"`
	CompressedSizeBytes *int64     `json:"compressed_size_bytes,omitempty"`
	Checksum            *string    `json:"checksum,omitempty"`
	Status              *string    `gorm:"index" json:"status,omitempty"`
	ExportedAt          *time.Time `gorm:"type:timestamptz" json:"exported_at,omitempty"`
	RestoredAt          *time.Time `gorm:"type:timestamptz" json:"restored_at,omitempty"`
	PurgedAt            *time.Time `gorm:"type:timestamptz" json:"purged_at,omitempty"`
	Timestamps
}
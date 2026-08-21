package models

import (
	"encoding/json"
	"time"
)

type LogArchive struct {
	ArchiveID           string          `gorm:"type:uuid;primaryKey" json:"archive_id"`
	ProductID           int             `gorm:"not null" json:"product_id"`
	EnvironmentID       *int            `json:"environment_id,omitempty"`
	PolicyID            *int            `gorm:"index" json:"policy_id,omitempty"`
	BackupType          string          `gorm:"type:varchar(20);not null;default:'POLICY';index" json:"backup_type"`
	BackupTag           string          `gorm:"type:varchar(180);not null;default:'';index" json:"backup_tag"`
	CoverageKey         *string         `gorm:"type:varchar(255);uniqueIndex" json:"coverage_key,omitempty"`
	ProjectID           *int            `json:"project_id,omitempty"`
	IndexName           *string         `gorm:"type:text" json:"index_name,omitempty"`
	CategoryID          *int            `json:"category_id,omitempty"`
	FeatureID           *int            `json:"feature_id,omitempty"`
	SubFeatureID        *int            `json:"sub_feature_id,omitempty"`
	ArchiveYear         int             `gorm:"not null" json:"archive_year"`
	ArchiveMonth        int             `gorm:"not null" json:"archive_month"`
	DateFrom            *time.Time      `gorm:"type:timestamptz" json:"date_from,omitempty"`
	DateTo              *time.Time      `gorm:"type:timestamptz" json:"date_to,omitempty"`
	StorageProvider     string          `gorm:"not null;default:GCS" json:"storage_provider"`
	BucketName          *string         `json:"bucket_name,omitempty"`
	ObjectKey           *string         `gorm:"type:text" json:"object_key,omitempty"`
	FilePath            *string         `gorm:"type:text" json:"file_path,omitempty"`
	SnapshotRepository  *string         `gorm:"type:text" json:"snapshot_repository,omitempty"`
	SnapshotName        *string         `gorm:"type:text" json:"snapshot_name,omitempty"`
	SnapshotUUID        *string         `gorm:"type:text" json:"snapshot_uuid,omitempty"`
	SnapshotStatus      *string         `gorm:"type:varchar(30)" json:"snapshot_status,omitempty"`
	SnapshotIndices     json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"snapshot_indices,omitempty"`
	SnapshotVerifiedAt  *time.Time      `gorm:"type:timestamptz" json:"snapshot_verified_at,omitempty"`
	RestoreIndexPattern *string         `gorm:"type:text" json:"restore_index_pattern,omitempty"`
	RestoreStatus       *string         `gorm:"type:varchar(30)" json:"restore_status,omitempty"`
	FileFormat          *string         `json:"file_format,omitempty"`
	TotalLogs           *int            `json:"total_logs,omitempty"`
	DocumentCount       int64           `gorm:"not null;default:0" json:"document_count"`
	OriginalSizeBytes   int64           `gorm:"not null;default:0" json:"original_size_bytes"`
	CompressedSizeBytes *int64          `json:"compressed_size_bytes,omitempty"`
	Checksum            *string         `json:"checksum,omitempty"`
	Format              string          `gorm:"type:varchar(20);not null;default:'NDJSON'" json:"format"`
	Compression         string          `gorm:"type:varchar(20);not null;default:'GZIP'" json:"compression"`
	Status              *string         `gorm:"index" json:"status,omitempty"`
	ExportedAt          *time.Time      `gorm:"type:timestamptz" json:"exported_at,omitempty"`
	RestoredAt          *time.Time      `gorm:"type:timestamptz" json:"restored_at,omitempty"`
	DeletedAt           *time.Time      `gorm:"type:timestamptz" json:"deleted_at,omitempty"`
	VerifiedAt          *time.Time      `gorm:"type:timestamptz" json:"verified_at,omitempty"`
	DeletedFromActiveAt *time.Time      `gorm:"type:timestamptz" json:"deleted_from_active_at,omitempty"`
	ManifestValid       bool            `gorm:"not null;default:false" json:"manifest_valid"`
	PurgedAt            *time.Time      `gorm:"type:timestamptz" json:"purged_at,omitempty"`
	Timestamps
}

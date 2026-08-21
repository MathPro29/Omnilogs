package dto

import (
	"time"

	"omnilogs-api/models"
)

type CreateArchiveRequest struct {
	ProductID     int        `json:"product_id"`
	EnvironmentID int        `json:"environment_id" binding:"required,gt=0"`
	PolicyID      *int       `json:"policy_id,omitempty"`
	DateFrom      *time.Time `json:"date_from,omitempty"`
	DateTo        *time.Time `json:"date_to,omitempty"`
	CategoryID    *int       `json:"category_id,omitempty"`
	FeatureID     *int       `json:"feature_id,omitempty"`
	SubFeatureID  *int       `json:"sub_feature_id,omitempty"`
	BackupType    string     `json:"backup_type,omitempty"`
	BackupTag     string     `json:"backup_tag,omitempty"`
}

type ListArchiveRequest struct {
	ProductID     int
	EnvironmentID int
	DateFrom      *time.Time
	DateTo        *time.Time
	CategoryID    *int
	FeatureID     *int
	SubFeatureID  *int
	BackupType    string
	BackupTag     string
}

type VerifyArchiveRequest struct {
	// PolicyID makes a user-initiated backfill deterministic when an
	// environment has more than one active retention policy.
	PolicyID *int `json:"policy_id,omitempty"`
}

type VerifyArchiveResponse struct {
	ArchiveID     string `json:"archive_id"`
	Status        string `json:"status"`
	ChecksumValid bool   `json:"checksum_valid"`
	ManifestValid bool   `json:"manifest_valid"`
	SnapshotValid bool   `json:"snapshot_valid"`
	DocumentCount int64  `json:"document_count"`
	DeletedActive bool   `json:"deleted_active_logs"`
}

type RestoreArchiveRequest struct {
	Mode     string `json:"mode"`
	Conflict string `json:"conflict_strategy,omitempty"`
}

type RestoreArchiveResponse struct {
	ArchiveID string `json:"archive_id"`
	Mode      string `json:"mode"`
	IndexName string `json:"index_name"`
	Snapshot  bool   `json:"snapshot_restore"`
	Restored  int64  `json:"restored_count"`
	Skipped   int64  `json:"skipped_count"`
	Conflict  string `json:"conflict_strategy,omitempty"`
}

type ImportArchiveResponse struct {
	ImportedCount int                 `json:"imported_count"`
	Archives      []models.LogArchive `json:"archives"`
}

type ArchiveGroupResponse struct {
	Period          string    `json:"period"`
	DateFrom        time.Time `json:"date_from"`
	DateTo          time.Time `json:"date_to"`
	CategoryID      *int      `json:"category_id,omitempty"`
	FeatureID       *int      `json:"feature_id,omitempty"`
	SubFeatureID    *int      `json:"sub_feature_id,omitempty"`
	DocumentCount   int64     `json:"document_count"`
	OriginalBytes   int64     `json:"original_size_bytes"`
	CompressedBytes int64     `json:"compressed_size_bytes"`
	Status          string    `json:"status"`
	ArchiveCount    int       `json:"archive_count"`
}

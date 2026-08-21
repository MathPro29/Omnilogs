package models

import (
	"encoding/json"
	"time"
)

// LogArchiveJob is durable metadata for archive work. The worker/API may
// process a job synchronously, but clients can always poll this record.
type LogArchiveJob struct {
	JobID         string          `gorm:"type:uuid;primaryKey" json:"job_id"`
	ProductID     int             `gorm:"not null;index" json:"product_id"`
	EnvironmentID int             `gorm:"not null;index" json:"environment_id"`
	Status        string          `gorm:"type:varchar(30);not null;index" json:"status"`
	ArchiveIDs    json.RawMessage `gorm:"type:jsonb;not null;default:'[]'" json:"archive_ids"`
	ErrorMessage  *string         `gorm:"type:text" json:"error_message,omitempty"`
	StartedAt     *time.Time      `gorm:"type:timestamptz" json:"started_at,omitempty"`
	CompletedAt   *time.Time      `gorm:"type:timestamptz" json:"completed_at,omitempty"`
	Timestamps
}

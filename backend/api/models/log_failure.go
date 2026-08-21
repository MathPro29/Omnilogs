package models

import (
	"encoding/json"
	"time"
)

type LogFailure struct {
	FailureID     string          `gorm:"type:uuid;primaryKey" json:"failure_id"`
	LogID         string          `gorm:"not null;index" json:"log_id"`
	AttemptNo     int             `gorm:"not null;default:1;uniqueIndex:uq_failure_attempt,priority:2" json:"attempt_no"`
	QueueItemID   int64           `gorm:"not null;uniqueIndex:uq_failure_attempt,priority:1" json:"queue_item_id"`
	BatchID       string          `gorm:"type:uuid;not null;index" json:"batch_id"`
	SequenceNo    int             `gorm:"not null;default:0" json:"sequence_no"`
	ProductID     *int            `json:"product_id,omitempty"`
	SourceID      *int            `json:"source_id,omitempty"`
	EnvironmentID *int            `json:"environment_id,omitempty"`
	FailureStage  string          `gorm:"not null" json:"failure_stage"`
	FailureType   string          `gorm:"not null" json:"failure_type"`
	Reason        *string         `gorm:"type:text" json:"reason,omitempty"`
	ErrorDetails  json.RawMessage `gorm:"type:jsonb" json:"error_details,omitempty"`
	RetryCount    int             `gorm:"not null;default:0" json:"retry_count"`
	MaxRetryCount int             `gorm:"not null;default:3" json:"max_retry_count"`
	NextRetryAt   *time.Time      `gorm:"type:timestamptz" json:"next_retry_at,omitempty"`
	LastRetryAt   *time.Time      `gorm:"type:timestamptz" json:"last_retry_at,omitempty"`
	Status        string          `gorm:"not null;default:FAILED" json:"status"`
	CreatedAt     *time.Time      `gorm:"type:timestamptz" json:"created_at,omitempty"`
	ResolvedAt    *time.Time      `gorm:"type:timestamptz" json:"resolved_at,omitempty"`
}

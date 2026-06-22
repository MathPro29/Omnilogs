package models

import (
	"encoding/json"
	"time"
)

type LogQueueItem struct {
	QueueItemID         int64           `gorm:"primaryKey;autoIncrement" json:"queue_item_id"`
	BatchID             string          `gorm:"type:uuid;not null;uniqueIndex:uq_batch_sequence,priority:1" json:"batch_id"`
	SequenceNo          int             `gorm:"not null;uniqueIndex:uq_batch_sequence,priority:2" json:"sequence_no"`
	SourceType          string          `gorm:"not null;default:UNKNOWN" json:"source_type"`
	SourcePlatform      string          `gorm:"not null;default:UNKNOWN" json:"source_platform"`
	InputPayload        json.RawMessage `gorm:"type:jsonb" json:"input_payload,omitempty"`
	PayloadSizeBytes    *int64          `json:"payload_size_bytes,omitempty"`
	PayloadPurged       bool            `gorm:"not null;default:false" json:"payload_purged"`
	DetectedProductCode *string         `gorm:"index" json:"detected_product_code,omitempty"`
	WorkerID            *string         `gorm:"index" json:"worker_id,omitempty"`
	ProcessingAttempts  int             `gorm:"not null;default:0" json:"processing_attempts"`
	Status              string          `gorm:"not null;default:PENDING" json:"status"`
	RetryCount          int             `gorm:"not null;default:0" json:"retry_count"`
	MaxRetryCount       int             `gorm:"not null;default:3" json:"max_retry_count"`
	NextRetryAt         *time.Time      `gorm:"type:timestamptz" json:"next_retry_at,omitempty"`
	LastRetryAt         *time.Time      `gorm:"type:timestamptz" json:"last_retry_at,omitempty"`
	ProcessingStartedAt *time.Time      `gorm:"type:timestamptz" json:"processing_started_at,omitempty"`
	ProcessedAt         *time.Time      `gorm:"type:timestamptz" json:"processed_at,omitempty"`
	ErrorMessage        *string         `gorm:"type:text" json:"error_message,omitempty"`
	RetentionUntil      *time.Time      `gorm:"type:timestamptz" json:"retention_until,omitempty"`
	PurgedAt            *time.Time      `gorm:"type:timestamptz" json:"purged_at,omitempty"`
	Timestamps
}

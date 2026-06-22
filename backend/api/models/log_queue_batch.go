package models

import "time"

type LogQueueBatch struct {
	BatchID                 string     `gorm:"type:uuid;primaryKey" json:"batch_id"`
	ProductID               *int       `json:"product_id,omitempty"`
	SourceID                *int       `json:"source_id,omitempty"`
	EnvironmentID           *int       `json:"environment_id,omitempty"`
	QueueKey                string     `gorm:"not null" json:"queue_key"`
	SourceType              string     `gorm:"not null;default:UNKNOWN" json:"source_type"`
	SourcePlatform          string     `gorm:"not null;default:UNKNOWN" json:"source_platform"`
	SourceIP                *string    `json:"source_ip,omitempty"`
	UserAgent               *string    `gorm:"type:text" json:"user_agent,omitempty"`
	IdempotencyKey          *string    `json:"idempotency_key,omitempty"`
	DetectedProductCode     *string    `gorm:"index" json:"detected_product_code,omitempty"`
	ProductResolutionStatus string     `gorm:"not null;default:RESOLVED" json:"product_resolution_status"`
	ProductResolutionReason *string    `gorm:"type:text" json:"product_resolution_reason,omitempty"`
	ProductResolvedAt       *time.Time `gorm:"type:timestamptz" json:"product_resolved_at,omitempty"`
	AvailableAt             *time.Time `gorm:"type:timestamptz" json:"available_at,omitempty"`
	WorkerID                *string    `gorm:"index" json:"worker_id,omitempty"`
	ProcessingAttempts      int        `gorm:"not null;default:0" json:"processing_attempts"`
	TotalLogs               int        `gorm:"not null;default:0" json:"total_logs"`
	Status                  string     `gorm:"not null;default:QUEUED" json:"status"`
	Priority                int        `gorm:"not null;default:0" json:"priority"`
	ReceivedAt              *time.Time `gorm:"type:timestamptz" json:"received_at,omitempty"`
	ProcessingStartedAt     *time.Time `gorm:"type:timestamptz" json:"processing_started_at,omitempty"`
	ProcessedAt             *time.Time `gorm:"type:timestamptz" json:"processed_at,omitempty"`
	RetentionUntil          *time.Time `gorm:"type:timestamptz" json:"retention_until,omitempty"`
	PurgedAt                *time.Time `gorm:"type:timestamptz" json:"purged_at,omitempty"`
	LockedBy                *string    `json:"locked_by,omitempty"`
	LockedAt                *time.Time `gorm:"type:timestamptz" json:"locked_at,omitempty"`
	ErrorMessage            *string    `gorm:"type:text" json:"error_message,omitempty"`
	Timestamps
}

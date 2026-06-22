package models

type LogIngestionPolicy struct {
	PolicyID                int     `gorm:"primaryKey;autoIncrement" json:"policy_id"`
	ProductID               int     `gorm:"not null;uniqueIndex:uq_ingestion_policy,priority:1" json:"product_id"`
	EnvironmentID           int     `gorm:"not null;uniqueIndex:uq_ingestion_policy,priority:2" json:"environment_id"`
	MaxLogsPerMinute        *int    `json:"max_logs_per_minute,omitempty"`
	MaxBatchSize            *int    `json:"max_batch_size,omitempty"`
	FlushIntervalSeconds    *int    `json:"flush_interval_seconds,omitempty"`
	MaxRetries              *int    `json:"max_retries,omitempty"`
	RetryIntervalSeconds    *int    `json:"retry_interval_seconds,omitempty"`
	QueueEnabled            bool    `gorm:"not null;default:true" json:"queue_enabled"`
	QueueWorkerEnabled      bool    `gorm:"not null;default:true" json:"queue_worker_enabled"`
	QueueLockTimeoutSeconds int     `gorm:"not null;default:600" json:"queue_lock_timeout_seconds"`
	StrictOrderingEnabled   bool    `gorm:"not null;default:true" json:"strict_ordering_enabled"`
	RetentionDays           *int    `json:"retention_days,omitempty"`
	ArchiveEnabled          bool    `gorm:"not null;default:false" json:"archive_enabled"`
	ArchiveFormat           *string `json:"archive_format,omitempty"`
	ArchiveStoragePath      *string `gorm:"type:text" json:"archive_storage_path,omitempty"`
	Timestamps
}

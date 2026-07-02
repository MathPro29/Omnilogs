package dto

type UpsertLogIngestionPolicyRequest struct {
	ProductID               int     `json:"product_id" binding:"required,gt=0"`
	EnvironmentID           int     `json:"environment_id" binding:"required,gt=0"`
	MaxLogsPerMinute        *int    `json:"max_logs_per_minute,omitempty" binding:"omitempty,gt=0"`
	MaxBatchSize            *int    `json:"max_batch_size,omitempty" binding:"omitempty,gt=0"`
	FlushIntervalSeconds    *int    `json:"flush_interval_seconds,omitempty" binding:"omitempty,gt=0"`
	MaxRetries              *int    `json:"max_retries,omitempty" binding:"omitempty,gte=0"`
	RetryIntervalSeconds    *int    `json:"retry_interval_seconds,omitempty" binding:"omitempty,gte=0"`
	QueueEnabled            *bool   `json:"queue_enabled,omitempty"`
	QueueWorkerEnabled      *bool   `json:"queue_worker_enabled,omitempty"`
	QueueLockTimeoutSeconds *int    `json:"queue_lock_timeout_seconds,omitempty" binding:"omitempty,gt=0"`
	StrictOrderingEnabled   *bool   `json:"strict_ordering_enabled,omitempty"`
	RetentionDays           *int    `json:"retention_days,omitempty" binding:"omitempty,gte=0"`
	ArchiveEnabled          *bool   `json:"archive_enabled,omitempty"`
	ArchiveFormat           *string `json:"archive_format,omitempty"`
	ArchiveStoragePath      *string `json:"archive_storage_path,omitempty"`
}

type LogIngestionPolicyResponse struct {
	PolicyID                int     `json:"policy_id"`
	ProductID               int     `json:"product_id"`
	EnvironmentID           int     `json:"environment_id"`
	MaxLogsPerMinute        *int    `json:"max_logs_per_minute,omitempty"`
	MaxBatchSize            *int    `json:"max_batch_size,omitempty"`
	FlushIntervalSeconds    *int    `json:"flush_interval_seconds,omitempty"`
	MaxRetries              *int    `json:"max_retries,omitempty"`
	RetryIntervalSeconds    *int    `json:"retry_interval_seconds,omitempty"`
	QueueEnabled            bool    `json:"queue_enabled"`
	QueueWorkerEnabled      bool    `json:"queue_worker_enabled"`
	QueueLockTimeoutSeconds int     `json:"queue_lock_timeout_seconds"`
	StrictOrderingEnabled   bool    `json:"strict_ordering_enabled"`
	RetentionDays           *int    `json:"retention_days,omitempty"`
	ArchiveEnabled          bool    `json:"archive_enabled"`
	ArchiveFormat           *string `json:"archive_format,omitempty"`
	ArchiveStoragePath      *string `json:"archive_storage_path,omitempty"`
	TimestampResponse
}

type CreateElasticIndexPolicyRequest struct {
	ProductID        int     `json:"product_id" binding:"required,gt=0"`
	EnvironmentID    *int    `json:"environment_id,omitempty" binding:"omitempty,gt=0"`
	IndexPrefix      string  `json:"index_prefix" binding:"required"`
	IndexPattern     *string `json:"index_pattern,omitempty"`
	WriteAlias       *string `json:"write_alias,omitempty"`
	RolloverType     string  `json:"rollover_type" binding:"required"`
	NumberOfShards   *int    `json:"number_of_shards,omitempty" binding:"omitempty,gt=0"`
	NumberOfReplicas *int    `json:"number_of_replicas,omitempty" binding:"omitempty,gte=0"`
	RetentionDays    *int    `json:"retention_days,omitempty" binding:"omitempty,gte=0"`
	SchemaVersion    *string `json:"schema_version,omitempty"`
}

type UpdateElasticIndexPolicyRequest struct {
	IndexPattern     *string `json:"index_pattern,omitempty"`
	WriteAlias       *string `json:"write_alias,omitempty"`
	RolloverType     *string `json:"rollover_type,omitempty"`
	NumberOfShards   *int    `json:"number_of_shards,omitempty" binding:"omitempty,gt=0"`
	NumberOfReplicas *int    `json:"number_of_replicas,omitempty" binding:"omitempty,gte=0"`
	RetentionDays    *int    `json:"retention_days,omitempty" binding:"omitempty,gte=0"`
	SchemaVersion    *string `json:"schema_version,omitempty"`
	IsActive         *bool   `json:"is_active,omitempty"`
}

type ElasticIndexPolicyResponse struct {
	ElasticPolicyID  int     `json:"elastic_policy_id"`
	ProductID        int     `json:"product_id"`
	EnvironmentID    *int    `json:"environment_id,omitempty"`
	IndexPrefix      string  `json:"index_prefix"`
	IndexPattern     *string `json:"index_pattern,omitempty"`
	WriteAlias       *string `json:"write_alias,omitempty"`
	RolloverType     string  `json:"rollover_type"`
	NumberOfShards   int     `json:"number_of_shards"`
	NumberOfReplicas int     `json:"number_of_replicas"`
	RetentionDays    *int    `json:"retention_days,omitempty"`
	SchemaVersion    string  `json:"schema_version"`
	IsActive         bool    `json:"is_active"`
	TimestampResponse
}

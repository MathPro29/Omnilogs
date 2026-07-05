package dto

import (
	"encoding/json"
	"time"
)

type IngestLogItemRequest struct {
	SequenceNo          int             `json:"sequence_no" binding:"required,gte=0"`
	SourceType          string          `json:"source_type" binding:"required"`
	SourcePlatform      string          `json:"source_platform" binding:"required"`
	InputPayload        json.RawMessage `json:"input_payload" binding:"required"`
	DetectedProductCode *string         `json:"detected_product_code,omitempty"`
}

type IngestLogBatchRequest struct {
	ProductID           *int                   `json:"product_id,omitempty" binding:"omitempty,gt=0"`
	SourceID            *int                   `json:"source_id,omitempty" binding:"omitempty,gt=0"`
	EnvironmentID       *int                   `json:"environment_id,omitempty" binding:"omitempty,gt=0"`
	QueueKey            string                 `json:"queue_key" binding:"required"`
	SourceType          string                 `json:"source_type" binding:"required"`
	SourcePlatform      string                 `json:"source_platform" binding:"required"`
	IdempotencyKey      *string                `json:"idempotency_key,omitempty"`
	DetectedProductCode *string                `json:"detected_product_code,omitempty"`
	Priority            int                    `json:"priority"`
	Logs                []IngestLogItemRequest `json:"logs" binding:"required,min=1,dive"`
}

type LogQueueBatchResponse struct {
	BatchID                 string     `json:"batch_id"`
	ProductID               *int       `json:"product_id,omitempty"`
	SourceID                *int       `json:"source_id,omitempty"`
	EnvironmentID           *int       `json:"environment_id,omitempty"`
	QueueKey                string     `json:"queue_key"`
	SourceType              string     `json:"source_type"`
	SourcePlatform          string     `json:"source_platform"`
	SourceIP                *string    `json:"source_ip,omitempty"`
	UserAgent               *string    `json:"user_agent,omitempty"`
	IdempotencyKey          *string    `json:"idempotency_key,omitempty"`
	DetectedProductCode     *string    `json:"detected_product_code,omitempty"`
	ProductResolutionStatus string     `json:"product_resolution_status"`
	ProductResolutionReason *string    `json:"product_resolution_reason,omitempty"`
	ProductResolvedAt       *time.Time `json:"product_resolved_at,omitempty"`
	AvailableAt             *time.Time `json:"available_at,omitempty"`
	WorkerID                *string    `json:"worker_id,omitempty"`
	ProcessingAttempts      int        `json:"processing_attempts"`
	TotalLogs               int        `json:"total_logs"`
	Status                  string     `json:"status"`
	Priority                int        `json:"priority"`
	ReceivedAt              *time.Time `json:"received_at,omitempty"`
	ProcessingStartedAt     *time.Time `json:"processing_started_at,omitempty"`
	ProcessedAt             *time.Time `json:"processed_at,omitempty"`
	RetentionUntil          *time.Time `json:"retention_until,omitempty"`
	PurgedAt                *time.Time `json:"purged_at,omitempty"`
	LockedBy                *string    `json:"locked_by,omitempty"`
	LockedAt                *time.Time `json:"locked_at,omitempty"`
	ErrorMessage            *string    `json:"error_message,omitempty"`
	TimestampResponse
}

type LogQueueItemResponse struct {
	QueueItemID         int64           `json:"queue_item_id"`
	BatchID             string          `json:"batch_id"`
	SequenceNo          int             `json:"sequence_no"`
	SourceType          string          `json:"source_type"`
	SourcePlatform      string          `json:"source_platform"`
	InputPayload        json.RawMessage `json:"input_payload,omitempty"`
	PayloadSizeBytes    *int64          `json:"payload_size_bytes,omitempty"`
	PayloadPurged       bool            `json:"payload_purged"`
	DetectedProductCode *string         `json:"detected_product_code,omitempty"`
	WorkerID            *string         `json:"worker_id,omitempty"`
	ProcessingAttempts  int             `json:"processing_attempts"`
	Status              string          `json:"status"`
	RetryCount          int             `json:"retry_count"`
	MaxRetryCount       int             `json:"max_retry_count"`
	NextRetryAt         *time.Time      `json:"next_retry_at,omitempty"`
	LastRetryAt         *time.Time      `json:"last_retry_at,omitempty"`
	ProcessingStartedAt *time.Time      `json:"processing_started_at,omitempty"`
	ProcessedAt         *time.Time      `json:"processed_at,omitempty"`
	ErrorMessage        *string         `json:"error_message,omitempty"`
	RetentionUntil      *time.Time      `json:"retention_until,omitempty"`
	PurgedAt            *time.Time      `json:"purged_at,omitempty"`
	TimestampResponse
}

type RetryLogFailureRequest struct {
	FailureID string `json:"failure_id" binding:"required,uuid"`
}

type ResolveLogFailureRequest struct {
	ResolutionNote *string `json:"resolution_note,omitempty"`
}

type LogFailureResponse struct {
	FailureID     string          `json:"failure_id"`
	AttemptNo     int             `json:"attempt_no"`
	QueueItemID   int64           `json:"queue_item_id"`
	BatchID       string          `json:"batch_id"`
	ProductID     *int            `json:"product_id,omitempty"`
	SourceID      *int            `json:"source_id,omitempty"`
	EnvironmentID *int            `json:"environment_id,omitempty"`
	FailureStage  string          `json:"failure_stage"`
	FailureType   string          `json:"failure_type"`
	Reason        *string         `json:"reason,omitempty"`
	ErrorDetails  json.RawMessage `json:"error_details,omitempty"`
	RetryCount    int             `json:"retry_count"`
	MaxRetryCount int             `json:"max_retry_count"`
	NextRetryAt   *time.Time      `json:"next_retry_at,omitempty"`
	LastRetryAt   *time.Time      `json:"last_retry_at,omitempty"`
	Status        string          `json:"status"`
	CreatedAt     *time.Time      `json:"created_at,omitempty"`
	ResolvedAt    *time.Time      `json:"resolved_at,omitempty"`
}

type LogIndexRefResponse struct {
	LogID              string     `json:"log_id"`
	ProductID          int        `json:"product_id"`
	ProjectID          *int       `json:"project_id,omitempty"`
	CategoryID         *int       `json:"category_id,omitempty"`
	EnvironmentID      int        `json:"environment_id"`
	SourceID           *int       `json:"source_id,omitempty"`
	BatchID            *string    `json:"batch_id,omitempty"`
	QueueItemID        *int64     `json:"queue_item_id,omitempty"`
	ResponseStatusCode *int       `json:"response_status_code,omitempty"`
	Reason             *string    `json:"reason,omitempty"`
	DurationMs         *int       `json:"duration_ms,omitempty"`
	LogLevel           *string    `json:"log_level,omitempty"`
	EventType          *string    `json:"event_type,omitempty"`
	SourceRequestID    *string    `json:"source_request_id,omitempty"`
	CorrelationID      *string    `json:"correlation_id,omitempty"`
	TraceID            *string    `json:"trace_id,omitempty"`
	SpanID             *string    `json:"span_id,omitempty"`
	ParentSpanID       *string    `json:"parent_span_id,omitempty"`
	RequestMethod      *string    `json:"request_method,omitempty"`
	RequestPath        *string    `json:"request_path,omitempty"`
	RoutePattern       *string    `json:"route_pattern,omitempty"`
	FeatureFullPath    *string    `json:"feature_full_path,omitempty"`
	FeaturePathIDs     *string    `json:"feature_path_ids,omitempty"`
	Timestamp          time.Time  `json:"timestamp"`
	IngestedAt         *time.Time `json:"ingested_at,omitempty"`
	ElasticIndex       string     `json:"elastic_index"`
	ElasticDocumentID  string     `json:"elastic_document_id"`
	IndexStatus        string     `json:"index_status"`
	IndexedAt          *time.Time `json:"indexed_at,omitempty"`
	LastSyncAt         *time.Time `json:"last_sync_at,omitempty"`
	SyncError          *string    `json:"sync_error,omitempty"`
	IsArchived         bool       `json:"is_archived"`
	ArchivedAt         *time.Time `json:"archived_at,omitempty"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
	TimestampResponse
}

type LogObjectStorageRefResponse struct {
	ObjectRefID     string     `json:"object_ref_id"`
	LogID           string     `json:"log_id"`
	ProductID       int        `json:"product_id"`
	EnvironmentID   *int       `json:"environment_id,omitempty"`
	StorageProvider string     `json:"storage_provider"`
	BucketName      *string    `json:"bucket_name,omitempty"`
	ObjectPath      string     `json:"object_path"`
	ObjectType      string     `json:"object_type"`
	FileFormat      *string    `json:"file_format,omitempty"`
	SizeBytes       *int64     `json:"size_bytes,omitempty"`
	Checksum        *string    `json:"checksum,omitempty"`
	IsEncrypted     bool       `json:"is_encrypted"`
	RetentionUntil  *time.Time `json:"retention_until,omitempty"`
	PurgedAt        *time.Time `json:"purged_at,omitempty"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
}

type CreateLogArchiveRequest struct {
	ProductID       int        `json:"product_id" binding:"required,gt=0"`
	EnvironmentID   *int       `json:"environment_id,omitempty" binding:"omitempty,gt=0"`
	ProjectID       *int       `json:"project_id,omitempty" binding:"omitempty,gt=0"`
	ArchiveYear     int        `json:"archive_year" binding:"required,min=1970"`
	ArchiveMonth    int        `json:"archive_month" binding:"required,min=1,max=12"`
	DateFrom        *time.Time `json:"date_from,omitempty"`
	DateTo          *time.Time `json:"date_to,omitempty"`
	StorageProvider string     `json:"storage_provider" binding:"required"`
	BucketName      *string    `json:"bucket_name,omitempty"`
	FileFormat      *string    `json:"file_format,omitempty"`
}

type LogArchiveResponse struct {
	ArchiveID           string     `json:"archive_id"`
	ProductID           int        `json:"product_id"`
	EnvironmentID       *int       `json:"environment_id,omitempty"`
	ProjectID           *int       `json:"project_id,omitempty"`
	ArchiveYear         int        `json:"archive_year"`
	ArchiveMonth        int        `json:"archive_month"`
	DateFrom            *time.Time `json:"date_from,omitempty"`
	DateTo              *time.Time `json:"date_to,omitempty"`
	StorageProvider     string     `json:"storage_provider"`
	BucketName          *string    `json:"bucket_name,omitempty"`
	FilePath            *string    `json:"file_path,omitempty"`
	FileFormat          *string    `json:"file_format,omitempty"`
	TotalLogs           *int       `json:"total_logs,omitempty"`
	CompressedSizeBytes *int64     `json:"compressed_size_bytes,omitempty"`
	Checksum            *string    `json:"checksum,omitempty"`
	Status              *string    `json:"status,omitempty"`
	ExportedAt          *time.Time `json:"exported_at,omitempty"`
	RestoredAt          *time.Time `json:"restored_at,omitempty"`
	DeletedAt           *time.Time `json:"deleted_at,omitempty"`
	PurgedAt            *time.Time `json:"purged_at,omitempty"`
	TimestampResponse
}

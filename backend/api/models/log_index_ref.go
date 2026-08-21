package models

import "time"

type LogIndexRef struct {
	LogID              string     `gorm:"type:uuid;primaryKey" json:"log_id"`
	ProductID          int        `gorm:"not null" json:"product_id"`
	ProjectID          *int       `json:"project_id,omitempty"`
	CategoryID         *int       `json:"category_id,omitempty"`
	EnvironmentID      int        `gorm:"not null" json:"environment_id"`
	SourceID           *int       `json:"source_id,omitempty"`
	BatchID            *string    `gorm:"type:uuid;index" json:"batch_id,omitempty"`
	QueueItemID        *int64     `gorm:"index" json:"queue_item_id,omitempty"`
	ResponseStatusCode *int       `json:"response_status_code,omitempty"`
	Reason             *string    `json:"reason,omitempty"`
	DurationMs         *int       `json:"duration_ms,omitempty"`
	LogLevel           *string    `json:"log_level,omitempty"`
	EventType          *string    `json:"event_type,omitempty"`
	SourceRequestID    *string    `gorm:"index" json:"source_request_id,omitempty"`
	CorrelationID      *string    `gorm:"index" json:"correlation_id,omitempty"`
	TraceID            *string    `gorm:"index" json:"trace_id,omitempty"`
	SpanID             *string    `json:"span_id,omitempty"`
	ParentSpanID       *string    `json:"parent_span_id,omitempty"`
	RequestMethod      *string    `json:"request_method,omitempty"`
	RequestPath        *string    `gorm:"type:text" json:"request_path,omitempty"`
	RoutePattern       *string    `gorm:"type:text" json:"route_pattern,omitempty"`
	RouteKey           *string    `gorm:"type:text;index" json:"route_key,omitempty"`
	ServiceName        *string    `gorm:"type:text;index" json:"service_name,omitempty"`
	RoutingStatus      string     `gorm:"not null;default:UNKNOWN;index" json:"routing_status"`
	RoutingMethod      *string    `gorm:"type:text" json:"routing_method,omitempty"`
	RoutingReason      *string    `gorm:"type:text" json:"routing_reason,omitempty"`
	FeatureFullPath    *string    `gorm:"type:text" json:"feature_full_path,omitempty"`
	FeaturePathIDs     *string    `json:"feature_path_ids,omitempty"`
	Timestamp          time.Time  `gorm:"type:timestamptz;not null" json:"timestamp"`
	IngestedAt         *time.Time `gorm:"type:timestamptz" json:"ingested_at,omitempty"`
	ElasticIndex       string     `gorm:"not null;uniqueIndex:uq_elastic_document,priority:1" json:"elastic_index"`
	ElasticDocumentID  string     `gorm:"not null;uniqueIndex:uq_elastic_document,priority:2" json:"elastic_document_id"`
	IndexStatus        string     `gorm:"not null;default:PENDING" json:"index_status"`
	IndexedAt          *time.Time `gorm:"type:timestamptz" json:"indexed_at,omitempty"`
	LastSyncAt         *time.Time `gorm:"type:timestamptz" json:"last_sync_at,omitempty"`
	SyncError          *string    `gorm:"type:text" json:"sync_error,omitempty"`
	IsArchived         bool       `gorm:"not null;default:false" json:"is_archived"`
	ArchivedAt         *time.Time `gorm:"type:timestamptz" json:"archived_at,omitempty"`
	DeletedAt          *time.Time `gorm:"type:timestamptz" json:"deleted_at,omitempty"`
	Timestamps
}

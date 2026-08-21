package usecase

import (
	"omnilogs-api/internal/main_logs/document"
	"time"
)

type SearchInput struct {
	ActorUserID      uint
	PlatformAdmin    bool
	ProductID        int64
	EnvironmentID    *int64
	ProjectID        *int64
	ProjectIDs       []int64
	CategoryID       *int64
	CategoryIDs      []int64
	Level            *string
	LogType          *string
	RequestID        *string
	TraceID          *string
	DateFrom         *time.Time
	StatusCode       *int64
	SourceID         *int64
	CustomFieldPath  *string
	CustomFieldValue *string
	SortField        *string
	SortOrder        string
	Keyword          *string
	Page             int
	PerPage          int
}

type MainLogDocument = document.MainLog

type SearchResult struct {
	Items []MainLogDocument
	Total int64
}

func uintToInt64Ptr(value uint) *int64 {
	result := int64(value)
	return &result
}

type SearchModelCondition struct {
	Type     string `json:"type"` // "condition"
	Field    string `json:"field"`
	DataType string `json:"data_type"` // "text", "number", "keyword"
	Operator string `json:"operator"`  // "eq", "neq", "contains", "gt", "gte", "lt", "lte", "exists", "not_exists"
	Value    any    `json:"value"`
	Enabled  bool   `json:"enabled"`
}

type SearchModelGroup struct {
	Type     string                 `json:"type"`     // "group"
	Operator string                 `json:"operator"` // "AND", "OR"
	Children []SearchModelCondition `json:"children"`
}

type SearchModelScope struct {
	ProductID     int64  `json:"product_id"`
	EnvironmentID *int64 `json:"environment_id,omitempty"`
	ProjectID     *int64 `json:"project_id,omitempty"`
	CategoryID    *int64 `json:"category_id,omitempty"`
	ArchiveID     string `json:"archive_id,omitempty"`
}

type SearchModelTimeRange struct {
	Field string `json:"field"`
	Type  string `json:"type"`
	Value string `json:"value"` // "24h", "all", etc.
}

type DynamicSearchInput struct {
	Version   int                  `json:"version"`
	Scope     SearchModelScope     `json:"scope"`
	TimeRange SearchModelTimeRange `json:"time_range"`
	Root      SearchModelGroup     `json:"root"`
	PageSize  int                  `json:"page_size,omitempty"`
	Page      int                  `json:"page,omitempty"`
	SortField *string              `json:"sort_field,omitempty"`
	SortOrder string               `json:"sort_order,omitempty"`

	// Internal context fields (injected by handler)
	ActorUserID   uint `json:"-"`
	PlatformAdmin bool `json:"-"`
}

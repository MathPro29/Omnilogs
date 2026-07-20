package usecase

import "omnilogs-api/internal/main_logs/document"

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

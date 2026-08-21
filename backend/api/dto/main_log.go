package dto

// MainLogSearchInput contains normalized search and authorization parameters.
// Query-string parsing remains in the HTTP handler so this DTO can also be used
// by non-HTTP callers.
type MainLogSearchInput struct {
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

// MainLogResponse is the stable API representation of a main log document.
type MainLogResponse struct {
	LogID           string         `json:"log_id"`
	ProductID       int64          `json:"product_id"`
	EnvironmentID   *int64         `json:"environment_id,omitempty"`
	SourceID        *int64         `json:"source_id,omitempty"`
	Timestamp       *string        `json:"timestamp,omitempty"`
	Level           *string        `json:"level,omitempty"`
	LogType         *string        `json:"log_type,omitempty"`
	Message         *string        `json:"message,omitempty"`
	RequestID       *string        `json:"request_id,omitempty"`
	TraceID         *string        `json:"trace_id,omitempty"`
	Method          *string        `json:"method,omitempty"`
	Path            *string        `json:"path,omitempty"`
	URL             *string        `json:"url,omitempty"`
	StatusCode      *int64         `json:"status_code,omitempty"`
	LatencyMs       *int64         `json:"latency_ms,omitempty"`
	RequestHeaders  map[string]any `json:"request_headers,omitempty"`
	ResponseHeaders map[string]any `json:"response_headers,omitempty"`
	RequestPayload  any            `json:"request_payload,omitempty"`
	ResponsePayload any            `json:"response_payload,omitempty"`
	ErrorCode       *string        `json:"error_code,omitempty"`
	ErrorMessage    *string        `json:"error_message,omitempty"`
	StackTrace      *string        `json:"stack_trace,omitempty"`
	CustomFields    map[string]any `json:"custom_fields,omitempty"`
	Raw             map[string]any `json:"raw"`
}

// MainLogSearchResult carries result data before the response envelope and
// pagination metadata are added by the handler.
type MainLogSearchResult struct {
	Items []MainLogResponse
	Total int64
}

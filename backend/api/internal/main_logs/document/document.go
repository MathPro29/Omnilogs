package document

import (
	"encoding/json"
	"strconv"
	"strings"
)

// MainLog represents the stable API shape returned by log search and detail
// endpoints. Raw intentionally retains the original Elasticsearch document.
type MainLog struct {
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

func Map(logID string, source map[string]any) MainLog {
	value := MainLog{LogID: logID, Raw: source}
	if source == nil {
		return value
	}

	value.ProductID = int64FromValue(source["product_id"])
	value.EnvironmentID = optionalInt64(source["environment_id"])
	value.SourceID = optionalInt64(source["source_id"])
	value.Timestamp = stringPointer(source["@timestamp"])

	payload, _ := source["payload"].(map[string]any)
	value.Level = stringPointer(payload["log_level"])
	value.LogType = stringPointer(payload["event_type"])
	value.Message = stringPointer(payload["message"])
	value.RequestID = stringPointer(payload["source_request_id"])
	value.TraceID = stringPointer(payload["trace_id"])
	value.Method = stringPointer(payload["request_method"])
	value.Path = stringPointer(payload["request_path"])
	value.URL = stringPointer(payload["url"])
	value.StatusCode = optionalInt64(payload["status_code"])
	value.LatencyMs = optionalInt64(payload["duration_ms"])
	value.ErrorCode = stringPointer(payload["error_code"])
	value.ErrorMessage = stringPointer(payload["error_message"])
	value.StackTrace = stringPointer(payload["stack_trace"])
	value.RequestHeaders, _ = payload["request_headers"].(map[string]any)
	value.ResponseHeaders, _ = payload["response_headers"].(map[string]any)
	value.RequestPayload = payload["request_payload"]
	value.ResponsePayload = payload["response_payload"]
	value.CustomFields, _ = payload["custom_fields"].(map[string]any)
	return value
}

func int64FromValue(value any) int64 {
	switch number := value.(type) {
	case int:
		return int64(number)
	case int32:
		return int64(number)
	case int64:
		return number
	case uint:
		if uint64(number) <= uint64(^uint64(0)>>1) {
			return int64(number)
		}
	case uint32:
		return int64(number)
	case uint64:
		if number <= uint64(^uint64(0)>>1) {
			return int64(number)
		}
	case float64:
		return int64(number)
	case float32:
		return int64(number)
	case json.Number:
		parsed, err := number.Int64()
		if err == nil {
			return parsed
		}
	case string:
		parsed, err := strconv.ParseInt(number, 10, 64)
		if err == nil {
			return parsed
		}
	}
	return 0
}

func optionalInt64(value any) *int64 {
	if value == nil {
		return nil
	}
	result := int64FromValue(value)
	return &result
}

func stringPointer(value any) *string {
	text, ok := value.(string)
	if !ok || strings.TrimSpace(text) == "" {
		return nil
	}
	return &text
}

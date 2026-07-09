package usecase

import (
	"strings"
)

func mapMainLog(logID string, source map[string]any) MainLogDocument {
	value := MainLogDocument{
		LogID: logID,
		Raw:   source,
	}
	if source == nil {
		return value
	}
	value.ProductID = int64FromMap(source, "product_id")
	value.EnvironmentID = optionalInt64FromMap(source, "environment_id")
	value.SourceID = optionalInt64FromMap(source, "source_id")
	value.Timestamp = stringPtrFromMap(source, "@timestamp")
	value.Level = stringPtrFromPayload(source, "log_level")
	value.LogType = stringPtrFromPayload(source, "event_type")
	value.Message = stringPtrFromPayload(source, "message")
	value.RequestID = stringPtrFromPayload(source, "source_request_id")
	value.TraceID = stringPtrFromPayload(source, "trace_id")
	value.Method = stringPtrFromPayload(source, "request_method")
	value.Path = stringPtrFromPayload(source, "request_path")
	value.URL = stringPtrFromPayload(source, "url")
	value.StatusCode = optionalInt64FromPayload(source, "status_code")
	value.LatencyMs = optionalInt64FromPayload(source, "duration_ms")
	value.ErrorCode = stringPtrFromPayload(source, "error_code")
	value.ErrorMessage = stringPtrFromPayload(source, "error_message")
	value.StackTrace = stringPtrFromPayload(source, "stack_trace")
	if payload, ok := source["payload"].(map[string]any); ok {
		value.RequestHeaders = mapFromValue(payload["request_headers"])
		value.ResponseHeaders = mapFromValue(payload["response_headers"])
		value.RequestPayload = payload["request_payload"]
		value.ResponsePayload = payload["response_payload"]
		value.CustomFields = mapFromValue(payload["custom_fields"])
	}
	return value
}

func uintToInt64Ptr(v uint) *int64 {
	result := int64(v)
	return &result
}

func mapFromValue(value any) map[string]any {
	result, _ := value.(map[string]any)
	return result
}

func int64FromMap(source map[string]any, key string) int64 {
	if value, ok := source[key].(float64); ok {
		return int64(value)
	}
	return 0
}

func optionalInt64FromMap(source map[string]any, key string) *int64 {
	if value, ok := source[key].(float64); ok {
		result := int64(value)
		return &result
	}
	return nil
}

func stringPtrFromMap(source map[string]any, key string) *string {
	value, ok := source[key].(string)
	if !ok || strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func stringPtrFromPayload(source map[string]any, key string) *string {
	payload, ok := source["payload"].(map[string]any)
	if !ok {
		return nil
	}
	return stringPtrFromMap(payload, key)
}

func optionalInt64FromPayload(source map[string]any, key string) *int64 {
	payload, ok := source["payload"].(map[string]any)
	if !ok {
		return nil
	}
	return optionalInt64FromMap(payload, key)
}

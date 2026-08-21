package usecase

import (
	"encoding/json"
	"fmt"
)

type PayloadLimits struct {
	MaxPayloadSizeBytes int
	MaxJSONDepth        int
	MaxFieldCount       int
	MaxStringLength     int
	MaxArrayLength      int
}

type PayloadLimitError struct {
	Code     string
	Message  string
	Limit    int
	Received int
}

func (e *PayloadLimitError) Error() string {
	return fmt.Sprintf("%s (maximum=%d, received=%d)", e.Message, e.Limit, e.Received)
}

func validatePayload(document json.RawMessage, limits PayloadLimits) error {
	if limits.MaxPayloadSizeBytes > 0 && len(document) > limits.MaxPayloadSizeBytes {
		return &PayloadLimitError{
			Code:     "PAYLOAD_TOO_LARGE",
			Message:  "Log payload exceeds the maximum allowed size",
			Limit:    limits.MaxPayloadSizeBytes,
			Received: len(document),
		}
	}

	var value any
	if err := json.Unmarshal(document, &value); err != nil {
		return err
	}
	fieldCount := 0
	return inspectJSONValue(value, 1, &fieldCount, limits)
}

func inspectJSONValue(value any, depth int, fieldCount *int, limits PayloadLimits) error {
	if limits.MaxJSONDepth > 0 && depth > limits.MaxJSONDepth {
		return &PayloadLimitError{Code: "JSON_DEPTH_EXCEEDED", Message: "Log payload exceeds the maximum JSON depth", Limit: limits.MaxJSONDepth, Received: depth}
	}
	switch typed := value.(type) {
	case map[string]any:
		*fieldCount += len(typed)
		if limits.MaxFieldCount > 0 && *fieldCount > limits.MaxFieldCount {
			return &PayloadLimitError{Code: "FIELD_COUNT_EXCEEDED", Message: "Log payload exceeds the maximum field count", Limit: limits.MaxFieldCount, Received: *fieldCount}
		}
		for _, child := range typed {
			if err := inspectJSONValue(child, depth+1, fieldCount, limits); err != nil {
				return err
			}
		}
	case []any:
		if limits.MaxArrayLength > 0 && len(typed) > limits.MaxArrayLength {
			return &PayloadLimitError{Code: "ARRAY_LENGTH_EXCEEDED", Message: "Log payload exceeds the maximum array length", Limit: limits.MaxArrayLength, Received: len(typed)}
		}
		for _, child := range typed {
			if err := inspectJSONValue(child, depth+1, fieldCount, limits); err != nil {
				return err
			}
		}
	case string:
		if limits.MaxStringLength > 0 && len([]rune(typed)) > limits.MaxStringLength {
			return &PayloadLimitError{Code: "STRING_LENGTH_EXCEEDED", Message: "Log payload contains a string longer than allowed", Limit: limits.MaxStringLength, Received: len([]rune(typed))}
		}
	}
	return nil
}

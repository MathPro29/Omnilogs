package usecase

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestValidatePayloadLimits(t *testing.T) {
	limits := PayloadLimits{
		MaxPayloadSizeBytes: 128,
		MaxJSONDepth:        3,
		MaxFieldCount:       3,
		MaxStringLength:     8,
		MaxArrayLength:      2,
	}
	tests := []struct {
		name string
		raw  string
		code string
	}{
		{name: "valid", raw: `{"level":"info"}`},
		{name: "depth", raw: `{"a":{"b":{"c":1}}}`, code: "JSON_DEPTH_EXCEEDED"},
		{name: "fields", raw: `{"a":1,"b":2,"c":3,"d":4}`, code: "FIELD_COUNT_EXCEEDED"},
		{name: "string", raw: `{"message":"123456789"}`, code: "STRING_LENGTH_EXCEEDED"},
		{name: "array", raw: `{"items":[1,2,3]}`, code: "ARRAY_LENGTH_EXCEEDED"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validatePayload(json.RawMessage(test.raw), limits)
			if test.code == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var limitErr *PayloadLimitError
			if !errors.As(err, &limitErr) || limitErr.Code != test.code {
				t.Fatalf("error = %#v, want %s", err, test.code)
			}
		})
	}
}

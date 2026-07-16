package document

import (
	"encoding/json"
	"testing"
)

func TestMapSupportsElasticsearchAndPostgresNumericTypes(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{name: "elasticsearch float", value: float64(42)},
		{name: "postgres int", value: int(42)},
		{name: "int64", value: int64(42)},
		{name: "json number", value: json.Number("42")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapped := Map("log-1", map[string]any{
				"product_id":     tt.value,
				"environment_id": tt.value,
				"payload": map[string]any{
					"status_code": tt.value,
					"duration_ms": tt.value,
				},
			})
			if mapped.ProductID != 42 || mapped.EnvironmentID == nil || *mapped.EnvironmentID != 42 {
				t.Fatalf("unexpected top-level IDs: %+v", mapped)
			}
			if mapped.StatusCode == nil || *mapped.StatusCode != 42 || mapped.LatencyMs == nil || *mapped.LatencyMs != 42 {
				t.Fatalf("unexpected payload numbers: %+v", mapped)
			}
		})
	}
}

func TestMapHandlesMissingPayload(t *testing.T) {
	mapped := Map("log-1", map[string]any{"product_id": int64(7)})
	if mapped.ProductID != 7 || mapped.Message != nil || mapped.Raw == nil {
		t.Fatalf("unexpected mapping: %+v", mapped)
	}
}

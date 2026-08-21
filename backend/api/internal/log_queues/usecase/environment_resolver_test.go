package usecase

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestUniqueEnvironmentCandidate(t *testing.T) {
	tests := []struct {
		name       string
		candidates map[int]string
		want       int
		wantErr    error
	}{
		{name: "one explicit environment", candidates: map[int]string{10: "request.environment_id"}, want: 10},
		{name: "same resolved environment remains one candidate", candidates: map[int]string{20: "api_key.environment_id"}, want: 20},
		{name: "missing mapping", candidates: map[int]string{}, wantErr: ErrUnmappedScope},
		{name: "DEV and PROD are ambiguous", candidates: map[int]string{1: "DEV", 2: "PROD"}, wantErr: ErrAmbiguousScope},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := uniqueEnvironmentCandidate(tt.candidates)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("environment = %d, error = %v, want %d", got, err, tt.want)
			}
		})
	}
}

func TestPayloadEnvironmentHintsSupportLegacyAndExplicitShapes(t *testing.T) {
	payload := map[string]any{
		"environment_code": "DEV",
		"metadata":         map[string]any{"environmentId": json.Number("7")},
		"routing": map[string]any{
			"explicitHierarchy": map[string]any{"environmentCode": "PROD"},
		},
	}
	hints := payloadEnvironmentHints(payload)
	if len(hints) != 3 {
		t.Fatalf("hints = %#v, want 3 scoped values", hints)
	}
	if hints[0].Code != "DEV" || hints[1].ID == nil || *hints[1].ID != 7 || hints[2].Code != "PROD" {
		t.Fatalf("unexpected environment hints: %#v", hints)
	}
}

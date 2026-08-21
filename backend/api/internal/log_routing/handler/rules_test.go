package handler

import "testing"

func TestValidRoutingConditions(t *testing.T) {
	tests := []struct {
		name       string
		conditions []routingConditionInput
		want       bool
	}{
		{
			name: "accepts stable HTTP fields",
			conditions: []routingConditionInput{
				{Field: "service", Operator: "equals", Value: "asw-web"},
				{Field: "request_path", Operator: "starts_with", Value: "/api/members"},
			},
			want: true,
		},
		{
			name:       "accepts payload operator",
			conditions: []routingConditionInput{{Field: "request_path", Operator: "contains", Value: "/api"}},
			want:       true,
		},
		{
			name:       "rejects unknown operator",
			conditions: []routingConditionInput{{Field: "request_path", Operator: "script", Value: "/api"}},
			want:       false,
		},
		{
			name:       "accepts range with bounds",
			conditions: []routingConditionInput{{Field: "response.status_code", Operator: "range", Min: 400, Max: 499}},
			want:       true,
		},
		{
			name:       "rejects invalid nested field",
			conditions: []routingConditionInput{{Field: "request-path", Operator: "equals", Value: "/api"}},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validRoutingConditions(tt.conditions); got != tt.want {
				t.Fatalf("validRoutingConditions() = %v, want %v", got, tt.want)
			}
		})
	}
}

package handler

import (
	"encoding/json"
	"testing"

	"omnilogs-api/dto"
)

func TestValidatePayloadScopeAllowsRoutingProjectToVary(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{name: "top level", data: `{"project_id":20}`},
		{name: "explicit camel case", data: `{"routing":{"explicitHierarchy":{"projectId":20}}}`},
		{name: "explicit snake case", data: `{"routing":{"explicit_hierarchy":{"project_id":20}}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs := []dto.IngestLogItemRequest{{Data: json.RawMessage(tt.data)}}
			if err := validatePayloadScope(logs, 1, 2); err != nil {
				t.Fatalf("routing project must not be constrained by API key: %v", err)
			}
		})
	}
}

func TestValidatePayloadScopeAllowsCustomFieldProject(t *testing.T) {
	logs := []dto.IngestLogItemRequest{{Data: json.RawMessage(`{"custom_fields":{"project_id":20}}`)}}
	if err := validatePayloadScope(logs, 1, 2); err != nil {
		t.Fatalf("custom project discriminator must not be constrained by API key: %v", err)
	}
}
func TestValidatePayloadScopeAllowsSameProject(t *testing.T) {
	logs := []dto.IngestLogItemRequest{{Data: json.RawMessage(`{"project_id":10,"category_code":"SHARED"}`)}}
	if err := validatePayloadScope(logs, 1, 2); err != nil {
		t.Fatalf("same-project payload rejected: %v", err)
	}
}

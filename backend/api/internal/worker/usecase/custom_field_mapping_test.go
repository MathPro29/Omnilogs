package usecase

import (
	"testing"

	"omnilogs-api/models"
)

func TestApplyConfiguredCustomFieldsMatchesSourceKeyAndUsesDefault(t *testing.T) {
	defaultValue := "unknown"
	sourceKey := "source_tier"
	fields := []models.LogFieldDefinition{{FieldKey: "customer_tier", ValueSourceKey: &sourceKey}, {FieldKey: "region", DefaultValue: &defaultValue}}
	payload := map[string]any{"source_tier": "gold"}
	applyConfiguredCustomFields(payload, fields)
	customFields := payload["custom_fields"].(map[string]any)
	if got := customFields["customer_tier"]; got != "gold" {
		t.Fatalf("customer_tier = %#v, want gold", got)
	}
	if got := customFields["region"]; got != "unknown" {
		t.Fatalf("region = %#v, want default unknown", got)
	}
}

func TestApplyConfiguredCustomFieldsPreservesExplicitCustomField(t *testing.T) {
	defaultValue := "unknown"
	payload := map[string]any{"custom_fields": map[string]any{"region": "apac"}}
	applyConfiguredCustomFields(payload, []models.LogFieldDefinition{{FieldKey: "region", DefaultValue: &defaultValue}})
	if got := payload["custom_fields"].(map[string]any)["region"]; got != "apac" {
		t.Fatalf("region = %#v, want explicit value apac", got)
	}
}

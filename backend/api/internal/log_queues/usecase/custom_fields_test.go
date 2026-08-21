package usecase

import "testing"

func TestMergeCustomFieldsKeepsPayloadValues(t *testing.T) {
	payload := map[string]any{"custom_fields": map[string]any{"region": "apac"}}
	mergeCustomFields(payload, map[string]any{"region": "eu", "tier": "gold"})
	fields := payload["custom_fields"].(map[string]any)
	if fields["region"] != "apac" || fields["tier"] != "gold" {
		t.Fatalf("unexpected merged custom fields: %#v", fields)
	}
}

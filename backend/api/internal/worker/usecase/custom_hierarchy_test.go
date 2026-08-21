package usecase

import "testing"

func TestNormalizeCustomHierarchyUsesSourceFields(t *testing.T) {
	payload := map[string]any{
		"request_path": "/api/anything",
		"custom_fields": map[string]any{
			"productCode":      "ASSETWISE",
			"environment_code": "UAT",
			"project_code":     "CONDO",
			"category_code":    "FINANCE",
			"feature_code":     "RECEIPTS",
			"subFeatureCode":   "CREATE",
		},
	}
	if !normalizeCustomHierarchy(payload) {
		t.Fatal("expected custom hierarchy to be detected")
	}
	if payload["project_code"] != "CONDO" || payload["category_code"] != "FINANCE" || payload["sub_feature_code"] != "CREATE" {
		t.Fatalf("unexpected canonical hierarchy: %#v", payload)
	}
	fields := payload["custom_fields"].(map[string]any)
	if fields["product_code"] != "ASSETWISE" || fields["environment_code"] != "UAT" || fields["sub_feature_code"] != "CREATE" {
		t.Fatalf("unexpected canonical custom fields: %#v", fields)
	}
}

func TestNormalizeCustomHierarchyDoesNotInferFromRequestPath(t *testing.T) {
	payload := map[string]any{"request_path": "/api/finance/receipts/create"}
	if normalizeCustomHierarchy(payload) {
		t.Fatal("request path must not be treated as hierarchy")
	}
	if _, exists := payload["project_code"]; exists {
		t.Fatal("request path must not create project mapping")
	}
}
func TestNormalizeCustomHierarchyUsesRoutingContext(t *testing.T) {
	payload := map[string]any{"routing_context": map[string]any{
		"project_code": "project_a", "feature_code": "others_bill", "sub_feature_code": "internet_bill",
		"type": "bill", "value": "internet",
	}}
	if !normalizeCustomHierarchy(payload) {
		t.Fatal("expected routing_context hierarchy to be detected")
	}
	for key, want := range map[string]any{
		"project_code": "project_a", "feature_code": "others_bill", "sub_feature_code": "internet_bill",
		"routing_type": "bill", "routing_value": "internet",
	} {
		if got := payload[key]; got != want {
			t.Fatalf("%s = %#v, want %#v", key, got, want)
		}
	}
}

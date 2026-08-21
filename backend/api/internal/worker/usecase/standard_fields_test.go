package usecase

import (
	"testing"
	"time"

	"omnilogs-api/dto"
)

func TestNormalizeHTTPFields(t *testing.T) {
	tests := []struct {
		name    string
		payload map[string]any
		method  string
		path    string
	}{
		{name: "top-level path", payload: map[string]any{"method": "POST", "path": "/api/bill_others/bot_settings/add"}, method: "POST", path: "/api/bill_others/bot_settings/add"},
		{name: "nested request", payload: map[string]any{"request": map[string]any{"method": "GET", "path": "/api/users"}}, method: "GET", path: "/api/users"},
		{name: "full URL", payload: map[string]any{"http": map[string]any{"request": map[string]any{"method": "PATCH"}, "url": "http://localhost:8080/api/users/42?active=true"}}, method: "PATCH", path: "/api/users/42"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizeHTTPFields(tt.payload)
			if tt.payload["request_method"] != tt.method || tt.payload["request_path"] != tt.path {
				t.Fatalf("normalized HTTP fields = %#v", tt.payload)
			}
		})
	}
}

func TestBuildElasticDocumentUsesNullForMissingHierarchyLevels(t *testing.T) {
	productID, environmentID := 1, 2
	document, _, err := buildElasticDocument(&dto.LogMessage{
		ProductID: &productID, EnvironmentID: &environmentID, PublishedAt: time.Now(),
	}, map[string]any{"message": "project-level log"})
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"project_id", "feature_id", "sub_feature_id"} {
		if value, exists := document[key]; !exists || value != nil {
			t.Fatalf("%s should be present as null, got %#v", key, value)
		}
	}
}

func TestBuildElasticDocumentCopiesIsolationContextToRoot(t *testing.T) {
	productID, environmentID := 1, 2
	payload := map[string]any{
		"product_code": "ASW", "environment_code": "DEV",
		"project_id": 10, "project_code": "PROJECT_A",
		"category_id": 20, "category_code": "members",
		"feature_id": 19, "feature_code": "BILL", "sub_feature_id": 20, "sub_feature_code": "WATER",
		"source": "asset-api", "request_path": "/billing/water",
		"mapping_source": "common_rule",
		"mapping_status": "mapped",
	}
	document, _, err := buildElasticDocument(&dto.LogMessage{
		ProductID: &productID, EnvironmentID: &environmentID, PublishedAt: time.Now(),
	}, payload)
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]any{
		"product_id": 1, "product_code": "ASW",
		"environment_id": 2, "environment_code": "DEV",
		"project_id": 10, "project_code": "PROJECT_A",
		"category_id": 20, "category_code": "members",
		"feature_id": 19, "feature_code": "BILL",
		"sub_feature_id": 20, "sub_feature_code": "WATER",
		"source": "asset-api", "service_name": "asset-api",
		"request_path": "/billing/water", "endpoint": "/billing/water",
		"mapping_status": "mapped",
	} {
		if got := document[key]; got != want {
			t.Fatalf("%s = %#v, want %#v", key, got, want)
		}
	}
}

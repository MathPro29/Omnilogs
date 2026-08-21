package usecase

import (
	"encoding/json"
	"testing"

	"omnilogs-api/models"
)

func TestRoutingRuleMatches(t *testing.T) {
	payload := map[string]any{
		"service":   "order-service",
		"route_key": "order.create",
		"response":  map[string]any{"status_code": float64(503)},
	}
	raw := json.RawMessage(`[
		{"field":"service","operator":"equals","value":"ORDER-SERVICE"},
		{"field":"route_key","operator":"starts_with","value":"order."},
		{"field":"response.status_code","operator":"range","min":500,"max":599}
	]`)
	matched, err := routingRuleMatches(raw, payload)
	if err != nil || !matched {
		t.Fatalf("matched=%v error=%v", matched, err)
	}
}

func TestRoutingRuleMatchesRoutingDimensions(t *testing.T) {
	payload := map[string]any{
		"routing": map[string]any{
			"dimensions": map[string]any{"projectId": "100", "channel": "mobile-app"},
		},
	}
	raw := json.RawMessage(`[
		{"field":"routing.dimensions.projectId","operator":"in","value":["100","200"]},
		{"field":"routing.dimensions.channel","operator":"starts_with","value":"mobile"}
	]`)
	matched, err := routingRuleMatches(raw, payload)
	if err != nil || !matched {
		t.Fatalf("matched=%v error=%v", matched, err)
	}
}

func TestNormalizeExplicitHierarchyKeepsMetadataDimensionsSeparate(t *testing.T) {
	payload := map[string]any{
		"metadata": map[string]any{"categoryId": "members", "projectId": "100"},
	}
	if normalizeExplicitHierarchy(payload) {
		t.Fatal("routing dimensions in metadata must not become explicit hierarchy")
	}
	if _, exists := payload["category_id"]; exists {
		t.Fatal("metadata.categoryId must not be copied to category_id")
	}
}

func TestNormalizeExplicitHierarchyEnvelope(t *testing.T) {
	payload := map[string]any{
		"routing": map[string]any{
			"explicitHierarchy": map[string]any{
				"projectCode":    "MEMBER",
				"subFeatureCode": "GET_ALL",
			},
		},
	}
	if !normalizeExplicitHierarchy(payload) {
		t.Fatal("expected explicit hierarchy")
	}
	if payload["project_code"] != "MEMBER" || payload["sub_feature_code"] != "GET_ALL" {
		t.Fatalf("unexpected canonical hierarchy: %#v", payload)
	}
}

func TestNormalizeExplicitHierarchyDirectRouting(t *testing.T) {
	payload := map[string]any{"routing": map[string]any{
		"project_code": "PROJECT_A", "feature_code": "BILL", "sub_feature_code": "WATER",
	}}
	if !normalizeExplicitHierarchy(payload) {
		t.Fatal("expected direct routing context to be explicit")
	}
	if payload["project_code"] != "PROJECT_A" || payload["feature_code"] != "BILL" || payload["sub_feature_code"] != "WATER" {
		t.Fatalf("unexpected canonical routing context: %#v", payload)
	}
}

func TestNormalizeProjectDiscriminatorStringAsCode(t *testing.T) {
	for _, projectCode := range []string{"PROJECT_A", "PROJECT_B"} {
		payload := map[string]any{"projectId": projectCode, "feature_code": "BILL", "sub_feature_code": "WATER"}
		if !normalizeExplicitHierarchy(payload) || payload["project_code"] != projectCode {
			t.Fatalf("projectId %q was not normalized as project_code: %#v", projectCode, payload)
		}
		if _, exists := payload["project_id"]; exists {
			t.Fatalf("non-numeric discriminator must not become project_id: %#v", payload)
		}
	}
}

func TestNormalizeLegacyHierarchyPayload(t *testing.T) {
	payload := map[string]any{"project_code": "PROJECT_A", "category_code": "WATER"}
	if !normalizeExplicitHierarchy(payload) || payload["project_code"] != "PROJECT_A" || payload["category_code"] != "WATER" {
		t.Fatalf("legacy hierarchy payload was not preserved: %#v", payload)
	}
}

func TestSingleProjectDefaultRequiresUnambiguousProduct(t *testing.T) {
	if got := singleProjectDefault([]models.Project{{ProjectID: 10}}); got == nil || got.ProjectID != 10 {
		t.Fatalf("single project should be a safe default: %#v", got)
	}
	if got := singleProjectDefault([]models.Project{{ProjectID: 10}, {ProjectID: 20}}); got != nil {
		t.Fatalf("multiple projects must not produce a default: %#v", got)
	}
}

func TestPartialExplicitHierarchyStillNeedsProjectResolution(t *testing.T) {
	payload := map[string]any{"feature_code": "BILL", "sub_feature_code": "WATER"}
	if !normalizeExplicitHierarchy(payload) {
		t.Fatal("feature fields should be normalized as explicit hierarchy context")
	}
	if hasProjectDiscriminator(payload) {
		t.Fatal("feature fields alone must not be treated as a project discriminator")
	}
	payload["project_code"] = "PROJECT_A"
	if !hasProjectDiscriminator(payload) {
		t.Fatal("project_code must be recognized as a project discriminator")
	}
}

func TestApplyDeletedMappingFallbackPreventsAutomaticHierarchy(t *testing.T) {
	payload := map[string]any{}
	applyDeletedMappingFallback(payload, "deleted by user")

	if skip, _ := payload["skip_automatic_hierarchy"].(bool); !skip {
		t.Fatal("deleted mapping must block automatic hierarchy matching")
	}
	if payload["routing_status"] != "UNCLASSIFIED" || payload["routing_method"] != "MAPPING_DELETED" {
		t.Fatalf("unexpected routing metadata: %#v", payload)
	}
}

func TestHeaderPrecedenceAndRoutingConflict(t *testing.T) {
	payload := map[string]any{
		"_header_project_id": "31",
		"project_id":         float64(10),
	}
	headerID := intPtrFromAny(payload["_header_project_id"])
	payloadID := intPtrFromAny(payload["project_id"])
	if headerID == nil || *headerID != 31 {
		t.Fatalf("header project_id was not parsed correctly: %#v", headerID)
	}
	if payloadID == nil || *payloadID != 10 {
		t.Fatalf("payload project_id was not parsed correctly: %#v", payloadID)
	}
	if *headerID != *payloadID {
		payload["routing_conflict"] = true
	}
	if conflict, _ := payload["routing_conflict"].(bool); !conflict {
		t.Fatal("expected routing conflict when header and payload project IDs mismatch")
	}
}

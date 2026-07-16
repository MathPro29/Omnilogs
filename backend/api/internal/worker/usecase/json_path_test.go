package usecase

import (
	"testing"

	"omnilogs-api/models"
)

func TestReadAndWriteNestedJSONPath(t *testing.T) {
	payload := map[string]any{
		"request": map[string]any{
			"user": map[string]any{"email": "old@example.com"},
			"items": []any{
				map[string]any{"sku": "A"},
				map[string]any{"sku": "B"},
			},
		},
	}
	value, found := readJSONPath(payload, "request.user.email")
	if !found || value != "old@example.com" {
		t.Fatalf("nested path = %#v, found=%v", value, found)
	}
	arrayValue, found := readJSONPath(payload, "request.items[].sku")
	if !found || len(arrayValue.([]any)) != 2 {
		t.Fatalf("array path = %#v, found=%v", arrayValue, found)
	}
	if err := writeJSONPath(payload, "request.user.email", "new@example.com"); err != nil {
		t.Fatal(err)
	}
	if value, _ := readJSONPath(payload, "request.user.email"); value != "new@example.com" {
		t.Fatalf("nested write = %#v", value)
	}
}

func TestNormalizeArrayPathForFavoriteMatching(t *testing.T) {
	if got := normalizeArrayPath("request.items[12].sku"); got != "request.items[].sku" {
		t.Fatalf("normalized path = %s", got)
	}
}

func TestReadJSONPathSupportsPayloadRoot(t *testing.T) {
	payload := map[string]any{"user": map[string]any{"email": "a@example.com"}}
	value, found := readJSONPath(payload, "$")
	if !found || value == nil {
		t.Fatalf("expected payload root to be readable, found=%v value=%v", found, value)
	}
}

func TestCanonicalDynamicPathSupportsLegacyAndDataRoots(t *testing.T) {
	for input, want := range map[string]string{
		"payload.request.id":  "request.id",
		"data.request.id":     "request.id",
		"raw.data.user.id":    "user.id",
		"raw.payload.user.id": "user.id",
		"fields.user.id":      "user.id",
		"data":                "$",
	} {
		if got := canonicalDynamicPath(input); got != want {
			t.Fatalf("canonicalDynamicPath(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestSensitiveMatchersUseCanonicalPathsBeforeKeyFallback(t *testing.T) {
	fieldPath := "raw.data.request.user.email"
	rulePath := "payload.request.items[].token"
	fieldKey := "token"
	fields := []models.LogFieldDefinition{{FieldKey: "email", FieldPath: &fieldPath}}
	rules := []models.LogMaskingRule{{FieldKey: &fieldKey, FieldPath: &rulePath}}
	matchers := buildSensitiveMatchers(fields, rules)

	if field := matchers.matchField("email", "request.user.email"); field == nil {
		t.Fatal("expected canonical field path to match")
	}
	if field := matchers.matchField("email", "response.user.email"); field != nil {
		t.Fatal("path-scoped field must not mask the same key at another path")
	}
	if rule := matchers.matchRule("token", "request.items[3].token"); rule == nil {
		t.Fatal("expected canonical array rule path to match")
	}
	if rule := matchers.matchRule("token", "response.token"); rule != nil {
		t.Fatal("path-scoped rule must not mask the same key at another path")
	}
}

package usecase

import "testing"

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

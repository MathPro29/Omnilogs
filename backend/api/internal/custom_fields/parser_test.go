package custom_fields

import "testing"

func TestParseJSONFieldsNestedAndArray(t *testing.T) {
	nodes, err := ParseJSONFields([]byte(`{"request":{"user":{"email":"a@example.com"},"items":[{"sku":"A","qty":2}]},"active":true}`))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	paths := map[string]bool{}
	var walk func([]JSONFieldNode)
	walk = func(items []JSONFieldNode) {
		for _, item := range items {
			paths[item.JSONPath] = true
			walk(item.Children)
		}
	}
	walk(nodes)
	for _, path := range []string{"request.user.email", "request.items[].sku", "request.items[].qty", "active"} {
		if !paths[path] {
			t.Fatalf("missing path %s", path)
		}
	}
}

func TestParseJSONFieldsDetectsTypes(t *testing.T) {
	nodes, err := ParseJSONFields([]byte(`{"count":2,"ok":true,"at":"2026-07-14T10:00:00Z","meta":{},"items":[]}`))
	if err != nil {
		t.Fatal(err)
	}
	types := map[string]string{}
	for _, node := range nodes {
		types[node.JSONPath] = node.DataType
	}
	for path, expected := range map[string]string{"count": "number", "ok": "boolean", "at": "datetime", "meta": "object", "items": "array"} {
		if types[path] != expected {
			t.Fatalf("%s type = %s, want %s", path, types[path], expected)
		}
	}
}

func TestFieldKeyFromPathIsSafe(t *testing.T) {
	if key := FieldKeyFromPath("request.users[].email"); key != "email" {
		t.Fatalf("unexpected key %s", key)
	}
}

func TestFieldKeyFromPathSupportsPayloadRoot(t *testing.T) {
	if key := FieldKeyFromPath("$"); key != "payload" {
		t.Fatalf("expected payload root key, got %s", key)
	}
}

func TestParseJSONFieldsDoesNotDuplicateArrayPaths(t *testing.T) {
	nodes, err := ParseJSONFields([]byte(`{"items":[{"sku":"A"},{"sku":"B","price":10}]}`))
	if err != nil {
		t.Fatal(err)
	}
	paths := map[string]int{}
	var walk func([]JSONFieldNode)
	walk = func(items []JSONFieldNode) {
		for _, item := range items {
			paths[item.JSONPath]++
			walk(item.Children)
		}
	}
	walk(nodes)
	if paths["items[].sku"] != 1 || paths["items[].price"] != 1 {
		t.Fatalf("array paths were not deduplicated: %#v", paths)
	}
}

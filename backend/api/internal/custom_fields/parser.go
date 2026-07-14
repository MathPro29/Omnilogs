package custom_fields

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

const (
	MaxSampleJSONBytes = 5 << 20
	MaxSampleJSONNodes = 10000
)

// JSONFieldNode is a display-ready JSON tree node. Array children use [] in
// their path so a favorite remains stable when the array length changes.
type JSONFieldNode struct {
	FieldName   string          `json:"field_name"`
	JSONPath    string          `json:"json_path"`
	SampleValue any             `json:"sample_value"`
	DataType    string          `json:"data_type"`
	IsLeaf      bool            `json:"is_leaf"`
	Children    []JSONFieldNode `json:"children,omitempty"`
}

func ParseJSONFields(raw []byte) ([]JSONFieldNode, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, errors.New("sample_json is required")
	}
	if len(raw) > MaxSampleJSONBytes {
		return nil, fmt.Errorf("sample_json exceeds %d bytes", MaxSampleJSONBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("invalid sample_json: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, errors.New("sample_json must contain one JSON value")
		}
		return nil, fmt.Errorf("invalid trailing JSON: %w", err)
	}

	count := 0
	if object, ok := value.(map[string]any); ok {
		keys := sortedKeys(object)
		result := make([]JSONFieldNode, 0, len(keys))
		for _, key := range keys {
			result = append(result, buildNode(key, key, object[key], &count))
		}
		return result, nil
	}
	return []JSONFieldNode{buildNode("$", "$", value, &count)}, nil
}

func buildNode(name, path string, value any, count *int) JSONFieldNode {
	(*count)++
	node := JSONFieldNode{
		FieldName: name, JSONPath: path, SampleValue: value,
		DataType: detectDataType(value), IsLeaf: true,
	}
	if *count >= MaxSampleJSONNodes {
		return node
	}
	switch typed := value.(type) {
	case map[string]any:
		for _, key := range sortedKeys(typed) {
			node.Children = append(node.Children, buildNode(key, path+"."+key, typed[key], count))
		}
	case []any:
		node.Children = buildArrayChildren(path, typed, count)
	}
	node.IsLeaf = len(node.Children) == 0
	return node
}

func buildArrayChildren(path string, values []any, count *int) []JSONFieldNode {
	seen := map[string]bool{}
	var children []JSONFieldNode
	for _, value := range values {
		switch typed := value.(type) {
		case map[string]any:
			for _, key := range sortedKeys(typed) {
				childPath := path + "[]." + key
				if seen[childPath] {
					continue
				}
				seen[childPath] = true
				children = append(children, buildNode(key, childPath, typed[key], count))
			}
		default:
			childPath := path + "[]"
			if !seen[childPath] {
				seen[childPath] = true
				children = append(children, buildNode("[]", childPath, value, count))
			}
		}
		if *count >= MaxSampleJSONNodes {
			break
		}
	}
	return children
}

func sortedKeys(value map[string]any) []string {
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func detectDataType(value any) string {
	switch typed := value.(type) {
	case nil:
		return "string"
	case bool:
		return "boolean"
	case json.Number:
		return "number"
	case string:
		if isDateTime(typed) {
			return "datetime"
		}
		return "string"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	default:
		return "string"
	}
}

func isDateTime(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	for _, layout := range []string{time.RFC3339, time.RFC3339Nano, "2006-01-02", "2006-01-02 15:04:05"} {
		if _, err := time.Parse(layout, value); err == nil {
			return true
		}
	}
	return false
}

func FieldKeyFromPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "$" {
		return "payload"
	}
	parts := strings.FieldsFunc(path, func(r rune) bool { return r == '.' || r == '[' || r == ']' })
	key := "field"
	if len(parts) > 0 {
		key = parts[len(parts)-1]
	}
	var builder strings.Builder
	for _, r := range strings.ToLower(key) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			builder.WriteRune(r)
		} else {
			builder.WriteRune('_')
		}
	}
	result := strings.Trim(builder.String(), "_")
	if result == "" || (result[0] >= '0' && result[0] <= '9') {
		result = "field_" + result
	}
	return result
}

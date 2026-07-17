package usecase

import (
	"fmt"
	"strings"
)

func canonicalDynamicPath(path string) string {
	path = strings.Trim(strings.TrimSpace(path), ".")
	if path == "$" {
		return "$"
	}

	// Field definitions can originate from the legacy payload, the current data
	// shape, or a raw sample. The worker always traverses the decoded document
	// root, so these transport prefixes must not participate in path matching.
	for {
		switch {
		case strings.HasPrefix(path, "raw."):
			path = strings.TrimPrefix(path, "raw.")
		case strings.HasPrefix(path, "payload."):
			path = strings.TrimPrefix(path, "payload.")
		case strings.HasPrefix(path, "data."):
			path = strings.TrimPrefix(path, "data.")
		case strings.HasPrefix(path, "fields."):
			path = strings.TrimPrefix(path, "fields.")
		default:
			if path == "raw" || path == "payload" || path == "data" || path == "fields" || path == "" {
				return "$"
			}
			return path
		}
	}
}

func readJSONPath(value any, path string) (any, bool) {
	if strings.TrimSpace(path) == "$" {
		return value, true
	}
	parts := strings.Split(strings.Trim(path, "."), ".")
	if len(parts) == 0 || parts[0] == "" {
		return nil, false
	}
	return readJSONPathParts(value, parts)
}

func readJSONPathParts(value any, parts []string) (any, bool) {
	if len(parts) == 0 {
		return value, true
	}
	if values, ok := value.([]any); ok {
		result := make([]any, 0, len(values))
		for _, item := range values {
			if child, found := readJSONPathParts(item, parts); found {
				result = append(result, child)
			}
		}
		return result, len(result) > 0
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, false
	}
	segment := parts[0]
	if strings.HasSuffix(segment, "[]") {
		array, ok := object[strings.TrimSuffix(segment, "[]")].([]any)
		if !ok {
			return nil, false
		}
		if len(parts) == 1 {
			return array, true
		}
		return readJSONPathParts(array, parts[1:])
	}
	child, ok := object[segment]
	if !ok {
		return nil, false
	}
	return readJSONPathParts(child, parts[1:])
}

func writeJSONPath(object map[string]any, path string, value any) error {
	if strings.TrimSpace(path) == "$" {
		return nil
	}
	return writeJSONPathParts(object, strings.Split(strings.Trim(path, "."), "."), value)
}

func writeJSONPathParts(object map[string]any, parts []string, value any) error {
	if len(parts) == 0 {
		return nil
	}
	segment := parts[0]
	if strings.HasSuffix(segment, "[]") {
		key := strings.TrimSuffix(segment, "[]")
		array, ok := object[key].([]any)
		if !ok {
			return fmt.Errorf("JSON path %s is not an array", strings.Join(parts, "."))
		}
		values, _ := value.([]any)
		for index, item := range array {
			child, ok := item.(map[string]any)
			if !ok || index >= len(values) {
				continue
			}
			if err := writeJSONPathParts(child, parts[1:], values[index]); err != nil {
				return err
			}
		}
		return nil
	}
	if len(parts) == 1 {
		object[segment] = value
		return nil
	}
	child, ok := object[segment].(map[string]any)
	if !ok {
		return fmt.Errorf("JSON path %s is not an object", strings.Join(parts, "."))
	}
	return writeJSONPathParts(child, parts[1:], value)
}

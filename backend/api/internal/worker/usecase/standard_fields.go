package usecase

import (
	"net/url"
	"strings"
)

// normalizeHTTPFields promotes common connector HTTP shapes to the canonical
// fields used by routing rules, index metadata, and Log Explorer.
func normalizeHTTPFields(payload map[string]any) {
	normalizeActorField(payload)
	copyCanonicalString(payload, "request_method",
		[]string{"request_method"}, []string{"method"}, []string{"request", "method"},
		[]string{"http", "method"}, []string{"http", "request", "method"}, []string{"metadata", "action", "method"}, []string{"action", "method"})
	copyCanonicalString(payload, "route_pattern",
		[]string{"route_pattern"}, []string{"route"}, []string{"request", "route"},
		[]string{"http", "route"})
	copyCanonicalString(payload, "service",
		[]string{"service"}, []string{"service_name"}, []string{"source"},
		[]string{"resource", "service", "name"})

	if firstString(payload, "request_path") != "" {
		return
	}
	for _, path := range [][]string{
		{"path"}, {"endpoint"}, {"request", "path"}, {"http", "path"},
		{"url"}, {"request", "url"}, {"http", "url"},
	} {
		if value, ok := nestedString(payload, path...); ok {
			if parsed, err := url.Parse(value); err == nil && parsed.Path != "" {
				value = parsed.EscapedPath()
			}
			if value = strings.TrimSpace(value); value != "" {
				payload["request_path"] = value
				return
			}
		}
	}
	if route := firstString(payload, "route_pattern"); route != "" {
		payload["request_path"] = route
	}
}

func copyCanonicalString(payload map[string]any, target string, candidates ...[]string) {
	if firstString(payload, target) != "" {
		return
	}
	for _, path := range candidates {
		if value, ok := nestedString(payload, path...); ok {
			payload[target] = value
			return
		}
	}
}

func nestedString(payload map[string]any, path ...string) (string, bool) {
	current := any(payload)
	for _, segment := range path {
		object, ok := current.(map[string]any)
		if !ok {
			return "", false
		}
		current, ok = object[segment]
		if !ok {
			return "", false
		}
	}
	value, ok := current.(string)
	value = strings.TrimSpace(value)
	return value, ok && value != ""
}

func copyStandardFields(document, payload map[string]any) {
	aliases := map[string][]string{
		"level":             {"level", "log_level"},
		"message":           {"message"},
		"service":           {"service", "source"},
		"service_name":      {"service_name", "service", "source"},
		"source":            {"source", "service"},
		"route_key":         {"route_key"},
		"event_name":        {"event_name", "event_type"},
		"trace_id":          {"trace_id"},
		"request_id":        {"request_id", "source_request_id"},
		"method":            {"method", "request_method"},
		"path":              {"path", "request_path", "route_pattern"},
		"request_path":      {"request_path", "path", "endpoint", "route_pattern"},
		"endpoint":          {"endpoint", "request_path", "path", "route_pattern"},
		"status_code":       {"status_code", "response_status_code"},
		"duration_ms":       {"duration_ms"},
		"product_code":      {"product_code"},
		"environment_code":  {"environment_code"},
		"project_code":      {"project_code"},
		"category_code":     {"category_code"},
		"feature_code":      {"feature_code"},
		"sub_feature_code":  {"sub_feature_code"},
		"source_project_id": {"source_project_id"},
		"mapping_source":    {"mapping_source"},
		"mapping_status":    {"mapping_status"},
		"mapping_reason":    {"mapping_reason"},
	}
	for target, candidates := range aliases {
		for _, candidate := range candidates {
			if value, exists := payload[candidate]; exists && value != nil {
				document[target] = value
				break
			}
		}
	}

	if request, ok := payload["request"].(map[string]any); ok {
		copyIfMissing(document, "method", request["method"])
		copyIfMissing(document, "path", request["path"])
	}
	if response, ok := payload["response"].(map[string]any); ok {
		copyIfMissing(document, "status_code", response["status_code"])
	}
	for _, key := range []string{
		"actor", "metadata", "custom_fields",
		"source_project_id", "project_id", "category_id", "feature_id", "sub_feature_id", "feature_path_ids", "feature_full_path",
		"routing_status", "routing_method", "routing_reason",
	} {
		if value, exists := payload[key]; exists {
			document[key] = value
		}
	}
}

func copyIfMissing(target map[string]any, key string, value any) {
	if value == nil {
		return
	}
	if _, exists := target[key]; !exists {
		target[key] = value
	}
}

// normalizeActorField keeps the user who performed the source action in one
// stable shape for Log Explorer. The source service remains responsible for
// providing actor data; API-key ingestion cannot infer an end user.
func normalizeActorField(payload map[string]any) {
	raw, exists := payload["actor"]
	if !exists || raw == nil {
		return
	}
	actor, ok := raw.(map[string]any)
	if !ok {
		if id, exists := actorValue(raw); exists {
			payload["actor"] = map[string]any{"id": id}
		}
		return
	}
	canonical := make(map[string]any, 4)
	copyActorValue(canonical, "id", actor, "id", "user_id", "userId", "userID", "actor_id", "actorId")
	copyActorValue(canonical, "name", actor, "name", "username", "user_name", "userName", "full_name", "fullName", "display_name", "displayName")
	copyActorValue(canonical, "role", actor, "role", "role_name", "roleName", "user_role", "userRole")
	copyActorValue(canonical, "project_id", actor, "project_id", "projectId")
	if len(canonical) > 0 {
		payload["actor"] = canonical
	}
}

func copyActorValue(target map[string]any, targetKey string, actor map[string]any, keys ...string) {
	for _, key := range keys {
		if value, ok := actorValue(actor[key]); ok {
			target[targetKey] = value
			return
		}
	}
}

func actorValue(value any) (any, bool) {
	if value == nil {
		return nil, false
	}
	if text, ok := value.(string); ok {
		text = strings.TrimSpace(text)
		return text, text != ""
	}
	return value, true
}

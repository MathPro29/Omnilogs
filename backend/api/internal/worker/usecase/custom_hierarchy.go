package usecase

import (
	"strings"

	"omnilogs-api/models"
)

var customHierarchyAliases = map[string][]string{
	"product_id":       {"product_id", "productId"},
	"product_code":     {"product_code", "productCode", "product"},
	"environment_id":   {"environment_id", "environmentId"},
	"environment_code": {"environment_code", "environmentCode", "environment"},
	"project_id":       {"project_id", "projectId"},
	"project_code":     {"project_code", "projectCode", "project"},
	"category_id":      {"category_id", "categoryId"},
	"category_code":    {"category_code", "categoryCode", "category"},
	"feature_code":     {"feature_code", "featureCode", "feature"},
	"sub_feature_code": {"sub_feature_code", "subFeatureCode", "sub_feature", "subFeature"},
}

var sourceProjectIDAliases = []string{
	"source_project_id", "sourceProjectId", "projects_id", "projectsId",
}

var routingContextAliases = map[string][]string{
	"routing_type":  {"routing_type", "routingType", "type"},
	"routing_value": {"routing_value", "routingValue", "value"},
}

// normalizeCustomHierarchy promotes source-owned custom fields to the
// canonical fields consumed by validation and search.
func normalizeCustomHierarchy(payload map[string]any) bool {
	customFields := customFieldsMap(payload)
	candidates := make([]map[string]any, 0, 5)
	if routingContext, ok := payload["routing_context"].(map[string]any); ok {
		candidates = append(candidates, routingContext)
		promoteRoutingDimensions(payload, customFields, routingContext)
	}
	if routingContext, ok := payload["routingContext"].(map[string]any); ok {
		candidates = append(candidates, routingContext)
		promoteRoutingDimensions(payload, customFields, routingContext)
	}
	if customFields != nil {
		candidates = append(candidates, customFields)
	}
	if routing, ok := payload["routing"].(map[string]any); ok {
		if explicit, ok := routing["explicit_hierarchy"].(map[string]any); ok {
			candidates = append(candidates, explicit)
		}
		if explicit, ok := routing["explicitHierarchy"].(map[string]any); ok {
			candidates = append(candidates, explicit)
		}
	}
	if actor, ok := payload["actor"].(map[string]any); ok {
		candidates = append(candidates, actor)
	}
	if metadata, ok := payload["metadata"].(map[string]any); ok {
		candidates = append(candidates, metadata)
		if actor, ok := metadata["actor"].(map[string]any); ok {
			candidates = append(candidates, actor)
		}
	}
	// Accept the old top-level contract during migration.
	candidates = append(candidates, payload)

	hasHierarchy := false
	// Preserve the source-system project ID independently from the OmniLogs hierarchy ID.
	if value, found := firstHierarchyValue(candidates, sourceProjectIDAliases...); found && !emptyHierarchyValue(value) {
		payload["source_project_id"] = value
	}
	for canonical, aliases := range customHierarchyAliases {
		value, found := firstHierarchyValue(candidates, aliases...)
		if !found || emptyHierarchyValue(value) {
			continue
		}
		if customFields == nil {
			customFields = map[string]any{}
			payload["custom_fields"] = customFields
		}
		customFields[canonical] = value
		payload[canonical] = value
		if canonical == "project_id" {
			if _, exists := payload["source_project_id"]; !exists {
				payload["source_project_id"] = value
			}
		}
		hasHierarchy = true
	}
	return hasHierarchy
}

func promoteRoutingDimensions(payload map[string]any, customFields map[string]any, routingContext map[string]any) {
	for canonical, aliases := range routingContextAliases {
		value, found := firstHierarchyValue([]map[string]any{routingContext}, aliases...)
		if !found || emptyHierarchyValue(value) {
			continue
		}
		payload[canonical] = value
		if customFields != nil {
			customFields[canonical] = value
		}
	}
}

func customFieldsMap(payload map[string]any) map[string]any {
	for _, key := range []string{"custom_fields", "customFields"} {
		if value, ok := payload[key].(map[string]any); ok {
			if key != "custom_fields" {
				payload["custom_fields"] = value
			}
			return value
		}
	}
	return nil
}

func firstHierarchyValue(candidates []map[string]any, aliases ...string) (any, bool) {
	for _, candidate := range candidates {
		if candidate == nil {
			continue
		}
		for _, alias := range aliases {
			if value, ok := candidate[alias]; ok && !emptyHierarchyValue(value) {
				return value, true
			}
		}
	}
	return nil, false
}

func emptyHierarchyValue(value any) bool {
	if value == nil {
		return true
	}
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text) == ""
	}
	return false
}

func hasCustomHierarchy(payload map[string]any) bool {
	customFields := customFieldsMap(payload)
	if customFields == nil {
		return false
	}
	for _, aliases := range customHierarchyAliases {
		if _, found := firstHierarchyValue([]map[string]any{customFields}, aliases...); found {
			return true
		}
	}
	return false
}

func setCustomHierarchyValue(payload map[string]any, key string, value any) {
	if emptyHierarchyValue(value) {
		return
	}
	customFields := customFieldsMap(payload)
	if customFields == nil {
		customFields = map[string]any{}
		payload["custom_fields"] = customFields
	}
	customFields[key] = value
}

// applyConfiguredCustomFields bridges source-owned key:value pairs and the
// Product's custom-field definitions. A source value is promoted only when its
// key matches the configured destination key (or value_source_key). Missing
// matches receive the configured default for consistent Log Explorer rows.
func applyConfiguredCustomFields(payload map[string]any, fields []models.LogFieldDefinition) {
	for _, field := range fields {
		target := strings.TrimSpace(field.FieldKey)
		if target == "" {
			continue
		}
		if customFields := customFieldsMap(payload); customFields != nil {
			if value, exists := customFields[target]; exists && !emptyHierarchyValue(value) {
				continue
			}
		}
		matched := false
		for _, sourceKey := range configuredFieldSourceKeys(field) {
			if value, exists := readConfiguredFieldValue(payload, sourceKey); exists && !emptyHierarchyValue(value) {
				setCustomHierarchyValue(payload, target, value)
				matched = true
				break
			}
		}
		if !matched && field.DefaultValue != nil && strings.TrimSpace(*field.DefaultValue) != "" {
			setCustomHierarchyValue(payload, target, *field.DefaultValue)
		}
	}
}

func configuredFieldSourceKeys(field models.LogFieldDefinition) []string {
	keys := []string{strings.TrimSpace(field.FieldKey)}
	if field.ValueSourceKey != nil {
		if key := strings.TrimSpace(*field.ValueSourceKey); key != "" {
			keys = append(keys, key)
		}
	}
	return keys
}

func readConfiguredFieldValue(payload map[string]any, key string) (any, bool) {
	if customFields := customFieldsMap(payload); customFields != nil {
		if value, exists := customFields[key]; exists {
			return value, true
		}
	}
	value, exists := payload[key]
	return value, exists
}

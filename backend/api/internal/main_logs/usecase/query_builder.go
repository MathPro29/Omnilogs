package usecase

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var elasticFieldSegmentPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

const maxFilterIDs = 100

func normalizeSearchInput(input SearchInput) SearchInput {
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.PerPage <= 0 {
		input.PerPage = 20
	} else if input.PerPage > 100 {
		input.PerPage = 100
	}
	input.ProjectIDs = normalizeFilterIDs(input.ProjectIDs)
	input.CategoryIDs = normalizeFilterIDs(input.CategoryIDs)
	return input
}

func normalizeFilterIDs(values []int64) []int64 {
	result := make([]int64, 0, min(len(values), maxFilterIDs))
	seen := make(map[int64]struct{}, min(len(values), maxFilterIDs))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
		if len(result) == maxFilterIDs {
			break
		}
	}
	return result
}

func normalizeCustomFieldPath(raw string) (string, bool) {
	path := strings.TrimSpace(raw)
	path = strings.TrimPrefix(path, "payload.")
	path = strings.ReplaceAll(path, "[]", "")
	segments := strings.Split(path, ".")
	if len(segments) == 0 || len(segments) > 20 {
		return "", false
	}
	for _, segment := range segments {
		if !elasticFieldSegmentPattern.MatchString(segment) {
			return "", false
		}
	}
	return "payload." + strings.Join(segments, "."), true
}

func validateSearchInput(input SearchInput) error {
	hasPath := input.CustomFieldPath != nil && strings.TrimSpace(*input.CustomFieldPath) != ""
	hasValue := input.CustomFieldValue != nil && strings.TrimSpace(*input.CustomFieldValue) != ""
	if hasPath != hasValue {
		return fmt.Errorf("%w: custom_field_path and custom_field_value must be provided together", ErrInvalidSearchFilter)
	}
	if hasPath {
		if _, valid := normalizeCustomFieldPath(*input.CustomFieldPath); !valid {
			return fmt.Errorf("%w: custom_field_path is not allowed", ErrInvalidSearchFilter)
		}
	}
	return nil
}

func buildSearchQuery(input SearchInput) map[string]any {
	input = normalizeSearchInput(input)
	filters := []map[string]any{
		{"term": map[string]any{"product_id": input.ProductID}},
	}
	if input.EnvironmentID != nil {
		filters = append(filters, map[string]any{"term": map[string]any{"environment_id": *input.EnvironmentID}})
	}
	if input.ProjectID != nil {
		filters = append(filters, map[string]any{"term": map[string]any{"payload.project_id": *input.ProjectID}})
	}
	if input.CategoryID != nil {
		categoryID := fmt.Sprintf("%d", *input.CategoryID)
		filters = append(filters, map[string]any{
			"bool": map[string]any{
				"should": []map[string]any{
					{"term": map[string]any{"payload.category_id": *input.CategoryID}},
					{"term": map[string]any{"payload.feature_path_ids.keyword": categoryID}},
					{"wildcard": map[string]any{"payload.feature_path_ids.keyword": categoryID + ",*"}},
					{"wildcard": map[string]any{"payload.feature_path_ids.keyword": "*," + categoryID}},
					{"wildcard": map[string]any{"payload.feature_path_ids.keyword": "*," + categoryID + ",*"}},
				},
				"minimum_should_match": 1,
			},
		})
	}
	// Multi-project filter
	if len(input.ProjectIDs) > 0 && input.ProjectID == nil {
		filters = append(filters, map[string]any{"terms": map[string]any{"payload.project_id": input.ProjectIDs}})
	}
	// Multi-category filter
	if len(input.CategoryIDs) > 0 && input.CategoryID == nil {
		shouldClauses := make([]map[string]any, 0)
		for _, catID := range input.CategoryIDs {
			catStr := fmt.Sprintf("%d", catID)
			shouldClauses = append(shouldClauses,
				map[string]any{"term": map[string]any{"payload.category_id": catID}},
				map[string]any{"term": map[string]any{"payload.feature_path_ids.keyword": catStr}},
				map[string]any{"wildcard": map[string]any{"payload.feature_path_ids.keyword": catStr + ",*"}},
				map[string]any{"wildcard": map[string]any{"payload.feature_path_ids.keyword": "*," + catStr}},
				map[string]any{"wildcard": map[string]any{"payload.feature_path_ids.keyword": "*," + catStr + ",*"}},
			)
		}
		filters = append(filters, map[string]any{
			"bool": map[string]any{
				"should":               shouldClauses,
				"minimum_should_match": 1,
			},
		})
	}
	if input.Level != nil && strings.TrimSpace(*input.Level) != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"payload.log_level.keyword": strings.TrimSpace(*input.Level)}})
	}
	if input.LogType != nil && strings.TrimSpace(*input.LogType) != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"payload.event_type.keyword": strings.TrimSpace(*input.LogType)}})
	}
	if input.RequestID != nil && strings.TrimSpace(*input.RequestID) != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"payload.source_request_id.keyword": strings.TrimSpace(*input.RequestID)}})
	}
	if input.TraceID != nil && strings.TrimSpace(*input.TraceID) != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"payload.trace_id.keyword": strings.TrimSpace(*input.TraceID)}})
	}
	if input.CustomFieldPath != nil && input.CustomFieldValue != nil && strings.TrimSpace(*input.CustomFieldPath) != "" {
		field, valid := normalizeCustomFieldPath(*input.CustomFieldPath)
		value := strings.TrimSpace(*input.CustomFieldValue)
		if valid && value != "" {
			shouldClauses := []map[string]any{
				{"term": map[string]any{field + ".keyword": value}},
				{"term": map[string]any{field: value}},
				{"match_phrase": map[string]any{field: value}},
			}

			// Try parsing as boolean
			if strings.EqualFold(value, "true") {
				shouldClauses = append(shouldClauses, map[string]any{"term": map[string]any{field: true}})
			} else if strings.EqualFold(value, "false") {
				shouldClauses = append(shouldClauses, map[string]any{"term": map[string]any{field: false}})
			}

			// Try parsing as numeric
			if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
				shouldClauses = append(shouldClauses, map[string]any{"term": map[string]any{field: intVal}})
			} else if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
				shouldClauses = append(shouldClauses, map[string]any{"term": map[string]any{field: floatVal}})
			}

			filters = append(filters, map[string]any{"bool": map[string]any{
				"should":               shouldClauses,
				"minimum_should_match": 1,
			}})
		}
	}

	must := []map[string]any{}
	if input.Keyword != nil && strings.TrimSpace(*input.Keyword) != "" {
		keyword := strings.TrimSpace(*input.Keyword)
		must = append(must, map[string]any{
			"bool": map[string]any{
				"should": []map[string]any{
					{
						"multi_match": map[string]any{
							"query": keyword,
							"type":  "best_fields",
							"fields": []string{
								"payload.message^4",
								"payload.error_message^3",
								"payload.stack_trace^2",
								"payload.request_path^3",
								"payload.route_pattern^2",
								"payload.url^2",
								"payload.event_type^2",
								"payload.source_request_id^3",
								"payload.trace_id^3",
								"payload.correlation_id^2",
								"payload.feature_full_path^2",
							},
							"operator": "and",
						},
					},
					{
						"multi_match": map[string]any{
							"query": keyword,
							"type":  "phrase_prefix",
							"fields": []string{
								"payload.message^5",
								"payload.request_path^3",
								"payload.route_pattern^2",
								"payload.url^2",
								"payload.event_type^2",
								"payload.source_request_id^4",
								"payload.trace_id^4",
								"payload.feature_full_path^2",
							},
						},
					},
					{
						"simple_query_string": map[string]any{
							"query": keyword + "*",
							"fields": []string{
								"payload.message^4",
								"payload.error_message^3",
								"payload.stack_trace^2",
								"payload.request_path^3",
								"payload.route_pattern^2",
								"payload.url^2",
								"payload.event_type^2",
								"payload.source_request_id^3",
								"payload.trace_id^3",
								"payload.correlation_id^2",
								"payload.feature_full_path^2",
							},
							"default_operator": "and",
						},
					},
				},
				"minimum_should_match": 1,
			},
		})
	}

	return buildSearchQueryBody(input, filters, must)
}

func buildSearchQueryBody(input SearchInput, filters, must []map[string]any) map[string]any {
	return map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"filter": filters,
				"must":   must,
			},
		},
		"sort": []any{
			map[string]any{"@timestamp": map[string]any{"order": "desc"}},
		},
		"from": (input.Page - 1) * input.PerPage,
		"size": input.PerPage,
	}
}

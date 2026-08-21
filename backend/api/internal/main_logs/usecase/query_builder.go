package usecase

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
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
	path = strings.TrimPrefix(path, "raw.")
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
	normalized := strings.Join(segments, ".")
	if normalized == "payload" || normalized == "data" {
		return "data", true
	}
	// payload is retained in _source; data is the sole searchable raw document.
	normalized = strings.TrimPrefix(normalized, "payload.")
	normalized = strings.TrimPrefix(normalized, "data.")
	return "data." + normalized, true
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
	if input.SortField != nil {
		if _, valid := normalizeCustomFieldPath(*input.SortField); !valid {
			return fmt.Errorf("%w: sort_field is not allowed", ErrInvalidSearchFilter)
		}
	}
	if input.SortOrder != "" && !strings.EqualFold(input.SortOrder, "asc") && !strings.EqualFold(input.SortOrder, "desc") {
		return fmt.Errorf("%w: sort_order must be asc or desc", ErrInvalidSearchFilter)
	}
	return nil
}

func buildSearchQuery(input SearchInput) map[string]any {
	input = normalizeSearchInput(input)
	filters := []map[string]any{
		{"term": map[string]any{"product_id": input.ProductID}},
	}

	// Retained fields from Omnilogs (หลัก)
	if input.DateFrom != nil {
		filters = append(filters, map[string]any{"range": map[string]any{"@timestamp": map[string]any{"gte": input.DateFrom.UTC().Format(time.RFC3339Nano)}}})
	}
	if input.StatusCode != nil {
		filters = append(filters, map[string]any{"term": map[string]any{"payload.status_code": *input.StatusCode}})
	}
	if input.SourceID != nil {
		filters = append(filters, map[string]any{"term": map[string]any{"source_id": *input.SourceID}})
	}

	if input.EnvironmentID != nil {
		filters = append(filters, map[string]any{"term": map[string]any{"environment_id": *input.EnvironmentID}})
	}
	if input.ProjectID != nil {
		projectID := *input.ProjectID
		filters = append(filters, map[string]any{"bool": map[string]any{
			"should": []map[string]any{
				{"term": map[string]any{"project_id": projectID}},
				{"term": map[string]any{"project_id": fmt.Sprintf("%d", projectID)}},
				{"term": map[string]any{"payload.project_id": projectID}},
				{"term": map[string]any{"data.project_id": projectID}},
				{"term": map[string]any{"actor.project_id": projectID}},
				{"term": map[string]any{"data.actor.project_id": projectID}},
				{"term": map[string]any{"payload.actor.project_id": projectID}},
				{"term": map[string]any{"metadata.actor.project_id": projectID}},
				{"term": map[string]any{"payload.metadata.actor.project_id": projectID}},
				{"term": map[string]any{"data.metadata.actor.project_id": projectID}},
				{"term": map[string]any{"payload.metadata.project_id": projectID}},
				{"term": map[string]any{"data.metadata.project_id": projectID}},
				{"term": map[string]any{"custom_fields.project_id": projectID}},
				{"term": map[string]any{"data.custom_fields.project_id": projectID}},
			},
			"minimum_should_match": 1,
		}})
	}
	if input.CategoryID != nil {
		categoryID := fmt.Sprintf("%d", *input.CategoryID)
		filters = append(filters, map[string]any{
			"bool": map[string]any{
				"should": []map[string]any{
					{"term": map[string]any{"category_id": *input.CategoryID}},
					{"term": map[string]any{"payload.category_id": *input.CategoryID}},
					{"term": map[string]any{"custom_fields.category_id": *input.CategoryID}},
					{"term": map[string]any{"data.custom_fields.category_id": *input.CategoryID}},
					{"term": map[string]any{"feature_path_ids": categoryID}},
					{"term": map[string]any{"payload.feature_path_ids.keyword": categoryID}},
					{"wildcard": map[string]any{"feature_path_ids": categoryID + ",*"}},
					{"wildcard": map[string]any{"payload.feature_path_ids.keyword": categoryID + ",*"}},
					{"wildcard": map[string]any{"feature_path_ids": "*," + categoryID}},
					{"wildcard": map[string]any{"payload.feature_path_ids.keyword": "*," + categoryID}},
					{"wildcard": map[string]any{"feature_path_ids": "*," + categoryID + ",*"}},
					{"wildcard": map[string]any{"payload.feature_path_ids.keyword": "*," + categoryID + ",*"}},
				},
				"minimum_should_match": 1,
			},
		})
	}
	// Multi-project filter
	if len(input.ProjectIDs) > 0 && input.ProjectID == nil {
		filters = append(filters, map[string]any{"bool": map[string]any{
			"should": []map[string]any{
				{"terms": map[string]any{"project_id": input.ProjectIDs}},
				{"terms": map[string]any{"payload.project_id": input.ProjectIDs}},
				{"terms": map[string]any{"data.project_id": input.ProjectIDs}},
				{"terms": map[string]any{"actor.project_id": input.ProjectIDs}},
				{"terms": map[string]any{"data.actor.project_id": input.ProjectIDs}},
				{"terms": map[string]any{"payload.actor.project_id": input.ProjectIDs}},
				{"terms": map[string]any{"metadata.actor.project_id": input.ProjectIDs}},
				{"terms": map[string]any{"payload.metadata.actor.project_id": input.ProjectIDs}},
				{"terms": map[string]any{"data.metadata.actor.project_id": input.ProjectIDs}},
				{"terms": map[string]any{"payload.metadata.project_id": input.ProjectIDs}},
				{"terms": map[string]any{"data.metadata.project_id": input.ProjectIDs}},
				{"terms": map[string]any{"custom_fields.project_id": input.ProjectIDs}},
				{"terms": map[string]any{"data.custom_fields.project_id": input.ProjectIDs}},
			},
			"minimum_should_match": 1,
		}})
	}
	// Multi-category filter
	if len(input.CategoryIDs) > 0 && input.CategoryID == nil {
		shouldClauses := make([]map[string]any, 0)
		for _, catID := range input.CategoryIDs {
			catStr := fmt.Sprintf("%d", catID)
			shouldClauses = append(shouldClauses,
				map[string]any{"term": map[string]any{"category_id": catID}},
				map[string]any{"term": map[string]any{"payload.category_id": catID}},
				map[string]any{"term": map[string]any{"custom_fields.category_id": catID}},
				map[string]any{"term": map[string]any{"data.custom_fields.category_id": catID}},
				map[string]any{"term": map[string]any{"feature_path_ids": catStr}},
				map[string]any{"term": map[string]any{"payload.feature_path_ids.keyword": catStr}},
				map[string]any{"wildcard": map[string]any{"feature_path_ids": catStr + ",*"}},
				map[string]any{"wildcard": map[string]any{"payload.feature_path_ids.keyword": catStr + ",*"}},
				map[string]any{"wildcard": map[string]any{"feature_path_ids": "*," + catStr}},
				map[string]any{"wildcard": map[string]any{"payload.feature_path_ids.keyword": "*," + catStr}},
				map[string]any{"wildcard": map[string]any{"feature_path_ids": "*," + catStr + ",*"}},
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
		must = append(must, buildGlobalSearchClause(strings.TrimSpace(*input.Keyword)))
	}

	return buildSearchQueryBody(input, filters, must)
}

func buildSort(input SearchInput) []any {
	if input.SortField == nil {
		return []any{map[string]any{"@timestamp": map[string]any{"order": "desc"}}}
	}
	field, valid := normalizeCustomFieldPath(*input.SortField)
	if !valid {
		return []any{map[string]any{"@timestamp": map[string]any{"order": "desc"}}}
	}
	order := strings.ToLower(input.SortOrder)
	if order == "" {
		order = "asc"
	}
	return []any{map[string]any{field: map[string]any{"order": order, "unmapped_type": "keyword", "missing": "_last"}}}
}

func buildSearchQueryBody(input SearchInput, filters, must []map[string]any) map[string]any {
	return map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"filter": filters,
				"must":   must,
			},
		},
		"sort": buildSort(input),
		"from": (input.Page - 1) * input.PerPage,
		"size": input.PerPage,
	}
}

package usecase

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func buildDynamicSearchQuery(input DynamicSearchInput) map[string]any {
	// Normalize pagination
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	} else if input.PageSize > 100 {
		input.PageSize = 100
	}

	filters := []map[string]any{
		{"term": map[string]any{"product_id": input.Scope.ProductID}},
	}

	if input.Scope.EnvironmentID != nil {
		filters = append(filters, map[string]any{"term": map[string]any{"environment_id": *input.Scope.EnvironmentID}})
	}

	if input.Scope.ProjectID != nil {
		projectID := *input.Scope.ProjectID
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

	if input.Scope.CategoryID != nil {
		categoryID := fmt.Sprintf("%d", *input.Scope.CategoryID)
		filters = append(filters, map[string]any{
			"bool": map[string]any{
				"should": []map[string]any{
					{"term": map[string]any{"category_id": *input.Scope.CategoryID}},
					{"term": map[string]any{"payload.category_id": *input.Scope.CategoryID}},
					{"term": map[string]any{"custom_fields.category_id": *input.Scope.CategoryID}},
					{"term": map[string]any{"data.custom_fields.category_id": *input.Scope.CategoryID}},
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

	// Handle Time Range
	if input.TimeRange.Value != "all" && input.TimeRange.Value != "" {
		val := input.TimeRange.Value
		var dateFrom time.Time
		now := time.Now()

		// Parse standard relative times like "24h", "7d", "30m"
		if strings.HasSuffix(val, "m") {
			if num, err := strconv.Atoi(strings.TrimSuffix(val, "m")); err == nil {
				dateFrom = now.Add(-time.Duration(num) * time.Minute)
			}
		} else if strings.HasSuffix(val, "h") {
			if num, err := strconv.Atoi(strings.TrimSuffix(val, "h")); err == nil {
				dateFrom = now.Add(-time.Duration(num) * time.Hour)
			}
		} else if strings.HasSuffix(val, "d") {
			if num, err := strconv.Atoi(strings.TrimSuffix(val, "d")); err == nil {
				dateFrom = now.AddDate(0, 0, -num)
			}
		}

		if !dateFrom.IsZero() {
			field := input.TimeRange.Field
			if field == "" {
				field = "@timestamp"
			}
			filters = append(filters, map[string]any{"range": map[string]any{field: map[string]any{"gte": dateFrom.UTC().Format(time.RFC3339Nano)}}})
		}
	}

	must := []map[string]any{}

	// Process dynamic children
	if input.Root.Type == "group" && len(input.Root.Children) > 0 {
		var groupClauses []map[string]any

		for _, child := range input.Root.Children {
			if !child.Enabled || child.Field == "" {
				continue
			}

			field := child.Field
			field = strings.TrimPrefix(field, "raw.")

			field = strings.ReplaceAll(field, "[]", "")

			var searchFields []string
			if field == "global_search" {
				searchFields = []string{"global_search"}
			} else if field == "log_id" || field == "id" {
				searchFields = []string{"_id"}
			} else if field == "product_id" || field == "environment_id" || field == "timestamp" || field == "@timestamp" {
				searchFields = []string{field}
			} else if strings.HasPrefix(field, "payload.") {
				searchFields = []string{"data." + strings.TrimPrefix(field, "payload.")}
			} else if strings.HasPrefix(field, "custom_fields.") {
				searchFields = []string{"data." + field, field}
			} else if strings.HasPrefix(field, "data.") || strings.HasPrefix(field, "metadata.") {
				searchFields = []string{field}
			} else {
				// Unqualified custom fields or virtual UI fields: search across possible locations
				searchFields = []string{
					"data." + field,
					field,
				}
				// Also handle mapping request.path to request_path
				if strings.HasPrefix(field, "request.") || strings.HasPrefix(field, "response.") || strings.HasPrefix(field, "headers.") {
					mapped := strings.Replace(field, ".", "_", 1)
					searchFields = append(searchFields, "data."+mapped)
				}
			}

			var fieldClauses []map[string]any
			for _, sf := range searchFields {
				clause := buildConditionClause(sf, child.Operator, child.Value)
				if clause != nil {
					fieldClauses = append(fieldClauses, clause)
				}
			}

			if len(fieldClauses) == 1 {
				groupClauses = append(groupClauses, fieldClauses[0])
			} else if len(fieldClauses) > 1 {
				// OR logic across the possible fields
				groupClauses = append(groupClauses, map[string]any{
					"bool": map[string]any{
						"should":               fieldClauses,
						"minimum_should_match": 1,
					},
				})
			}
		}

		if len(groupClauses) > 0 {
			if strings.EqualFold(input.Root.Operator, "OR") {
				must = append(must, map[string]any{"bool": map[string]any{"should": groupClauses, "minimum_should_match": 1}})
			} else {
				must = append(must, groupClauses...)
			}
		}
	}

	// Sort
	sortItems := []any{}
	if input.SortField != nil && *input.SortField != "" {
		field := *input.SortField
		if field == "log_id" || field == "id" {
			field = "_id"
		} else if !strings.HasPrefix(field, "payload.") && !strings.HasPrefix(field, "data.") && field != "@timestamp" && field != "timestamp" {
			field = "data." + field
		}

		order := strings.ToLower(input.SortOrder)
		if order == "" {
			order = "asc"
		}
		sortItems = append(sortItems, map[string]any{field: map[string]any{"order": order, "unmapped_type": "keyword", "missing": "_last"}})
	}
	// Default sort
	sortItems = append(sortItems, map[string]any{"@timestamp": map[string]any{"order": "desc"}})

	return map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"filter": filters,
				"must":   must,
			},
		},
		"sort":             sortItems,
		"from":             (input.Page - 1) * input.PageSize,
		"size":             input.PageSize,
		"track_total_hits": true,
		"_source": map[string]any{
			"includes": []string{
				"@timestamp", "product_id", "environment_id", "source_id", "source_project_id",
				"project_id", "category_id", "feature_path_ids", "feature_full_path",
				"routing_status", "routing_method",
				"payload.*", "data.*",
			},
		},
	}
}

func buildConditionClause(field, operator string, value any) map[string]any {
	if field == "global_search" {
		return buildGlobalSearchClause(fmt.Sprint(value))
	}

	// Root metadata fields like _id, product_id, environment_id don't have .keyword subfields
	keywordField := field
	if field != "_id" && field != "product_id" && field != "environment_id" && field != "timestamp" && field != "@timestamp" {
		keywordField = field + ".keyword"
	}

	switch operator {
	case "eq":
		return map[string]any{"bool": map[string]any{
			"should": []map[string]any{
				{"term": map[string]any{keywordField: value}},
				{"term": map[string]any{field: value}},
				{"match_phrase": map[string]any{field: value}},
			},
			"minimum_should_match": 1,
		}}
	case "neq":
		return map[string]any{"bool": map[string]any{
			"must_not": []map[string]any{
				{"term": map[string]any{keywordField: value}},
				{"term": map[string]any{field: value}},
				{"match_phrase": map[string]any{field: value}},
			},
		}}
	case "contains":
		strVal := fmt.Sprintf("%v", value)
		return map[string]any{"bool": map[string]any{
			"should": []map[string]any{
				{"match_phrase": map[string]any{field: strVal}},
				{"wildcard": map[string]any{keywordField: "*" + strVal + "*"}},
			},
			"minimum_should_match": 1,
		}}
	case "exists":
		return map[string]any{"exists": map[string]any{"field": field}}
	case "not_exists":
		return map[string]any{"bool": map[string]any{"must_not": []map[string]any{{"exists": map[string]any{"field": field}}}}}
	case "gt", "gte", "lt", "lte":
		return map[string]any{"range": map[string]any{field: map[string]any{operator: value}}}
	}
	return nil
}

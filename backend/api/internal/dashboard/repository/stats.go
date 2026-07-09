package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"omnilogs-api/dto"
)

func (r *repository) GetLogStats(ctx context.Context, query dto.LogQuery) (map[string]any, error) {
	mustQueries := []map[string]any{
		{
			"term": map[string]any{
				"product_id": query.ProductID,
			},
		},
	}
	if query.EnvironmentID > 0 {
		mustQueries = append(mustQueries, map[string]any{
			"term": map[string]any{
				"environment_id": query.EnvironmentID,
			},
		})
	}

	// Multi-project filter for stats
	projectIDs := parseCSVInt64(query.ProjectIDs)
	if len(projectIDs) > 0 {
		mustQueries = append(mustQueries, map[string]any{
			"terms": map[string]any{
				"payload.project_id": projectIDs,
			},
		})
	}

	// Multi-category filter for stats
	categoryIDs := parseCSVInt64(query.CategoryIDs)
	if len(categoryIDs) > 0 {
		shouldClauses := make([]map[string]any, 0)
		for _, catID := range categoryIDs {
			catStr := fmt.Sprintf("%d", catID)
			shouldClauses = append(shouldClauses,
				map[string]any{"term": map[string]any{"payload.category_id": catID}},
				map[string]any{"term": map[string]any{"payload.feature_path_ids.keyword": catStr}},
				map[string]any{"wildcard": map[string]any{"payload.feature_path_ids.keyword": catStr + ",*"}},
				map[string]any{"wildcard": map[string]any{"payload.feature_path_ids.keyword": "*," + catStr}},
				map[string]any{"wildcard": map[string]any{"payload.feature_path_ids.keyword": "*," + catStr + ",*"}},
			)
		}
		mustQueries = append(mustQueries, map[string]any{
			"bool": map[string]any{
				"should":               shouldClauses,
				"minimum_should_match": 1,
			},
		})
	}

	if query.Search != "" {
		mustQueries = append(mustQueries, map[string]any{
			"multi_match": map[string]any{
				"query":  query.Search,
				"fields": []string{"payload.*"},
			},
		})
	}

	rangeQuery := map[string]any{}
	if query.StartTime != "" {
		rangeQuery["gte"] = query.StartTime
	}
	if query.EndTime != "" {
		rangeQuery["lte"] = query.EndTime
	}
	if len(rangeQuery) > 0 {
		mustQueries = append(mustQueries, map[string]any{
			"range": map[string]any{
				"@timestamp": rangeQuery,
			},
		})
	}

	esQuery := map[string]any{
		"size": 0,
		"query": map[string]any{
			"bool": map[string]any{
				"must": mustQueries,
			},
		},
		"aggs": map[string]any{
			"log_levels": map[string]any{
				"terms": map[string]any{
					"field": "payload.log_level.keyword",
					"size":  10,
				},
			},
			"logs_over_time": map[string]any{
				"date_histogram": map[string]any{
					"field":             "@timestamp",
					"calendar_interval": "hour",
					"min_doc_count":     0,
				},
				"aggs": map[string]any{
					"by_level": map[string]any{
						"terms": map[string]any{
							"field": "payload.log_level.keyword",
							"size":  10,
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(esQuery); err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	indices := r.resolveSearchIndices(query.ProductID)
	res, err := r.esClient.Search(
		r.esClient.Search.WithContext(ctx),
		r.esClient.Search.WithIndex(indices...),
		r.esClient.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch error: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch failed: %s", res.String())
	}

	var searchResult map[string]any
	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to parse search result: %w", err)
	}

	return searchResult, nil
}

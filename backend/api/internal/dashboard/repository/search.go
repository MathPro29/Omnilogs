package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"omnilogs-api/dto"
)

// buildElasticsearchQuery สร้าง dynamic query สำหรับส่งให้ Elasticsearch
func buildElasticsearchQuery(query dto.LogQuery) map[string]any {
	filters := []map[string]any{
		{
			"term": map[string]any{
				"product_id": query.ProductID,
			},
		},
	}

	if query.EnvironmentID > 0 {
		filters = append(filters, map[string]any{
			"term": map[string]any{
				"environment_id": query.EnvironmentID,
			},
		})
	}

	if query.LogLevel != "" {
		filters = append(filters, map[string]any{
			"term": map[string]any{
				"payload.log_level.keyword": query.LogLevel, // ใช้ payload.log_level หรือฟิลด์อื่นๆ ตามที่บันทึก
			},
		})
	}

	// Multi-project filter
	projectIDs := parseCSVInt64(query.ProjectIDs)
	if len(projectIDs) > 0 {
		filters = append(filters, map[string]any{
			"terms": map[string]any{
				"payload.project_id": projectIDs,
			},
		})
	}

	// Multi-category filter
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
		filters = append(filters, map[string]any{
			"bool": map[string]any{
				"should":               shouldClauses,
				"minimum_should_match": 1,
			},
		})
	}

	mustQueries := make([]map[string]any, 0, 1)
	if query.Search != "" {
		mustQueries = append(mustQueries, map[string]any{
			"multi_match": map[string]any{
				"query":  query.Search,
				"fields": []string{"payload.*"}, // ค้นหาในทุกฟิลด์ของ payload
			},
		})
	}

	// กรองช่วงเวลา (Timestamp)
	rangeQuery := map[string]any{}
	if query.StartTime != "" {
		rangeQuery["gte"] = query.StartTime
	}
	if query.EndTime != "" {
		rangeQuery["lte"] = query.EndTime
	}

	if len(rangeQuery) > 0 {
		filters = append(filters, map[string]any{
			"range": map[string]any{
				"@timestamp": rangeQuery,
			},
		})
	}

	return map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"filter": filters,
				"must":   mustQueries,
			},
		},
		"sort": []any{
			map[string]any{
				"@timestamp": map[string]any{
					"order": "desc", // เอาข้อมูลล่าสุดขึ้นก่อน
				},
			},
		},
		"from": query.Offset,
		"size": query.Limit,
	}
}

func (r *repository) SearchLogs(ctx context.Context, query dto.LogQuery) (map[string]any, error) {
	esQuery := buildElasticsearchQuery(query)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(esQuery); err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	indices := r.resolveSearchIndices(ctx, query.ProductID, query.EnvironmentID)
	startedAt := time.Now()
	defer r.logSlowElasticsearch("dashboard_search", startedAt, indices, query.Limit)
	res, err := r.esClient.Search(
		r.esClient.Search.WithContext(ctx),
		r.esClient.Search.WithIndex(indices...),
		r.esClient.Search.WithBody(&buf),
		r.esClient.Search.WithTrackTotalHits(true),
		r.esClient.Search.WithTimeout(r.queryTimeout),
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
	if timedOut, _ := searchResult["timed_out"].(bool); timedOut {
		return nil, context.DeadlineExceeded
	}

	return searchResult, nil
}

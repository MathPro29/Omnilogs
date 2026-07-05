package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"omnilogs-api/dto"
	"omnilogs-api/models"

	"github.com/elastic/go-elasticsearch/v8"
	"gorm.io/gorm"
)

type Repository interface {
	SearchLogs(ctx context.Context, query dto.LogQuery) (map[string]any, error)
	GetLogStats(ctx context.Context, query dto.LogQuery) (map[string]any, error)
	GetLogDetail(ctx context.Context, indexName, logID string) (map[string]any, error)
	GetAuditLogs(ctx context.Context, query dto.AuditLogQuery) ([]models.SystemAuditLog, int64, error)
}

type repository struct {
	db       *gorm.DB
	esClient *elasticsearch.Client
}

func NewRepository(db *gorm.DB, esClient *elasticsearch.Client) Repository {
	return &repository{
		db:       db,
		esClient: esClient,
	}
}

// parseCSVInt64 parses a comma-separated string of int64 values
func parseCSVInt64(csv string) []int64 {
	if strings.TrimSpace(csv) == "" {
		return nil
	}
	parts := strings.Split(csv, ",")
	result := make([]int64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if v, err := strconv.ParseInt(part, 10, 64); err == nil && v > 0 {
			result = append(result, v)
		}
	}
	return result
}

// buildElasticsearchQuery สร้าง dynamic query สำหรับส่งให้ Elasticsearch
func buildElasticsearchQuery(query dto.LogQuery) map[string]any {
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

	if query.LogLevel != "" {
		mustQueries = append(mustQueries, map[string]any{
			"term": map[string]any{
				"payload.log_level.keyword": query.LogLevel, // ใช้ payload.log_level หรือฟิลด์อื่นๆ ตามที่บันทึก
			},
		})
	}

	// Multi-project filter
	projectIDs := parseCSVInt64(query.ProjectIDs)
	if len(projectIDs) > 0 {
		mustQueries = append(mustQueries, map[string]any{
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
		mustQueries = append(mustQueries, map[string]any{
			"range": map[string]any{
				"@timestamp": rangeQuery,
			},
		})
	}

	return map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"must": mustQueries,
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

	indices := r.resolveSearchIndices(query.ProductID)
	res, err := r.esClient.Search(
		r.esClient.Search.WithContext(ctx),
		r.esClient.Search.WithIndex(indices...),
		r.esClient.Search.WithBody(&buf),
		r.esClient.Search.WithTrackTotalHits(true),
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

func (r *repository) GetLogDetail(ctx context.Context, indexName, logID string) (map[string]any, error) {
	res, err := r.esClient.Get(
		indexName,
		logID,
		r.esClient.Get.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch error: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("log not found") // You can handle custom errors if needed
		}
		return nil, fmt.Errorf("elasticsearch failed: %s", res.String())
	}

	var doc map[string]any
	if err := json.NewDecoder(res.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("failed to parse document: %w", err)
	}

	return doc, nil
}

func (r *repository) GetAuditLogs(ctx context.Context, q dto.AuditLogQuery) ([]models.SystemAuditLog, int64, error) {
	var auditLogs []models.SystemAuditLog
	query := r.db.WithContext(ctx).Model(&models.SystemAuditLog{})

	if q.ProductID != nil {
		query = query.Where("product_id = ?", *q.ProductID)
	}
	if q.ProjectID != nil {
		projIDStr := fmt.Sprintf("%d", *q.ProjectID)
		query = query.Where(
			"metadata->'query'->>'project_id' = ? OR metadata->'request'->>'project_id' = ? OR path LIKE ?",
			projIDStr, projIDStr, "%/projects/"+projIDStr+"%",
		)
	}
	if q.FeatureID != nil {
		featIDStr := fmt.Sprintf("%d", *q.FeatureID)
		query = query.Where(
			"metadata->'query'->>'category_id' = ? OR metadata->'request'->>'category_id' = ? OR "+
				"metadata->'query'->>'feature_id' = ? OR metadata->'request'->>'feature_id' = ? OR "+
				"path LIKE ? OR path LIKE ?",
			featIDStr, featIDStr, featIDStr, featIDStr, "%/features/"+featIDStr, "%/features/"+featIDStr+"/%",
		)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Limit(q.Limit).Offset(q.Offset).Find(&auditLogs).Error; err != nil {
		return nil, 0, err
	}

	return auditLogs, total, nil
}

func (r *repository) resolveSearchIndices(productID int) []string {
	indices := []string{fmt.Sprintf("omnilogs-product-%d-*", productID)}
	var policies []models.ElasticIndexPolicy
	if err := r.db.Where("product_id = ? AND is_active = TRUE", productID).Find(&policies).Error; err == nil {
		for _, policy := range policies {
			if strings.TrimSpace(policy.IndexPrefix) != "" {
				prefix := strings.ToLower(strings.TrimSpace(policy.IndexPrefix))
				prefix = strings.ReplaceAll(prefix, "_", "-")
				prefix = strings.ReplaceAll(prefix, " ", "-")
				indices = append(indices, fmt.Sprintf("%s-*", prefix))
			}
		}
	}
	return indices
}

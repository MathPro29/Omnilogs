package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"omnilogs-api/dto"
	"omnilogs-api/models"

	"github.com/elastic/go-elasticsearch/v8"
	"gorm.io/gorm"
)

type Provider struct {
	db *gorm.DB
	es *elasticsearch.Client
}

func New(db *gorm.DB, es *elasticsearch.Client) *Provider { return &Provider{db: db, es: es} }

func (p *Provider) IndexNames(ctx context.Context, productID, environmentID int, from, to *time.Time) ([]string, error) {
	if p.es == nil {
		return nil, fmt.Errorf("elasticsearch client is not configured")
	}
	patterns := []string{fmt.Sprintf("omnilogs-product-%d-env-%d-*", productID, environmentID)}
	if p.db != nil {
		var policies []models.ElasticIndexPolicy
		if err := p.db.WithContext(ctx).Where("product_id = ? AND is_active = TRUE AND (environment_id IS NULL OR environment_id = ?)", productID, environmentID).Find(&policies).Error; err != nil {
			return nil, err
		}
		for _, policy := range policies {
			prefix := strings.ToLower(strings.TrimSpace(policy.IndexPrefix))
			prefix = strings.ReplaceAll(strings.ReplaceAll(prefix, "_", "-"), " ", "-")
			if prefix != "" {
				patterns = append(patterns, fmt.Sprintf("%s-env-%d-*", prefix, environmentID))
			}
		}
	}
	patterns = uniqueStrings(patterns)
	response, err := p.es.Indices.Get(patterns, p.es.Indices.Get.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return []string{}, nil
	}
	if response.IsError() {
		return nil, fmt.Errorf("elasticsearch index lookup returned status %s", response.Status())
	}
	var indices map[string]json.RawMessage
	if err := json.NewDecoder(response.Body).Decode(&indices); err != nil {
		return nil, err
	}
	result := make([]string, 0, len(indices))
	for name := range indices {
		if !indexMatchesRange(name, from, to) {
			continue
		}
		result = append(result, name)
	}
	sort.Strings(result)
	return result, nil
}

func indexMatchesRange(indexName string, from, to *time.Time) bool {
	date, err := IndexDate(indexName)
	if err != nil {
		// Rollback indices created by older releases ended with an archive UUID
		// instead of a date. Include them and rely on the mandatory timestamp,
		// product and environment query guards to select the requested day.
		return strings.Contains(indexName, "-search-v3-rollback-")
	}
	return (from == nil || !date.Before(startOfDay(from.UTC()))) &&
		(to == nil || date.Before(startOfDay(to.UTC())))
}

// OldestTimestamp returns the first active log in the exact policy scope.
// It is used only to seed an existing-log backfill; normal runs reuse daily
// archive metadata and therefore do not repeatedly scan unbounded history.
func (p *Provider) OldestTimestamp(ctx context.Context, productID, environmentID int, categoryID, featureID, subFeatureID *int) (*time.Time, error) {
	indices, err := p.IndexNames(ctx, productID, environmentID, nil, nil)
	if err != nil || len(indices) == 0 {
		return nil, err
	}
	query := BuildQuery(productID, environmentID, nil, nil, categoryID, featureID, subFeatureID)
	query["size"] = 0
	query["aggs"] = map[string]any{"oldest": map[string]any{"min": map[string]any{"field": "@timestamp"}}}
	body, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	response, err := p.es.Search(
		p.es.Search.WithContext(ctx),
		p.es.Search.WithIndex(indices...),
		p.es.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.IsError() {
		return nil, fmt.Errorf("oldest active log search returned status %s", response.Status())
	}
	var payload struct {
		Aggregations struct {
			Oldest metric `json:"oldest"`
		} `json:"aggregations"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return parseMillis(payload.Aggregations.Oldest.Value), nil
}

// StorageBytes returns the physical Elasticsearch store used by all active
// and rollback indices for one product/environment, including replica bytes
// that are actually allocated.
func (p *Provider) StorageBytes(ctx context.Context, productID, environmentID int) (int64, error) {
	indices, err := p.IndexNames(ctx, productID, environmentID, nil, nil)
	if err != nil || len(indices) == 0 {
		return 0, err
	}
	response, err := p.es.Indices.Stats(
		p.es.Indices.Stats.WithContext(ctx),
		p.es.Indices.Stats.WithIndex(indices...),
		p.es.Indices.Stats.WithMetric("store"),
	)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	if response.IsError() {
		return 0, fmt.Errorf("elasticsearch storage stats returned status %s", response.Status())
	}
	var payload struct {
		All struct {
			Total struct {
				Store struct {
					SizeInBytes int64 `json:"size_in_bytes"`
				} `json:"store"`
			} `json:"total"`
		} `json:"_all"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return 0, err
	}
	return payload.All.Total.Store.SizeInBytes, nil
}

func IndexDate(indexName string) (time.Time, error) {
	parts := strings.Split(indexName, "-")
	if len(parts) == 0 {
		return time.Time{}, fmt.Errorf("invalid index name")
	}
	return time.Parse("2006.01.02", parts[len(parts)-1])
}

func BuildQuery(productID, environmentID int, from, to *time.Time, categoryID, featureID, subFeatureID *int) map[string]any {
	filters := []any{
		map[string]any{"term": map[string]any{"product_id": productID}},
		map[string]any{"term": map[string]any{"environment_id": environmentID}},
	}
	if from != nil || to != nil {
		rangeValue := map[string]any{}
		if from != nil {
			rangeValue["gte"] = from.UTC().Format(time.RFC3339Nano)
		}
		if to != nil {
			rangeValue["lt"] = to.UTC().Format(time.RFC3339Nano)
		}
		filters = append(filters, map[string]any{"range": map[string]any{"@timestamp": rangeValue}})
	}
	for field, value := range map[string]*int{"category_id": categoryID, "feature_id": featureID, "sub_feature_id": subFeatureID} {
		if value != nil && *value > 0 {
			filters = append(filters, map[string]any{"term": map[string]any{field: *value}})
		}
	}
	return map[string]any{"query": map[string]any{"bool": map[string]any{"filter": filters}}}
}

func (p *Provider) Historical(ctx context.Context, req dto.RetentionHistoricalRequest) (*dto.RetentionHistoricalResponse, error) {
	group, err := normalizeGroupBy(req.GroupBy)
	if err != nil {
		return nil, err
	}
	from, to := normalizeRange(req.DateFrom, req.DateTo)
	indices, err := p.IndexNames(ctx, req.ProductID, req.EnvironmentID, from, to)
	if err != nil {
		return nil, err
	}
	result := &dto.RetentionHistoricalResponse{ProductID: req.ProductID, EnvironmentID: req.EnvironmentID, Buckets: []dto.HistoricalLogBucket{}}
	if len(indices) == 0 {
		return result, nil
	}
	query := BuildQuery(req.ProductID, req.EnvironmentID, from, to, req.CategoryID, req.FeatureID, req.SubFeatureID)
	query["size"] = 0
	query["track_total_hits"] = true
	query["aggs"] = map[string]any{
		"oldest":         map[string]any{"min": map[string]any{"field": "@timestamp"}},
		"newest":         map[string]any{"max": map[string]any{"field": "@timestamp"}},
		"estimated_size": map[string]any{"sum": map[string]any{"field": "payload_size_bytes"}},
		"periods": map[string]any{"date_histogram": map[string]any{"field": "@timestamp", "calendar_interval": group, "min_doc_count": 1}, "aggs": map[string]any{
			"oldest":         map[string]any{"min": map[string]any{"field": "@timestamp"}},
			"newest":         map[string]any{"max": map[string]any{"field": "@timestamp"}},
			"estimated_size": map[string]any{"sum": map[string]any{"field": "payload_size_bytes"}},
		}},
	}
	body, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	response, err := p.es.Search(p.es.Search.WithContext(ctx), p.es.Search.WithIndex(indices...), p.es.Search.WithBody(bytes.NewReader(body)), p.es.Search.WithTrackTotalHits(true))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.IsError() {
		return nil, fmt.Errorf("historical log search returned status %s", response.Status())
	}
	var payload historicalSearchResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}
	result.DocumentCount = payload.Hits.Total.Value
	result.EstimatedSize = int64(payload.Aggregations.EstimatedSize.Value)
	result.OldestLog = parseMillis(payload.Aggregations.Oldest.Value)
	result.NewestLog = parseMillis(payload.Aggregations.Newest.Value)
	if result.DocumentCount > 0 {
		result.ArchiveStatus = p.archiveStatus(ctx, req.ProductID, req.EnvironmentID, from, to)
	}
	for _, bucket := range payload.Aggregations.Periods.Buckets {
		bucketFrom := time.UnixMilli(bucket.Key).UTC()
		bucketTo := nextBoundary(bucketFrom, group)
		result.Buckets = append(result.Buckets, dto.HistoricalLogBucket{
			Period: bucket.KeyAsString, DateFrom: bucketFrom, DateTo: bucketTo,
			OldestLog: parseMillis(bucket.Oldest.Value), NewestLog: parseMillis(bucket.Newest.Value),
			DocumentCount: bucket.DocCount, EstimatedBytes: int64(bucket.EstimatedSize.Value),
			ArchiveStatus: p.archiveStatus(ctx, req.ProductID, req.EnvironmentID, &bucketFrom, &bucketTo),
		})
	}
	return result, nil
}

func (p *Provider) Count(ctx context.Context, productID, environmentID int, from, to time.Time, categoryID, featureID, subFeatureID *int) (int64, error) {
	indices, err := p.IndexNames(ctx, productID, environmentID, &from, &to)
	if err != nil {
		return 0, err
	}
	if len(indices) == 0 {
		return 0, nil
	}
	query := BuildQuery(productID, environmentID, &from, &to, categoryID, featureID, subFeatureID)
	body, err := json.Marshal(query)
	if err != nil {
		return 0, err
	}
	response, err := p.es.Count(p.es.Count.WithContext(ctx), p.es.Count.WithIndex(indices...), p.es.Count.WithBody(bytes.NewReader(body)))
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	if response.IsError() {
		return 0, fmt.Errorf("historical log count returned status %s", response.Status())
	}
	var value struct {
		Count int64 `json:"count"`
	}
	if err := json.NewDecoder(response.Body).Decode(&value); err != nil {
		return 0, err
	}
	return value.Count, nil
}

func (p *Provider) DeleteActive(ctx context.Context, productID, environmentID int, from, to time.Time, categoryID, featureID, subFeatureID *int, expected int64) (int64, error) {
	if expected < 0 {
		return 0, fmt.Errorf("expected archive document count must not be negative")
	}
	indices, err := p.IndexNames(ctx, productID, environmentID, &from, &to)
	if err != nil {
		return 0, err
	}
	if expected == 0 {
		return 0, nil
	}
	if len(indices) == 0 {
		return 0, fmt.Errorf("safe delete guard failed: no indices matched")
	}
	actual, err := p.Count(ctx, productID, environmentID, from, to, categoryID, featureID, subFeatureID)
	if err != nil {
		return 0, err
	}
	if actual != expected {
		return 0, fmt.Errorf("safe delete guard failed: archive count %d does not match delete candidate count %d", expected, actual)
	}
	query := BuildQuery(productID, environmentID, &from, &to, categoryID, featureID, subFeatureID)
	body, err := json.Marshal(query)
	if err != nil {
		return 0, err
	}
	response, err := p.es.DeleteByQuery(indices, bytes.NewReader(body), p.es.DeleteByQuery.WithContext(ctx), p.es.DeleteByQuery.WithConflicts("abort"), p.es.DeleteByQuery.WithRefresh(true))
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	if response.IsError() {
		return 0, fmt.Errorf("safe delete returned status %s", response.Status())
	}
	var result struct {
		Deleted int64 `json:"deleted"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return 0, err
	}
	if result.Deleted != expected {
		return result.Deleted, fmt.Errorf("safe delete removed %d documents, expected %d", result.Deleted, expected)
	}
	return result.Deleted, nil
}

func normalizeRange(from, to *time.Time) (*time.Time, *time.Time) {
	now := time.Now().UTC()
	if from == nil {
		value := now.AddDate(-1, 0, 0)
		from = &value
	}
	if to == nil {
		value := now.Add(24 * time.Hour)
		to = &value
	}
	return from, to
}

func normalizeGroupBy(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "day":
		return "day", nil
	case "week":
		return "week", nil
	case "month":
		return "month", nil
	case "year":
		return "year", nil
	default:
		return "", fmt.Errorf("group_by must be day, week, month, or year")
	}
}

func nextBoundary(value time.Time, group string) time.Time {
	switch group {
	case "week":
		return value.AddDate(0, 0, 7)
	case "month":
		return value.AddDate(0, 1, 0)
	case "year":
		return value.AddDate(1, 0, 0)
	default:
		return value.AddDate(0, 0, 1)
	}
}

func startOfDay(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func (p *Provider) archiveStatus(ctx context.Context, productID, environmentID int, from, to *time.Time) string {
	if p.db == nil {
		return "NOT_ARCHIVED"
	}
	var archives []models.LogArchive
	query := p.db.WithContext(ctx).Where("product_id = ? AND environment_id = ? AND status IN ?", productID, environmentID, []string{"ARCHIVED", "VERIFIED", "FAILED"})
	if from != nil {
		query = query.Where("date_to >= ?", from)
	}
	if to != nil {
		query = query.Where("date_from < ?", to)
	}
	if query.Order("verified_at DESC NULLS LAST, created_at DESC").Find(&archives).Error != nil || len(archives) == 0 {
		return "NOT_ARCHIVED"
	}
	for _, archive := range archives {
		if archive.Status != nil && *archive.Status == "VERIFIED" {
			return "VERIFIED"
		}
	}
	for _, archive := range archives {
		if archive.Status != nil && *archive.Status == "ARCHIVED" {
			return "ARCHIVED"
		}
	}
	return "FAILED"
}

type historicalSearchResponse struct {
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
	} `json:"hits"`
	Aggregations struct {
		Oldest        metric `json:"oldest"`
		Newest        metric `json:"newest"`
		EstimatedSize metric `json:"estimated_size"`
		Periods       struct {
			Buckets []struct {
				Key           int64  `json:"key"`
				KeyAsString   string `json:"key_as_string"`
				DocCount      int64  `json:"doc_count"`
				Oldest        metric `json:"oldest"`
				Newest        metric `json:"newest"`
				EstimatedSize metric `json:"estimated_size"`
			} `json:"buckets"`
		} `json:"periods"`
	} `json:"aggregations"`
}
type metric struct {
	Value float64 `json:"value"`
}

func parseMillis(value float64) *time.Time {
	if value == 0 {
		return nil
	}
	t := time.UnixMilli(int64(value)).UTC()
	return &t
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"omnilogs-api/models"
)

type indexedLogMeta struct {
	ProjectID       *int
	CategoryID      *int
	DurationMs      *int
	LogLevel        *string
	EventType       *string
	SourceRequestID *string
	CorrelationID   *string
	TraceID         *string
	SpanID          *string
	ParentSpanID    *string
	RequestMethod   *string
	RequestPath     *string
	RoutePattern    *string
	FeatureFullPath *string
	FeaturePathIDs  *string
	Timestamp       time.Time
}

func (u *usecase) resolveIndexName(ctx context.Context, productID int, environmentID *int, timestamp time.Time) string {
	policy, err := u.repo.FindIndexPolicy(ctx, productID, environmentID)
	if err == nil && strings.TrimSpace(policy.IndexPrefix) != "" {
		return fmt.Sprintf("%s-%s", normalizeIndexSegment(policy.IndexPrefix), timestamp.UTC().Format("2006.01.02"))
	}
	return fmt.Sprintf("omnilogs-product-%d-%s", productID, timestamp.UTC().Format("2006.01.02"))
}

func (u *usecase) indexDocument(ctx context.Context, indexName, docID string, document map[string]any) (int, error) {
	body, err := json.Marshal(document)
	if err != nil {
		return 0, err
	}

	res, err := u.esClient.Index(
		indexName,
		bytes.NewReader(body),
		u.esClient.Index.WithDocumentID(docID),
		u.esClient.Index.WithContext(ctx),
	)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return res.StatusCode, fmt.Errorf("elasticsearch returned status %d: %s", res.StatusCode, res.String())
	}
	return res.StatusCode, nil
}

// รับ Raw JSON ให้อยู่ในรูป map[string]any
func buildElasticDocument(batch *models.LogQueueBatch, item *models.LogQueueItem) (map[string]any, indexedLogMeta, error) {
	var payload any
	if err := json.Unmarshal(item.InputPayload, &payload); err != nil {
		return nil, indexedLogMeta{}, err
	}

	// [KEY: Elastic Documents]
	document := map[string]any{
		"batch_id":        batch.BatchID,
		"queue_item_id":   item.QueueItemID,
		"sequence_no":     item.SequenceNo,
		"source_type":     item.SourceType,
		"source_platform": item.SourcePlatform,
		"received_at":     time.Now().UTC().Format(time.RFC3339Nano),
		"payload":         payload,
	}
	if batch.ProductID != nil {
		document["product_id"] = *batch.ProductID
	}
	if batch.EnvironmentID != nil {
		document["environment_id"] = *batch.EnvironmentID
	}
	if batch.SourceID != nil {
		document["source_id"] = *batch.SourceID
	}

	// ใช้สำหรับ Data Validate และ ไปบันทึกใน log_index_ref
	meta := indexedLogMeta{Timestamp: time.Now().UTC()}
	if asMap, ok := payload.(map[string]any); ok {
		meta.ProjectID = intPtrFromAny(asMap["project_id"])
		meta.CategoryID = intPtrFromAny(asMap["category_id"])
		meta.DurationMs = intPtrFromAny(asMap["duration_ms"])
		meta.LogLevel = stringPtrFromAny(asMap["log_level"])
		meta.EventType = stringPtrFromAny(asMap["event_type"])
		meta.SourceRequestID = stringPtrFromAny(asMap["source_request_id"])
		meta.CorrelationID = stringPtrFromAny(asMap["correlation_id"])
		meta.TraceID = stringPtrFromAny(asMap["trace_id"])
		meta.SpanID = stringPtrFromAny(asMap["span_id"])
		meta.ParentSpanID = stringPtrFromAny(asMap["parent_span_id"])
		meta.RequestMethod = stringPtrFromAny(asMap["request_method"])
		meta.RequestPath = stringPtrFromAny(asMap["request_path"])
		meta.RoutePattern = stringPtrFromAny(asMap["route_pattern"])
		meta.FeatureFullPath = stringPtrFromAny(asMap["feature_full_path"])
		meta.FeaturePathIDs = stringPtrFromAny(asMap["feature_path_ids"])
		if timestamp, ok := timePtrFromAny(asMap["timestamp"]); ok {
			meta.Timestamp = timestamp
		}
	}
	document["@timestamp"] = meta.Timestamp.UTC().Format(time.RFC3339Nano)

	return document, meta, nil
}

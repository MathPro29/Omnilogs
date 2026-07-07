package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"omnilogs-api/internal/queue"
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

func (u *usecase) bulkIndexDocuments(ctx context.Context, entries []processedLog) ([]processedLog, error) {
	var body bytes.Buffer
	for i := range entries {
		meta := map[string]any{
			"index": map[string]any{
				"_index": entries[i].indexName,
				"_id":    entries[i].message.LogID,
			},
		}
		metaBytes, err := json.Marshal(meta)
		if err != nil {
			return nil, err
		}
		docBytes, err := json.Marshal(entries[i].document)
		if err != nil {
			return nil, err
		}
		body.Write(metaBytes)
		body.WriteByte('\n')
		body.Write(docBytes)
		body.WriteByte('\n')
	}

	res, err := u.esClient.Bulk(bytes.NewReader(body.Bytes()), u.esClient.Bulk.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch bulk request failed with status %d", res.StatusCode)
	}

	var response struct {
		Errors bool `json:"errors"`
		Items  []map[string]struct {
			Status int            `json:"status"`
			Error  map[string]any `json:"error"`
		} `json:"items"`
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	successful := make([]processedLog, 0, len(entries))
	for i := range entries {
		if i >= len(response.Items) {
			return nil, fmt.Errorf("unexpected elasticsearch bulk response length")
		}

		item := response.Items[i]["index"]
		if item.Status >= 300 {
			reason := "bulk index item failed"
			if msg, ok := item.Error["reason"].(string); ok && msg != "" {
				reason = msg
			}
			if err := u.failMessage(ctx, entries[i].natsMsg, entries[i].message, "ELASTICSEARCH", "INDEX_REQUEST_FAILED", reason, true); err != nil {
				return nil, err
			}
			continue
		}

		indexedAt := time.Now().UTC()
		entries[i].responseCode = item.Status
		entries[i].indexRef.ResponseStatusCode = &item.Status
		entries[i].indexRef.IndexedAt = &indexedAt
		entries[i].indexRef.LastSyncAt = &indexedAt
		successful = append(successful, entries[i])
	}

	return successful, nil
}

func buildElasticDocument(message *queue.LogMessage, payload map[string]any) (map[string]any, indexedLogMeta, error) {
	document := map[string]any{
		"batch_id":        message.BatchID,
		"queue_item_id":   message.QueueItemID,
		"sequence_no":     message.SequenceNo,
		"source_type":     message.SourceType,
		"source_platform": message.SourcePlatform,
		"received_at":     message.PublishedAt.UTC().Format(time.RFC3339Nano),
		"payload":         payload,
	}
	if message.ProductID != nil {
		document["product_id"] = *message.ProductID
	}
	if message.EnvironmentID != nil {
		document["environment_id"] = *message.EnvironmentID
	}
	if message.SourceID != nil {
		document["source_id"] = *message.SourceID
	}

	meta := indexedLogMeta{Timestamp: time.Now().UTC()}
	meta.ProjectID = intPtrFromAny(payload["project_id"])
	meta.CategoryID = intPtrFromAny(payload["category_id"])
	meta.DurationMs = intPtrFromAny(payload["duration_ms"])
	meta.LogLevel = stringPtrFromAny(payload["log_level"])
	meta.EventType = stringPtrFromAny(payload["event_type"])
	meta.SourceRequestID = stringPtrFromAny(payload["source_request_id"])
	meta.CorrelationID = stringPtrFromAny(payload["correlation_id"])
	meta.TraceID = stringPtrFromAny(payload["trace_id"])
	meta.SpanID = stringPtrFromAny(payload["span_id"])
	meta.ParentSpanID = stringPtrFromAny(payload["parent_span_id"])
	meta.RequestMethod = stringPtrFromAny(payload["request_method"])
	meta.RequestPath = stringPtrFromAny(payload["request_path"])
	meta.RoutePattern = stringPtrFromAny(payload["route_pattern"])
	meta.FeatureFullPath = stringPtrFromAny(payload["feature_full_path"])
	meta.FeaturePathIDs = stringPtrFromAny(payload["feature_path_ids"])
	if timestamp, ok := timePtrFromAny(payload["timestamp"]); ok {
		meta.Timestamp = timestamp
	}
	document["@timestamp"] = meta.Timestamp.UTC().Format(time.RFC3339Nano)

	return document, meta, nil
}

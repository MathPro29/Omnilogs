package worker

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"omnilogs-api/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	defaultPollInterval = 2 * time.Second
	defaultLockTimeout  = 10 * time.Minute
)

// Processor ใช้รัน pipeline ประมวลผล log แบบ asynchronous โดยดึง batch จาก
// PostgreSQL, ส่งแต่ละ log ไปยัง Elasticsearch และอัปเดตตารางติดตามสถานะ
// เพื่อให้ฝั่ง API สามารถตรวจสอบผลการประมวลผลย้อนหลังได้
type Processor struct {
	db         *gorm.DB
	elasticURL string
	workerID   string
	client     *http.Client
}

func NewProcessor(db *gorm.DB, elasticURL string) *Processor {
	return &Processor{
		db:         db,
		elasticURL: strings.TrimRight(strings.TrimSpace(elasticURL), "/"),
		workerID:   "worker-" + newUUID(),
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Run จะวน poll หา batch ที่รอประมวลผลอยู่เรื่อย ๆ จนกว่าจะมีการสั่งปิด worker
func (p *Processor) Run(ctx context.Context) error {
	ticker := time.NewTicker(defaultPollInterval)
	defer ticker.Stop()

	for {
		if err := p.processNextBatch(ctx); err != nil {
			log.Printf("worker %s: batch processing failed: %v", p.workerID, err)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (p *Processor) processNextBatch(ctx context.Context) error {
	batch, err := p.claimNextBatch(ctx)
	if err != nil || batch == nil {
		return err
	}

	log.Printf("worker %s: claimed batch %s with %d log(s)", p.workerID, batch.BatchID, batch.TotalLogs)

	items, err := p.loadPendingItems(ctx, batch.BatchID)
	if err != nil {
		return p.failBatch(ctx, batch, err.Error())
	}
	if len(items) == 0 {
		return p.finishBatch(ctx, batch.BatchID, "COMPLETED", nil)
	}

	var processedCount int
	var retryCount int
	var failedCount int

	for i := range items {
		if err := p.processItem(ctx, batch, &items[i]); err != nil {
			if strings.Contains(err.Error(), "retry scheduled") {
				retryCount++
			} else {
				failedCount++
			}
			continue
		}
		processedCount++
	}

	status := "COMPLETED"
	switch {
	case processedCount == 0 && retryCount > 0 && failedCount == 0:
		status = "RETRY_PENDING"
	case processedCount == 0 && failedCount > 0:
		status = "FAILED"
	case retryCount > 0 || failedCount > 0:
		status = "PARTIAL"
	}

	message := fmt.Sprintf("processed=%d retry_pending=%d failed=%d", processedCount, retryCount, failedCount)
	return p.finishBatch(ctx, batch.BatchID, status, &message)
}

func (p *Processor) claimNextBatch(ctx context.Context) (*models.LogQueueBatch, error) {
	var claimed *models.LogQueueBatch
	err := p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var batch models.LogQueueBatch
		lockCutoff := time.Now().Add(-defaultLockTimeout)

		query := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status IN ?", []string{"QUEUED", "RETRY_PENDING", "PARTIAL"}).
			Where("(available_at IS NULL OR available_at <= ?)", time.Now()).
			Where("(locked_at IS NULL OR locked_at < ?)", lockCutoff).
			Order("priority DESC").
			Order("received_at ASC NULLS LAST").
			Order("created_at ASC NULLS LAST")

		if err := query.First(&batch).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil
			}
			return err
		}

		now := time.Now()
		updates := map[string]any{
			"status":                "PROCESSING",
			"worker_id":             p.workerID,
			"locked_by":             p.workerID,
			"locked_at":             now,
			"processing_started_at": now,
			"processing_attempts":   gorm.Expr("processing_attempts + 1"),
			"error_message":         nil,
		}
		if err := tx.Model(&models.LogQueueBatch{}).
			Where("batch_id = ?", batch.BatchID).
			Updates(updates).Error; err != nil {
			return err
		}

		batch.Status = "PROCESSING"
		batch.WorkerID = stringPtr(p.workerID)
		batch.LockedBy = stringPtr(p.workerID)
		batch.LockedAt = &now
		batch.ProcessingStartedAt = &now
		claimed = &batch
		return nil
	})
	return claimed, err
}

func (p *Processor) loadPendingItems(ctx context.Context, batchID string) ([]models.LogQueueItem, error) {
	var items []models.LogQueueItem
	err := p.db.WithContext(ctx).
		Where("batch_id = ?", batchID).
		Where("status IN ?", []string{"PENDING", "RETRY_PENDING"}).
		Where("(next_retry_at IS NULL OR next_retry_at <= ?)", time.Now()).
		Order("sequence_no ASC").
		Find(&items).Error
	return items, err
}

func (p *Processor) processItem(ctx context.Context, batch *models.LogQueueBatch, item *models.LogQueueItem) error {
	now := time.Now()
	if err := p.db.WithContext(ctx).Model(&models.LogQueueItem{}).
		Where("queue_item_id = ?", item.QueueItemID).
		Updates(map[string]any{
			"status":                "PROCESSING",
			"worker_id":             p.workerID,
			"processing_started_at": now,
			"processing_attempts":   gorm.Expr("processing_attempts + 1"),
			"error_message":         nil,
		}).Error; err != nil {
		return err
	}

	if batch.ProductID == nil || batch.EnvironmentID == nil {
		return p.handleItemFailure(ctx, batch, item, "VALIDATION", "BATCH_METADATA_MISSING", "product_id and environment_id are required before indexing")
	}

	document, meta, err := buildElasticDocument(batch, item)
	if err != nil {
		return p.handleItemFailure(ctx, batch, item, "TRANSFORM", "INVALID_PAYLOAD", err.Error())
	}

	indexName := p.resolveIndexName(ctx, *batch.ProductID, batch.EnvironmentID, meta.Timestamp)
	docID := newUUID()
	indexedAt := time.Now()

	statusCode, syncErr := p.indexDocument(ctx, indexName, docID, document)
	if syncErr != nil {
		return p.handleItemFailure(ctx, batch, item, "ELASTICSEARCH", "INDEX_REQUEST_FAILED", syncErr.Error())
	}

	indexRef := models.LogIndexRef{
		LogID:              docID,
		ProductID:          *batch.ProductID,
		ProjectID:          meta.ProjectID,
		CategoryID:         meta.CategoryID,
		EnvironmentID:      *batch.EnvironmentID,
		SourceID:           batch.SourceID,
		BatchID:            &batch.BatchID,
		QueueItemID:        &item.QueueItemID,
		ResponseStatusCode: &statusCode,
		DurationMs:         meta.DurationMs,
		LogLevel:           meta.LogLevel,
		EventType:          meta.EventType,
		SourceRequestID:    meta.SourceRequestID,
		CorrelationID:      meta.CorrelationID,
		TraceID:            meta.TraceID,
		SpanID:             meta.SpanID,
		ParentSpanID:       meta.ParentSpanID,
		RequestMethod:      meta.RequestMethod,
		RequestPath:        meta.RequestPath,
		RoutePattern:       meta.RoutePattern,
		FeatureFullPath:    meta.FeatureFullPath,
		FeaturePathIDs:     meta.FeaturePathIDs,
		Timestamp:          meta.Timestamp,
		IngestedAt:         &indexedAt,
		ElasticIndex:       indexName,
		ElasticDocumentID:  docID,
		IndexStatus:        "INDEXED",
		IndexedAt:          &indexedAt,
		LastSyncAt:         &indexedAt,
	}

	if err := p.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "elastic_index"}, {Name: "elastic_document_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"index_status", "indexed_at", "last_sync_at", "sync_error", "response_status_code", "updated_at"}),
	}).Create(&indexRef).Error; err != nil {
		return p.handleItemFailure(ctx, batch, item, "DATABASE", "INDEX_REF_CREATE_FAILED", err.Error())
	}

	return p.db.WithContext(ctx).Model(&models.LogQueueItem{}).
		Where("queue_item_id = ?", item.QueueItemID).
		Updates(map[string]any{
			"status":        "PROCESSED",
			"processed_at":  indexedAt,
			"last_retry_at": indexedAt,
			"error_message": nil,
		}).Error
}

func (p *Processor) handleItemFailure(ctx context.Context, batch *models.LogQueueBatch, item *models.LogQueueItem, stage, failureType, reason string) error {
	now := time.Now()
	retryCount := item.RetryCount + 1

	nextStatus := "FAILED"
	var nextRetryAt *time.Time
	if retryCount <= item.MaxRetryCount {
		nextStatus = "RETRY_PENDING"
		retryAt := now.Add(time.Duration(retryCount) * time.Minute)
		nextRetryAt = &retryAt
	}

	reasonCopy := reason
	itemFailure := models.LogFailure{
		FailureID:     newUUID(),
		AttemptNo:     item.ProcessingAttempts + 1,
		QueueItemID:   item.QueueItemID,
		BatchID:       batch.BatchID,
		ProductID:     batch.ProductID,
		SourceID:      batch.SourceID,
		EnvironmentID: batch.EnvironmentID,
		FailureStage:  stage,
		FailureType:   failureType,
		Reason:        &reasonCopy,
		RetryCount:    retryCount,
		MaxRetryCount: item.MaxRetryCount,
		NextRetryAt:   nextRetryAt,
		LastRetryAt:   &now,
		Status:        nextStatus,
	}
	_ = p.db.WithContext(ctx).Create(&itemFailure).Error

	_ = p.db.WithContext(ctx).Model(&models.LogQueueItem{}).
		Where("queue_item_id = ?", item.QueueItemID).
		Updates(map[string]any{
			"status":                nextStatus,
			"retry_count":           retryCount,
			"next_retry_at":         nextRetryAt,
			"last_retry_at":         now,
			"error_message":         reason,
			"processing_started_at": now,
		}).Error

	if nextStatus == "RETRY_PENDING" {
		return fmt.Errorf("retry scheduled: %s", reason)
	}
	return fmt.Errorf(reason)
}

func (p *Processor) finishBatch(ctx context.Context, batchID, status string, message *string) error {
	now := time.Now()
	return p.db.WithContext(ctx).Model(&models.LogQueueBatch{}).
		Where("batch_id = ?", batchID).
		Updates(map[string]any{
			"status":        status,
			"processed_at":  now,
			"locked_by":     nil,
			"locked_at":     nil,
			"error_message": message,
		}).Error
}

func (p *Processor) failBatch(ctx context.Context, batch *models.LogQueueBatch, reason string) error {
	return p.finishBatch(ctx, batch.BatchID, "FAILED", &reason)
}

func (p *Processor) resolveIndexName(ctx context.Context, productID int, environmentID *int, timestamp time.Time) string {
	var policy models.ElasticIndexPolicy
	query := p.db.WithContext(ctx).Where("product_id = ? AND is_active = TRUE", productID)
	if environmentID != nil {
		query = query.Where("(environment_id IS NULL OR environment_id = ?)", *environmentID).Order("environment_id DESC NULLS LAST")
	} else {
		query = query.Where("environment_id IS NULL")
	}

	if err := query.First(&policy).Error; err == nil && strings.TrimSpace(policy.IndexPrefix) != "" {
		return fmt.Sprintf("%s-%s", normalizeIndexSegment(policy.IndexPrefix), timestamp.UTC().Format("2006.01.02"))
	}
	return fmt.Sprintf("omnilogs-product-%d-%s", productID, timestamp.UTC().Format("2006.01.02"))
}

func (p *Processor) indexDocument(ctx context.Context, indexName, docID string, document map[string]any) (int, error) {
	body, err := json.Marshal(document)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/%s/_doc/%s", p.elasticURL, indexName, docID), bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return resp.StatusCode, fmt.Errorf("elasticsearch returned %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}
	return resp.StatusCode, nil
}

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

// buildElasticDocument ใช้จัดรูป JSON ที่อยู่ใน queue ให้กลายเป็น document
// มาตรฐานสำหรับ Elasticsearch และดึง metadata ที่ต้องการไปเก็บใน
// log_index_refs เพื่อใช้ค้นหาและออกรายงานภายหลัง
func buildElasticDocument(batch *models.LogQueueBatch, item *models.LogQueueItem) (map[string]any, indexedLogMeta, error) {
	var payload any
	if err := json.Unmarshal(item.InputPayload, &payload); err != nil {
		return nil, indexedLogMeta{}, err
	}

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

func intPtrFromAny(value any) *int {
	switch typed := value.(type) {
	case float64:
		result := int(typed)
		return &result
	case int:
		result := typed
		return &result
	}
	return nil
}

func stringPtrFromAny(value any) *string {
	typed, ok := value.(string)
	if !ok {
		return nil
	}
	typed = strings.TrimSpace(typed)
	if typed == "" {
		return nil
	}
	return &typed
}

func timePtrFromAny(value any) (time.Time, bool) {
	raw, ok := value.(string)
	if !ok {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

func normalizeIndexSegment(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "_", "-")
	value = strings.ReplaceAll(value, " ", "-")
	if value == "" {
		return "omnilogs"
	}
	return value
}

func stringPtr(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func newUUID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "00000000-0000-4000-8000-000000000000"
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	dst := make([]byte, 36)
	hex.Encode(dst[0:8], bytes[0:4])
	dst[8] = '-'
	hex.Encode(dst[9:13], bytes[4:6])
	dst[13] = '-'
	hex.Encode(dst[14:18], bytes[6:8])
	dst[18] = '-'
	hex.Encode(dst[19:23], bytes[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:36], bytes[10:16])
	return string(dst)
}

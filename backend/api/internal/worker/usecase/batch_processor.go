package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"sync"
	"time"

	"omnilogs-api/internal/queue"
	"omnilogs-api/models"
	"omnilogs-api/utils"

	"github.com/nats-io/nats.go"
)

type batchCache struct {
	fields     sync.Map // productID -> []models.LogFieldDefinition
	rules      sync.Map // productID -> []models.LogMaskingRule
	hierarchy  sync.Map // string -> error (caching valid or invalid results)
	indexNames sync.Map // productID-environmentID-date -> index name
}

type sensitiveMatchers struct {
	fieldByKey  map[string]*models.LogFieldDefinition
	fieldByPath map[string]*models.LogFieldDefinition
	ruleByKey   map[string]*models.LogMaskingRule
	ruleByPath  map[string]*models.LogMaskingRule
}

type processedLog struct {
	message         queue.LogMessage
	natsMsg         *nats.Msg
	document        map[string]any
	meta            indexedLogMeta
	indexName       string
	responseCode    int
	secrets         []models.LogSensitiveFieldSecret
	objectRef       models.LogObjectStorageRef
	indexRef        models.LogIndexRef
	originalPayload []byte
}

func (u *usecase) processFetchedMessages(ctx context.Context, msgs []*nats.Msg) error {
	batchIDs := collectBatchIDs(msgs)
	if len(batchIDs) > 0 {
		defer u.refreshBatchStatuses(ctx, batchIDs)
	}

	if err := u.markFetchedBatchesProcessing(ctx, msgs); err != nil {
		return err
	}

	prepared, err := u.prepareLogs(ctx, msgs)
	if err != nil {
		return err
	}
	if len(prepared) == 0 {
		return nil
	}

	prepared, err = u.bulkIndexDocuments(ctx, prepared)
	if err != nil {
		for _, item := range prepared {
			if retryErr := u.failMessage(ctx, item.natsMsg, item.message, "ELASTICSEARCH", "BULK_REQUEST_FAILED", err.Error(), true); retryErr != nil {
				return retryErr
			}
		}
		return nil
	}
	if len(prepared) == 0 {
		return nil
	}

	indexRefs := make([]models.LogIndexRef, 0, len(prepared))
	objectRefs := make([]models.LogObjectStorageRef, 0, len(prepared))
	secrets := make([]models.LogSensitiveFieldSecret, 0, len(prepared))

	for _, item := range prepared {
		indexRefs = append(indexRefs, item.indexRef)
		objectRefs = append(objectRefs, item.objectRef)
		secrets = append(secrets, item.secrets...)
	}

	if err := u.repo.BulkPersistSuccesses(ctx, indexRefs, objectRefs, secrets); err != nil {
		return err
	}

	for _, item := range prepared {
		if err := item.natsMsg.Ack(); err != nil {
			return err
		}

		if item.message.ProductID != nil {
			subject := fmt.Sprintf("omnilogs.logs.live.%d", *item.message.ProductID)
			item.document["log_id"] = item.message.LogID
			if payload, err := json.Marshal(item.document); err == nil {
				_ = u.natsQueue.Broadcast(subject, payload)
			}
		}
	}

	return nil
}

func collectBatchIDs(msgs []*nats.Msg) map[string]struct{} {
	batchIDs := make(map[string]struct{})
	for _, msg := range msgs {
		envelope, err := decodeMessage(msg)
		if err != nil || envelope.BatchID == "" {
			continue
		}
		batchIDs[envelope.BatchID] = struct{}{}
	}
	return batchIDs
}

func (u *usecase) refreshBatchStatuses(ctx context.Context, batchIDs map[string]struct{}) {
	for batchID := range batchIDs {
		if err := u.repo.RefreshBatchStatusFromResults(ctx, batchID); err != nil {
			slog.Error("failed to refresh batch status", "batch_id", batchID, "error", err)
		}
	}
}

func (u *usecase) markFetchedBatchesProcessing(ctx context.Context, msgs []*nats.Msg) error {
	batchIDs := make(map[string]struct{})
	for _, msg := range msgs {
		envelope, err := decodeMessage(msg)
		if err != nil || envelope.BatchID == "" {
			continue
		}
		batchIDs[envelope.BatchID] = struct{}{}
	}
	if len(batchIDs) == 0 {
		return
	}

	ids := make([]string, 0, len(batchIDs))
	for batchID := range batchIDs {
		ids = append(ids, batchID)
	}

	now := time.Now().UTC()
	if err := u.repo.DB().WithContext(ctx).Model(&models.LogQueueBatch{}).
		Where("batch_id IN ?", ids).
		Updates(map[string]any{
			"status":                "PROCESSING",
			"processing_started_at": now,
			"processed_at":          nil,
			"error_message":         nil,
		}).Error; err != nil {
		slog.Error("failed to mark fetched batches as processing", "error", err)
		return err
	}
	return nil
}

func (u *usecase) prepareLogs(ctx context.Context, msgs []*nats.Msg) ([]processedLog, error) {
	cache := &batchCache{}
	uniqueProductIDs := make(map[int]struct{})

	for _, msg := range msgs {
		envelope, err := decodeMessage(msg)
		if err == nil && envelope.ProductID != nil {
			uniqueProductIDs[*envelope.ProductID] = struct{}{}
		}
	}

	for pid := range uniqueProductIDs {
		fields, err := u.repo.GetActiveFieldDefinitions(ctx, pid)
		if err != nil {
			return nil, fmt.Errorf("load field definitions for product %d: %w", pid, err)
		}
		cache.fields.Store(pid, fields)

		rules, err := u.repo.GetLogMaskingRules(ctx, pid)
		if err != nil {
			return nil, fmt.Errorf("load masking rules for product %d: %w", pid, err)
		}
		cache.rules.Store(pid, rules)
	}

	type prepareResult struct {
		entry *processedLog
		err   error
	}
	workerCount := runtime.GOMAXPROCS(0) * 2
	if workerCount < 1 {
		workerCount = 1
	}
	if workerCount > len(msgs) {
		workerCount = len(msgs)
	}
	jobs := make(chan *nats.Msg)
	results := make(chan prepareResult, len(msgs))
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for msg := range jobs {
				entry, err := u.prepareLog(ctx, msg, cache)
				results <- prepareResult{entry: entry, err: err}
			}
		}()
	}
	for _, msg := range msgs {
		jobs <- msg
	}
	close(jobs)
	wg.Wait()
	close(results)

	prepared := make([]processedLog, 0, len(msgs))
	for result := range results {
		if result.err != nil {
			return nil, result.err
		}
		if result.entry != nil {
			prepared = append(prepared, *result.entry)
		}
	}
	return prepared, nil
}

func (u *usecase) prepareLog(ctx context.Context, msg *nats.Msg, cache *batchCache) (*processedLog, error) {
	envelope, err := decodeMessage(msg)
	if err != nil {
		return nil, msg.Ack()
	}

	if envelope.ProductID == nil || envelope.EnvironmentID == nil {
		return nil, u.failMessage(ctx, msg, envelope, "VALIDATION", "BATCH_METADATA_MISSING", "product_id and environment_id are required before indexing", false)
	}

	var fields []models.LogFieldDefinition
	if val, ok := cache.fields.Load(*envelope.ProductID); ok {
		fields = val.([]models.LogFieldDefinition)
	}

	var rules []models.LogMaskingRule
	if val, ok := cache.rules.Load(*envelope.ProductID); ok {
		rules = val.([]models.LogMaskingRule)
	}
	var payload map[string]any
	if err := json.Unmarshal(envelope.InputPayload, &payload); err != nil {
		return nil, u.failMessage(ctx, msg, envelope, "TRANSFORM", "INVALID_PAYLOAD", err.Error(), false)
	}
	if payload == nil {
		return nil, u.failMessage(ctx, msg, envelope, "TRANSFORM", "INVALID_PAYLOAD", "payload is null or empty", false)
	}
	_, rawMeta, err := buildElasticDocument(&envelope, payload)
	if err != nil {
		return nil, u.failMessage(ctx, msg, envelope, "TRANSFORM", "INVALID_PAYLOAD", err.Error(), false)
	}
	fields = applicableFieldDefinitions(fields, rawMeta)
	sensitiveFields := make([]models.LogFieldDefinition, 0, len(fields))
	for _, field := range fields {
		if field.IsSensitive {
			sensitiveFields = append(sensitiveFields, field)
		}
	}
	matchers := buildSensitiveMatchers(sensitiveFields, rules)
	// ควบคุมการประมวลผล
	// 1. EXTRACT & TRANSFORM & VALIDATE PIPELINE
	for _, field := range fields {
		path := strings.TrimSpace(field.FieldKey)
		if field.FieldPath != nil && strings.TrimSpace(*field.FieldPath) != "" {
			path = canonicalDynamicPath(*field.FieldPath)
		}
		if path == "" {
			continue
		}
		val, exists := readJSONPath(payload, path)
		if exists {
			transformed, err := PipelineTransformer(ctx, field, val)
			if err != nil {
				return nil, u.failMessage(ctx, msg, envelope, "TRANSFORM", "TRANSFORMATION_FAILED", err.Error(), false)
			}
			if path != "$" {
				if err := writeJSONPath(payload, path, transformed); err != nil {
					return nil, u.failMessage(ctx, msg, envelope, "TRANSFORM", "TRANSFORMATION_FAILED", err.Error(), false)
				}
			}

			if err := PipelineValidator(ctx, field, transformed); err != nil {
				return nil, u.failMessage(ctx, msg, envelope, "VALIDATION", "INVALID_FIELD_VALUE", err.Error(), false)
			}
		} else {
			if err := PipelineValidator(ctx, field, nil); err != nil {
				return nil, u.failMessage(ctx, msg, envelope, "VALIDATION", "INVALID_FIELD_VALUE", err.Error(), false)
			}
		}
	}

	originalPayload := append([]byte(nil), envelope.InputPayload...)
	document, meta, err := buildElasticDocument(&envelope, payload)
	if err != nil {
		return nil, u.failMessage(ctx, msg, envelope, "TRANSFORM", "INVALID_PAYLOAD", err.Error(), false)
	}
	if err := u.validateLogHierarchy(ctx, *envelope.ProductID, *envelope.EnvironmentID, meta, cache); err != nil {
		return nil, u.failMessage(ctx, msg, envelope, "VALIDATION", "INVALID_LOG_HIERARCHY", err.Error(), false)
	}

	secrets, err := u.maskPayloadAndExtractSecrets(payload, "", *envelope.ProductID, envelope.LogID, fields, rules, matchers)
	if err != nil {
		return nil, u.failMessage(ctx, msg, envelope, "TRANSFORM", "MASKING_FAILED", err.Error(), true)
	}

	document, meta, err = buildElasticDocument(&envelope, payload)
	if err != nil {
		return nil, u.failMessage(ctx, msg, envelope, "TRANSFORM", "PAYLOAD_MARSHAL_FAILED", err.Error(), false)
	}

	indexedAt := time.Now().UTC()
	indexName := u.resolveIndexNameCached(ctx, *envelope.ProductID, envelope.EnvironmentID, meta.Timestamp, cache)
	objectRef, err := u.buildObjectStorageRef(envelope, originalPayload)
	if err != nil {
		return nil, u.failMessage(ctx, msg, envelope, "DATABASE", "PAYLOAD_ARCHIVE_CREATE_FAILED", err.Error(), true)
	}

	indexRef := models.LogIndexRef{
		LogID:             envelope.LogID,
		ProductID:         *envelope.ProductID,
		ProjectID:         meta.ProjectID,
		CategoryID:        meta.CategoryID,
		EnvironmentID:     *envelope.EnvironmentID,
		SourceID:          envelope.SourceID,
		BatchID:           &envelope.BatchID,
		QueueItemID:       &envelope.QueueItemID,
		DurationMs:        meta.DurationMs,
		LogLevel:          meta.LogLevel,
		EventType:         meta.EventType,
		SourceRequestID:   meta.SourceRequestID,
		CorrelationID:     meta.CorrelationID,
		TraceID:           meta.TraceID,
		SpanID:            meta.SpanID,
		ParentSpanID:      meta.ParentSpanID,
		RequestMethod:     meta.RequestMethod,
		RequestPath:       meta.RequestPath,
		RoutePattern:      meta.RoutePattern,
		FeatureFullPath:   meta.FeatureFullPath,
		FeaturePathIDs:    meta.FeaturePathIDs,
		Timestamp:         meta.Timestamp,
		IngestedAt:        &indexedAt,
		ElasticIndex:      indexName,
		ElasticDocumentID: envelope.LogID,
		IndexStatus:       "INDEXED",
		IndexedAt:         &indexedAt,
		LastSyncAt:        &indexedAt,
	}

	return &processedLog{
		message:         envelope,
		natsMsg:         msg,
		document:        document,
		meta:            meta,
		indexName:       indexName,
		secrets:         secrets,
		objectRef:       objectRef,
		indexRef:        indexRef,
		originalPayload: originalPayload,
	}, nil
}

func canonicalDynamicPath(path string) string {
	path = strings.Trim(strings.TrimSpace(path), ".")
	if path == "$" {
		return "$"
	}

	// Field definitions can originate from the legacy payload, the current data
	// shape, or a raw sample. The worker always traverses the decoded document
	// root, so these transport prefixes must not participate in path matching.
	for {
		switch {
		case strings.HasPrefix(path, "raw."):
			path = strings.TrimPrefix(path, "raw.")
		case strings.HasPrefix(path, "payload."):
			path = strings.TrimPrefix(path, "payload.")
		case strings.HasPrefix(path, "data."):
			path = strings.TrimPrefix(path, "data.")
		case strings.HasPrefix(path, "fields."):
			path = strings.TrimPrefix(path, "fields.")
		default:
			if path == "raw" || path == "payload" || path == "data" || path == "fields" || path == "" {
				return "$"
			}
			return path
		}
	}
}

func readJSONPath(value any, path string) (any, bool) {
	if strings.TrimSpace(path) == "$" {
		return value, true
	}
	parts := strings.Split(strings.Trim(path, "."), ".")
	if len(parts) == 0 || parts[0] == "" {
		return nil, false
	}
	return readJSONPathParts(value, parts)
}

func readJSONPathParts(value any, parts []string) (any, bool) {
	if len(parts) == 0 {
		return value, true
	}
	if values, ok := value.([]any); ok {
		result := make([]any, 0, len(values))
		for _, item := range values {
			if child, found := readJSONPathParts(item, parts); found {
				result = append(result, child)
			}
		}
		return result, len(result) > 0
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, false
	}
	segment := parts[0]
	if strings.HasSuffix(segment, "[]") {
		array, ok := object[strings.TrimSuffix(segment, "[]")].([]any)
		if !ok {
			return nil, false
		}
		if len(parts) == 1 {
			return array, true
		}
		return readJSONPathParts(array, parts[1:])
	}
	child, ok := object[segment]
	if !ok {
		return nil, false
	}
	return readJSONPathParts(child, parts[1:])
}

func writeJSONPath(object map[string]any, path string, value any) error {
	if strings.TrimSpace(path) == "$" {
		return nil
	}
	return writeJSONPathParts(object, strings.Split(strings.Trim(path, "."), "."), value)
}

func writeJSONPathParts(object map[string]any, parts []string, value any) error {
	if len(parts) == 0 {
		return nil
	}
	segment := parts[0]
	if strings.HasSuffix(segment, "[]") {
		key := strings.TrimSuffix(segment, "[]")
		array, ok := object[key].([]any)
		if !ok {
			return fmt.Errorf("JSON path %s is not an array", strings.Join(parts, "."))
		}
		values, _ := value.([]any)
		for index, item := range array {
			child, ok := item.(map[string]any)
			if !ok || index >= len(values) {
				continue
			}
			if err := writeJSONPathParts(child, parts[1:], values[index]); err != nil {
				return err
			}
		}
		return nil
	}
	if len(parts) == 1 {
		object[segment] = value
		return nil
	}
	child, ok := object[segment].(map[string]any)
	if !ok {
		return fmt.Errorf("JSON path %s is not an object", strings.Join(parts, "."))
	}
	return writeJSONPathParts(child, parts[1:], value)
}

func (u *usecase) buildObjectStorageRef(envelope queue.LogMessage, payload []byte) (models.LogObjectStorageRef, error) {
	encryptedPayload, err := utils.EncryptAESGCM(string(payload), []byte(u.encryptionKey))
	if err != nil {
		return models.LogObjectStorageRef{}, err
	}

	sizeBytes := int64(len(payload))
	checksum := utils.SHA256Hex(string(payload))
	fileFormat := "JSON"
	keyRef := "DATA_ENCRYPTION_KEY"
	algorithm := "AES-256-GCM"
	objectPath := fmt.Sprintf("postgres://log-payloads/%s", envelope.LogID)

	return models.LogObjectStorageRef{
		ObjectRefID:         newUUID(),
		LogID:               envelope.LogID,
		ProductID:           *envelope.ProductID,
		EnvironmentID:       envelope.EnvironmentID,
		StorageProvider:     "POSTGRES",
		ObjectPath:          objectPath,
		ObjectType:          "INPUT_PAYLOAD",
		FileFormat:          &fileFormat,
		SizeBytes:           &sizeBytes,
		Checksum:            &checksum,
		EncryptedPayload:    &encryptedPayload,
		EncryptionKeyRef:    &keyRef,
		EncryptionAlgorithm: &algorithm,
		IsEncrypted:         true,
		RetentionUntil:      envelope.RetentionUntil,
	}, nil
}

func (u *usecase) failMessage(ctx context.Context, msg *nats.Msg, envelope queue.LogMessage, stage string, failureType string, reason string, retryable bool) error {
	attemptNo := 1
	if metadata, err := msg.Metadata(); err == nil && metadata.NumDelivered > 0 {
		attemptNo = int(metadata.NumDelivered)
	}
	maxRetryCount := envelope.MaxRetryCount
	if maxRetryCount <= 0 {
		maxRetryCount = 3
	}
	retryCount := attemptNo - 1
	status := "FAILED"
	var nextRetryAt *time.Time
	if retryable && retryCount < maxRetryCount {
		delay := time.Second << min(retryCount, 5)
		next := time.Now().UTC().Add(delay)
		nextRetryAt = &next
		status = "RETRYING"
	}
	reasonCopy := reason
	failure := models.LogFailure{
		FailureID:     newUUID(),
		LogID:         envelope.LogID,
		AttemptNo:     attemptNo,
		QueueItemID:   envelope.QueueItemID,
		BatchID:       envelope.BatchID,
		SequenceNo:    envelope.SequenceNo,
		ProductID:     envelope.ProductID,
		SourceID:      envelope.SourceID,
		EnvironmentID: envelope.EnvironmentID,
		FailureStage:  stage,
		FailureType:   failureType,
		Reason:        &reasonCopy,
		ErrorDetails:  envelope.InputPayload, // เก็บ Payload ดิบเอาไว้สำหรับนำไปทำ Retry
		RetryCount:    retryCount,
		MaxRetryCount: maxRetryCount,
		NextRetryAt:   nextRetryAt,
		LastRetryAt:   func() *time.Time { now := time.Now().UTC(); return &now }(),
		Status:        status,
	}

	if err := u.repo.CreateFailure(ctx, &failure); err != nil {
		return err
	}
	if status == "RETRYING" {
		return msg.NakWithDelay(time.Until(*nextRetryAt))
	}

	if err := msg.Ack(); err != nil {
		return err
	}
	return nil
}

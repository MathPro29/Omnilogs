package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime"
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
	matchers   sync.Map // productID -> *sensitiveMatchers
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

	u.markFetchedBatchesProcessing(ctx, msgs)

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

func (u *usecase) markFetchedBatchesProcessing(ctx context.Context, msgs []*nats.Msg) {
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
	}
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
		fields, err := u.repo.GetSensitiveFieldDefinitions(ctx, pid)
		if err != nil {
			return nil, fmt.Errorf("load sensitive field definitions for product %d: %w", pid, err)
		}
		cache.fields.Store(pid, fields)

		rules, err := u.repo.GetLogMaskingRules(ctx, pid)
		if err != nil {
			return nil, fmt.Errorf("load masking rules for product %d: %w", pid, err)
		}
		cache.rules.Store(pid, rules)
		cache.matchers.Store(pid, buildSensitiveMatchers(fields, rules))
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
	var matchers *sensitiveMatchers
	if val, ok := cache.matchers.Load(*envelope.ProductID); ok {
		matchers = val.(*sensitiveMatchers)
	}

	var payload map[string]any
	if err := json.Unmarshal(envelope.InputPayload, &payload); err != nil {
		return nil, u.failMessage(ctx, msg, envelope, "TRANSFORM", "INVALID_PAYLOAD", err.Error(), false)
	}
	if payload == nil {
		return nil, u.failMessage(ctx, msg, envelope, "TRANSFORM", "INVALID_PAYLOAD", "payload is null or empty", false)
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

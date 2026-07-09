package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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
		return err
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
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		prepared []processedLog
		firstErr error
	)

	cache := &batchCache{}
	uniqueProductIDs := make(map[int]struct{})

	for _, msg := range msgs {
		envelope, err := decodeMessage(msg)
		if err == nil && envelope.ProductID != nil {
			uniqueProductIDs[*envelope.ProductID] = struct{}{}
		}
	}

	for pid := range uniqueProductIDs {
		fields, _ := u.repo.GetSensitiveFieldDefinitions(ctx, pid)
		cache.fields.Store(pid, fields)

		rules, _ := u.repo.GetLogMaskingRules(ctx, pid)
		cache.rules.Store(pid, rules)
		cache.matchers.Store(pid, buildSensitiveMatchers(fields, rules))
	}

	for _, msg := range msgs {
		wg.Add(1)
		go func(msg *nats.Msg) {
			defer wg.Done()

			entry, err := u.prepareLog(ctx, msg, cache)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}
			if entry == nil {
				return
			}

			mu.Lock()
			prepared = append(prepared, *entry)
			mu.Unlock()
		}(msg)
	}

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
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
	_ = retryable
	reasonCopy := reason
	failure := models.LogFailure{
		FailureID:     newUUID(),
		LogID:         envelope.LogID,
		AttemptNo:     1,
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
		RetryCount:    0,
		MaxRetryCount: 0,
		NextRetryAt:   nil,
		LastRetryAt:   nil,
		Status:        "FAILED",
	}

	if err := u.repo.CreateFailure(ctx, &failure); err != nil {
		return err
	}

	if err := msg.Ack(); err != nil {
		return err
	}
	return nil
}

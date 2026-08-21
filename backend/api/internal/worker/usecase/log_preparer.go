package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"omnilogs-api/dto"
	"omnilogs-api/models"
	"omnilogs-api/utils"

	"github.com/nats-io/nats.go"
)

type batchCache struct {
	fields            sync.Map // productID -> []models.LogFieldDefinition
	rules             sync.Map // productID -> []models.LogMaskingRule
	hierarchy         sync.Map // string -> error (caching valid or invalid results)
	resolvedHierarchy sync.Map // product/code/ids -> resolved hierarchy metadata
	projectFeatures   sync.Map // product/project -> []models.ProjectFeature
	indexNames        sync.Map // productID-environmentID-date -> index name
}

type sensitiveMatchers struct {
	fieldByKey  map[string]*models.LogFieldDefinition
	fieldByPath map[string]*models.LogFieldDefinition
	ruleByKey   map[string]*models.LogMaskingRule
	ruleByPath  map[string]*models.LogMaskingRule
}

type processedLog struct {
	message         dto.LogMessage
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
	normalizeHTTPFields(payload)
	normalizeCustomHierarchy(payload)
	// Evaluate saved route mappings before resolving the hierarchy. A routing
	// rule supplies the canonical project/category IDs that resolveLogHierarchy
	// validates and expands for the indexed document.
	if err := u.resolveConfiguredRouting(ctx, envelope, payload); err != nil {
		return nil, u.failMessage(ctx, msg, envelope, "VALIDATION", "INVALID_ROUTING", err.Error(), false)
	}
	if err := u.resolveLogHierarchy(ctx, *envelope.ProductID, *envelope.EnvironmentID, payload, cache); err != nil {
		return nil, u.failMessage(ctx, msg, envelope, "VALIDATION", "INVALID_LOG_HIERARCHY", err.Error(), false)
	}
	if err := u.enrichMappingContext(ctx, envelope, payload); err != nil {
		return nil, u.failMessage(ctx, msg, envelope, "VALIDATION", "INVALID_MAPPING_CONTEXT", err.Error(), false)
	}
	_, rawMeta, err := buildElasticDocument(&envelope, payload)
	if err != nil {
		return nil, u.failMessage(ctx, msg, envelope, "TRANSFORM", "INVALID_PAYLOAD", err.Error(), false)
	}
	fields = applicableFieldDefinitions(fields, rawMeta)
	applyConfiguredCustomFields(payload, fields)
	sensitiveFields := make([]models.LogFieldDefinition, 0, len(fields))
	for _, field := range fields {
		if field.IsSensitive {
			sensitiveFields = append(sensitiveFields, field)
		}
	}
	matchers := buildSensitiveMatchers(sensitiveFields, rules)

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
		RouteKey:          meta.RouteKey,
		ServiceName:       meta.ServiceName,
		RoutingStatus:     routingStatus(meta.RoutingStatus),
		RoutingMethod:     meta.RoutingMethod,
		RoutingReason:     meta.RoutingReason,
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

func (u *usecase) buildObjectStorageRef(envelope dto.LogMessage, payload []byte) (models.LogObjectStorageRef, error) {
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

func routingStatus(value *string) string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return "UNCLASSIFIED"
	}
	return strings.TrimSpace(*value)
}

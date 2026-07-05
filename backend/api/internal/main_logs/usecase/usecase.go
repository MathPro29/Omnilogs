package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	auditusecase "omnilogs-api/internal/audit_logs/usecase"
	"omnilogs-api/models"
	"omnilogs-api/responses"
	"omnilogs-api/utils"

	"github.com/elastic/go-elasticsearch/v8"
	"gorm.io/gorm"
)

var ErrMainLogNotFound = errors.New("main log not found")

type SearchInput struct {
	ActorUserID    uint
	PlatformAdmin  bool
	ProductID      int64
	EnvironmentID  *int64
	ProjectID      *int64
	ProjectIDs     []int64
	CategoryID     *int64
	CategoryIDs    []int64
	Level          *string
	LogType        *string
	RequestID      *string
	TraceID        *string
	Keyword        *string
	Page           int
	PerPage        int
}

type MainLogDocument struct {
	LogID           string         `json:"log_id"`
	ProductID       int64          `json:"product_id"`
	EnvironmentID   *int64         `json:"environment_id,omitempty"`
	SourceID        *int64         `json:"source_id,omitempty"`
	Timestamp       *string        `json:"timestamp,omitempty"`
	Level           *string        `json:"level,omitempty"`
	LogType         *string        `json:"log_type,omitempty"`
	Message         *string        `json:"message,omitempty"`
	RequestID       *string        `json:"request_id,omitempty"`
	TraceID         *string        `json:"trace_id,omitempty"`
	Method          *string        `json:"method,omitempty"`
	Path            *string        `json:"path,omitempty"`
	URL             *string        `json:"url,omitempty"`
	StatusCode      *int64         `json:"status_code,omitempty"`
	LatencyMs       *int64         `json:"latency_ms,omitempty"`
	RequestHeaders  map[string]any `json:"request_headers,omitempty"`
	ResponseHeaders map[string]any `json:"response_headers,omitempty"`
	RequestPayload  any            `json:"request_payload,omitempty"`
	ResponsePayload any            `json:"response_payload,omitempty"`
	ErrorCode       *string        `json:"error_code,omitempty"`
	ErrorMessage    *string        `json:"error_message,omitempty"`
	StackTrace      *string        `json:"stack_trace,omitempty"`
	CustomFields    map[string]any `json:"custom_fields,omitempty"`
	Raw             map[string]any `json:"raw"`
}

type SearchResult struct {
	Items []MainLogDocument
	Total int64
}

type Usecase interface {
	Search(ctx context.Context, input SearchInput, requestID, traceID, ipAddress, userAgent *string) (*SearchResult, error)
	FindByID(ctx context.Context, actorUserID uint, platformAdmin bool, productID int64, logID string, requestID, traceID, ipAddress, userAgent *string) (*MainLogDocument, error)
	FindByAudit(ctx context.Context, actorUserID uint, platformAdmin bool, auditID string, requestID, traceID, ipAddress, userAgent *string) (*models.SystemAuditLog, *MainLogDocument, string, error)
}

type usecase struct {
	db           *gorm.DB
	esClient     *elasticsearch.Client
	auditUsecase auditusecase.Usecase
	encryptionKey string
}

func NewUsecase(db *gorm.DB, esClient *elasticsearch.Client, auditUsecase auditusecase.Usecase, encryptionKey string) Usecase {
	return &usecase{db: db, esClient: esClient, auditUsecase: auditUsecase, encryptionKey: encryptionKey}
}

func (u *usecase) Search(ctx context.Context, input SearchInput, requestID, traceID, ipAddress, userAgent *string) (*SearchResult, error) {
	if err := u.ensureProductAccess(ctx, input.ActorUserID, input.PlatformAdmin, input.ProductID); err != nil {
		_ = u.auditUsecase.Record(ctx, auditusecase.RecordAuditInput{
			ActorUserID: uintToInt64Ptr(input.ActorUserID),
			ProductID:   &input.ProductID,
			Action:      models.AuditActionSearchLogs,
			ResourceType: models.AuditResourceTypeMainLog,
			RequestID:   requestID,
			TraceID:     traceID,
			Result:      models.AuditResultDenied,
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
		})
		return nil, err
	}

	query := buildSearchQuery(input)
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, err
	}

	indices := u.resolveSearchIndices(input.ProductID)
	res, err := u.esClient.Search(
		u.esClient.Search.WithContext(ctx),
		u.esClient.Search.WithIndex(indices...),
		u.esClient.Search.WithBody(&buf),
		u.esClient.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch returned status %d", res.StatusCode)
	}

	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, err
	}

	result := &SearchResult{Items: []MainLogDocument{}}
	hits, _ := payload["hits"].(map[string]any)
	if totalMap, ok := hits["total"].(map[string]any); ok {
		if value, ok := totalMap["value"].(float64); ok {
			result.Total = int64(value)
		}
	}
	if hitList, ok := hits["hits"].([]any); ok {
		for _, item := range hitList {
			hitMap, ok := item.(map[string]any)
			if !ok {
				continue
			}
			source, _ := hitMap["_source"].(map[string]any)
			logID, _ := hitMap["_id"].(string)
			result.Items = append(result.Items, mapMainLog(logID, source))
		}
	}

	_ = u.auditUsecase.Record(ctx, auditusecase.RecordAuditInput{
		ActorUserID:  uintToInt64Ptr(input.ActorUserID),
		ProductID:    &input.ProductID,
		Action:       models.AuditActionSearchLogs,
		ResourceType: models.AuditResourceTypeMainLog,
		RequestID:    requestID,
		TraceID:      traceID,
		Result:       models.AuditResultSuccess,
		Metadata: map[string]any{
			"page":       input.Page,
			"per_page":   input.PerPage,
			"result_cnt": result.Total,
		},
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})

	return result, nil
}

func (u *usecase) FindByID(ctx context.Context, actorUserID uint, platformAdmin bool, productID int64, logID string, requestID, traceID, ipAddress, userAgent *string) (*MainLogDocument, error) {
	if err := u.ensureProductAccess(ctx, actorUserID, platformAdmin, productID); err != nil {
		_ = u.auditUsecase.Record(ctx, auditusecase.RecordAuditInput{
			ActorUserID:  uintToInt64Ptr(actorUserID),
			ProductID:    &productID,
			Action:       models.AuditActionViewLogDetail,
			ResourceType: models.AuditResourceTypeMainLog,
			ResourceID:   &logID,
			RequestID:    requestID,
			TraceID:      traceID,
			Result:       models.AuditResultDenied,
			IPAddress:    ipAddress,
			UserAgent:    userAgent,
		})
		return nil, err
	}

	value, err := u.findMainLogByID(ctx, productID, logID)
	if err != nil {
		result := models.AuditResultFailed
		if errors.Is(err, ErrMainLogNotFound) {
			result = models.AuditResultFailed
		}
		_ = u.auditUsecase.Record(ctx, auditusecase.RecordAuditInput{
			ActorUserID:  uintToInt64Ptr(actorUserID),
			ProductID:    &productID,
			Action:       models.AuditActionViewLogDetail,
			ResourceType: models.AuditResourceTypeMainLog,
			ResourceID:   &logID,
			RequestID:    requestID,
			TraceID:      traceID,
			Result:       result,
			IPAddress:    ipAddress,
			UserAgent:    userAgent,
		})
		return nil, err
	}

	_ = u.auditUsecase.Record(ctx, auditusecase.RecordAuditInput{
		ActorUserID:  uintToInt64Ptr(actorUserID),
		ProductID:    &productID,
		Action:       models.AuditActionViewLogDetail,
		ResourceType: models.AuditResourceTypeMainLog,
		ResourceID:   &logID,
		RequestID:    requestID,
		TraceID:      traceID,
		Result:       models.AuditResultSuccess,
		Metadata: map[string]any{
			"view_mode": "DETAIL",
		},
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})

	return value, nil
}

func (u *usecase) FindByAudit(ctx context.Context, actorUserID uint, platformAdmin bool, auditID string, requestID, traceID, ipAddress, userAgent *string) (*models.SystemAuditLog, *MainLogDocument, string, error) {
	audit, err := u.auditUsecase.FindByID(ctx, auditID)
	if err != nil {
		return nil, nil, "", err
	}
	if audit.ProductID == nil || audit.ResourceID == nil || audit.ResourceType != models.AuditResourceTypeMainLog {
		return audit, nil, "NOT_FOUND", nil
	}
	value, err := u.FindByID(ctx, actorUserID, platformAdmin, *audit.ProductID, *audit.ResourceID, requestID, traceID, ipAddress, userAgent)
	if err != nil {
		if errors.Is(err, ErrMainLogNotFound) {
			return audit, nil, "NOT_FOUND", nil
		}
		return audit, nil, "", err
	}
	return audit, value, "", nil
}

func (u *usecase) findMainLogByID(ctx context.Context, productID int64, logID string) (*MainLogDocument, error) {
	var ref models.LogIndexRef
	if err := u.db.WithContext(ctx).Where("log_id = ? AND product_id = ?", logID, productID).First(&ref).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMainLogNotFound
		}
		return nil, err
	}

	res, err := u.esClient.Get(ref.ElasticIndex, ref.ElasticDocumentID, u.esClient.Get.WithContext(ctx))
	if err != nil {
		return u.findMainLogFromPostgresPayload(ctx, &ref)
	}
	defer res.Body.Close()
	if res.StatusCode == 404 {
		return u.findMainLogFromPostgresPayload(ctx, &ref)
	}
	if res.IsError() {
		return u.findMainLogFromPostgresPayload(ctx, &ref)
	}

	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, err
	}
	source, _ := payload["_source"].(map[string]any)
	value := mapMainLog(ref.ElasticDocumentID, source)
	return &value, nil
}

func (u *usecase) findMainLogFromPostgresPayload(ctx context.Context, ref *models.LogIndexRef) (*MainLogDocument, error) {
	var objectRef models.LogObjectStorageRef
	if err := u.db.WithContext(ctx).
		Where("log_id = ? AND object_type = ?", ref.LogID, "INPUT_PAYLOAD").
		Order("created_at DESC NULLS LAST").
		First(&objectRef).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMainLogNotFound
		}
		return nil, err
	}

	if objectRef.EncryptedPayload == nil || strings.TrimSpace(*objectRef.EncryptedPayload) == "" {
		return nil, ErrMainLogNotFound
	}

	decryptedPayload, err := utils.DecryptAESGCM(*objectRef.EncryptedPayload, []byte(u.encryptionKey))
	if err != nil {
		return nil, err
	}

	var payload any
	if err := json.Unmarshal([]byte(decryptedPayload), &payload); err != nil {
		return nil, err
	}

	source := map[string]any{
		"product_id":     ref.ProductID,
		"environment_id": ref.EnvironmentID,
		"@timestamp":     ref.Timestamp.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
		"payload":        payload,
		"restored_from":  "postgres_payload",
	}
	if ref.SourceID != nil {
		source["source_id"] = *ref.SourceID
	}

	value := mapMainLog(ref.LogID, source)
	return &value, nil
}

func (u *usecase) ensureProductAccess(ctx context.Context, actorUserID uint, platformAdmin bool, productID int64) error {
	if platformAdmin {
		return nil
	}
	var count int64
	if err := u.db.WithContext(ctx).
		Model(&models.ProductMembership{}).
		Where("user_id = ? AND product_id = ? AND is_active = TRUE", actorUserID, productID).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return responses.ErrForbidden
	}
	return nil
}

func buildSearchQuery(input SearchInput) map[string]any {
	filters := []map[string]any{
		{"term": map[string]any{"product_id": input.ProductID}},
	}
	if input.EnvironmentID != nil {
		filters = append(filters, map[string]any{"term": map[string]any{"environment_id": *input.EnvironmentID}})
	}
	if input.ProjectID != nil {
		filters = append(filters, map[string]any{"term": map[string]any{"payload.project_id": *input.ProjectID}})
	}
	if input.CategoryID != nil {
		categoryID := fmt.Sprintf("%d", *input.CategoryID)
		filters = append(filters, map[string]any{
			"bool": map[string]any{
				"should": []map[string]any{
					{"term": map[string]any{"payload.category_id": *input.CategoryID}},
					{"term": map[string]any{"payload.feature_path_ids.keyword": categoryID}},
					{"wildcard": map[string]any{"payload.feature_path_ids.keyword": categoryID + ",*"}},
					{"wildcard": map[string]any{"payload.feature_path_ids.keyword": "*," + categoryID}},
					{"wildcard": map[string]any{"payload.feature_path_ids.keyword": "*," + categoryID + ",*"}},
				},
				"minimum_should_match": 1,
			},
		})
	}
	// Multi-project filter
	if len(input.ProjectIDs) > 0 && input.ProjectID == nil {
		filters = append(filters, map[string]any{"terms": map[string]any{"payload.project_id": input.ProjectIDs}})
	}
	// Multi-category filter
	if len(input.CategoryIDs) > 0 && input.CategoryID == nil {
		shouldClauses := make([]map[string]any, 0)
		for _, catID := range input.CategoryIDs {
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
	if input.Level != nil && strings.TrimSpace(*input.Level) != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"payload.log_level.keyword": strings.TrimSpace(*input.Level)}})
	}
	if input.LogType != nil && strings.TrimSpace(*input.LogType) != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"payload.event_type.keyword": strings.TrimSpace(*input.LogType)}})
	}
	if input.RequestID != nil && strings.TrimSpace(*input.RequestID) != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"payload.source_request_id.keyword": strings.TrimSpace(*input.RequestID)}})
	}
	if input.TraceID != nil && strings.TrimSpace(*input.TraceID) != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"payload.trace_id.keyword": strings.TrimSpace(*input.TraceID)}})
	}

	must := []map[string]any{}
	if input.Keyword != nil && strings.TrimSpace(*input.Keyword) != "" {
		must = append(must, map[string]any{
			"multi_match": map[string]any{
				"query": strings.TrimSpace(*input.Keyword),
				"fields": []string{
					"payload.message",
					"payload.error_message",
					"payload.request_path",
					"payload.route_pattern",
				},
			},
		})
	}

	return map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"filter": filters,
				"must":   must,
			},
		},
		"sort": []any{
			map[string]any{"@timestamp": map[string]any{"order": "desc"}},
		},
		"from": (input.Page - 1) * input.PerPage,
		"size": input.PerPage,
	}
}

func mapMainLog(logID string, source map[string]any) MainLogDocument {
	value := MainLogDocument{
		LogID: logID,
		Raw:   source,
	}
	if source == nil {
		return value
	}
	value.ProductID = int64FromMap(source, "product_id")
	value.EnvironmentID = optionalInt64FromMap(source, "environment_id")
	value.SourceID = optionalInt64FromMap(source, "source_id")
	value.Timestamp = stringPtrFromMap(source, "@timestamp")
	value.Level = stringPtrFromPayload(source, "log_level")
	value.LogType = stringPtrFromPayload(source, "event_type")
	value.Message = stringPtrFromPayload(source, "message")
	value.RequestID = stringPtrFromPayload(source, "source_request_id")
	value.TraceID = stringPtrFromPayload(source, "trace_id")
	value.Method = stringPtrFromPayload(source, "request_method")
	value.Path = stringPtrFromPayload(source, "request_path")
	value.URL = stringPtrFromPayload(source, "url")
	value.StatusCode = optionalInt64FromPayload(source, "status_code")
	value.LatencyMs = optionalInt64FromPayload(source, "duration_ms")
	value.ErrorCode = stringPtrFromPayload(source, "error_code")
	value.ErrorMessage = stringPtrFromPayload(source, "error_message")
	value.StackTrace = stringPtrFromPayload(source, "stack_trace")
	if payload, ok := source["payload"].(map[string]any); ok {
		value.RequestHeaders = mapFromValue(payload["request_headers"])
		value.ResponseHeaders = mapFromValue(payload["response_headers"])
		value.RequestPayload = payload["request_payload"]
		value.ResponsePayload = payload["response_payload"]
		value.CustomFields = mapFromValue(payload["custom_fields"])
	}
	return value
}

func uintToInt64Ptr(v uint) *int64 {
	result := int64(v)
	return &result
}

func mapFromValue(value any) map[string]any {
	result, _ := value.(map[string]any)
	return result
}

func int64FromMap(source map[string]any, key string) int64 {
	if value, ok := source[key].(float64); ok {
		return int64(value)
	}
	return 0
}

func optionalInt64FromMap(source map[string]any, key string) *int64 {
	if value, ok := source[key].(float64); ok {
		result := int64(value)
		return &result
	}
	return nil
}

func stringPtrFromMap(source map[string]any, key string) *string {
	value, ok := source[key].(string)
	if !ok || strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func stringPtrFromPayload(source map[string]any, key string) *string {
	payload, ok := source["payload"].(map[string]any)
	if !ok {
		return nil
	}
	return stringPtrFromMap(payload, key)
}

func optionalInt64FromPayload(source map[string]any, key string) *int64 {
	payload, ok := source["payload"].(map[string]any)
	if !ok {
		return nil
	}
	return optionalInt64FromMap(payload, key)
}

func (u *usecase) resolveSearchIndices(productID int64) []string {
	indices := []string{fmt.Sprintf("omnilogs-product-%d-*", productID)}
	var policies []models.ElasticIndexPolicy
	if err := u.db.Where("product_id = ? AND is_active = TRUE", productID).Find(&policies).Error; err == nil {
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

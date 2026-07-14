package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	auditusecase "omnilogs-api/internal/audit_logs/usecase"
	"omnilogs-api/internal/main_logs/repository"
	"omnilogs-api/models"
	"omnilogs-api/responses"
	"omnilogs-api/utils"

	"gorm.io/gorm"
)

var ErrMainLogNotFound = errors.New("main log not found")
var ErrInvalidSearchFilter = errors.New("invalid log search filter")

type SearchInput struct {
	ActorUserID      uint
	PlatformAdmin    bool
	ProductID        int64
	EnvironmentID    *int64
	ProjectID        *int64
	ProjectIDs       []int64
	CategoryID       *int64
	CategoryIDs      []int64
	Level            *string
	LogType          *string
	RequestID        *string
	TraceID          *string
	CustomFieldPath  *string
	CustomFieldValue *string
	Keyword          *string
	Page             int
	PerPage          int
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
	repo          repository.Repository
	auditUsecase  auditusecase.Usecase
	encryptionKey string
}

func NewUsecase(repo repository.Repository, auditUsecase auditusecase.Usecase, encryptionKey string) Usecase {
	return &usecase{repo: repo, auditUsecase: auditUsecase, encryptionKey: encryptionKey}
}

func (u *usecase) Search(ctx context.Context, input SearchInput, requestID, traceID, ipAddress, userAgent *string) (*SearchResult, error) {
	input = normalizeSearchInput(input)
	if err := validateSearchInput(input); err != nil {
		return nil, err
	}
	if err := u.ensureProductAccess(ctx, input.ActorUserID, input.PlatformAdmin, input.ProductID); err != nil {
		_ = u.auditUsecase.Record(ctx, auditusecase.RecordAuditInput{
			ActorUserID:  uintToInt64Ptr(input.ActorUserID),
			ProductID:    &input.ProductID,
			Action:       models.AuditActionSearchLogs,
			ResourceType: models.AuditResourceTypeMainLog,
			RequestID:    requestID,
			TraceID:      traceID,
			Result:       models.AuditResultDenied,
			IPAddress:    ipAddress,
			UserAgent:    userAgent,
		})
		return nil, err
	}

	query := buildSearchQuery(input)
	queryBody, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	indices := u.resolveSearchIndices(ctx, input.ProductID)
	res, err := u.repo.SearchLogs(ctx, indices, queryBody)
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
	if timedOut, _ := payload["timed_out"].(bool); timedOut {
		return nil, context.DeadlineExceeded
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
	ref, err := u.repo.GetLogIndexRef(ctx, productID, logID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMainLogNotFound
		}
		return nil, err
	}

	res, err := u.repo.GetMainLogFromES(ctx, ref.ElasticIndex, ref.ElasticDocumentID)
	if err != nil {
		return u.findMainLogFromPostgresPayload(ctx, ref)
	}
	defer res.Body.Close()
	if res.StatusCode == 404 {
		return u.findMainLogFromPostgresPayload(ctx, ref)
	}
	if res.IsError() {
		return u.findMainLogFromPostgresPayload(ctx, ref)
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
	objectRef, err := u.repo.GetPostgresPayload(ctx, ref.LogID)
	if err != nil {
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
	hasMembership, err := u.repo.HasProductMembership(ctx, actorUserID, productID)
	if err != nil {
		return err
	}
	if !hasMembership {
		return responses.ErrForbidden
	}
	return nil
}

func (u *usecase) resolveSearchIndices(ctx context.Context, productID int64) []string {
	indices := []string{fmt.Sprintf("omnilogs-product-%d-*", productID)}
	seen := map[string]struct{}{indices[0]: {}}
	if policies, err := u.repo.GetActiveIndexPolicies(ctx, productID); err == nil {
		for _, policy := range policies {
			if strings.TrimSpace(policy.IndexPrefix) != "" {
				prefix := strings.ToLower(strings.TrimSpace(policy.IndexPrefix))
				prefix = strings.ReplaceAll(prefix, "_", "-")
				prefix = strings.ReplaceAll(prefix, " ", "-")
				pattern := fmt.Sprintf("%s-*", prefix)
				if _, exists := seen[pattern]; !exists {
					seen[pattern] = struct{}{}
					indices = append(indices, pattern)
				}
			}
		}
	}
	return indices
}

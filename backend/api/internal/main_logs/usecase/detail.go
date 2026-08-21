package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	auditusecase "omnilogs-api/internal/audit_logs/usecase"
	"omnilogs-api/internal/main_logs/document"
	"omnilogs-api/models"
	"omnilogs-api/utils"

	"gorm.io/gorm"
)

func (u *usecase) FindByID(ctx context.Context, actorUserID uint, platformAdmin bool, productID int64, logID string, requestID, traceID, ipAddress, userAgent *string) (*MainLogDocument, error) {
	if err := u.ensureProductAccess(ctx, actorUserID, platformAdmin, productID); err != nil {
		u.recordDetailAudit(ctx, actorUserID, productID, logID, models.AuditResultDenied, requestID, traceID, ipAddress, userAgent)
		return nil, err
	}

	value, err := u.findMainLogByID(ctx, productID, logID)
	if err != nil {
		u.recordDetailAudit(ctx, actorUserID, productID, logID, models.AuditResultFailed, requestID, traceID, ipAddress, userAgent)
		return nil, err
	}
	u.recordDetailAudit(ctx, actorUserID, productID, logID, models.AuditResultSuccess, requestID, traceID, ipAddress, userAgent)
	return value, nil
}

func (u *usecase) recordDetailAudit(ctx context.Context, actorUserID uint, productID int64, logID, result string, requestID, traceID, ipAddress, userAgent *string) {
	input := auditusecase.RecordAuditInput{
		ActorUserID: uintToInt64Ptr(actorUserID), ProductID: &productID,
		Action: models.AuditActionViewLogDetail, ResourceType: models.AuditResourceTypeMainLog,
		ResourceID: &logID, RequestID: requestID, TraceID: traceID, Result: result,
		IPAddress: ipAddress, UserAgent: userAgent,
	}
	if result == models.AuditResultSuccess {
		input.Metadata = map[string]any{"view_mode": "DETAIL"}
	}
	_ = u.auditUsecase.Record(ctx, input)
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
	if errors.Is(err, ErrMainLogNotFound) {
		return audit, nil, "NOT_FOUND", nil
	}
	if err != nil {
		return audit, nil, "", err
	}
	return audit, value, "", nil
}

func (u *usecase) findMainLogByID(ctx context.Context, productID int64, logID string) (*MainLogDocument, error) {
	ref, err := u.repo.GetLogIndexRef(ctx, productID, logID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMainLogNotFound
	}
	if err != nil {
		return nil, err
	}

	response, err := u.repo.GetMainLogFromES(ctx, ref.ElasticIndex, ref.ElasticDocumentID)
	if err != nil {
		return u.findMainLogFromPostgresPayload(ctx, ref)
	}
	defer response.Body.Close()
	if response.IsError() {
		return u.findMainLogFromPostgresPayload(ctx, ref)
	}
	var payload struct {
		Source map[string]any `json:"_source"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}
	value := document.Map(ref.ElasticDocumentID, payload.Source)
	return &value, nil
}

func (u *usecase) findMainLogFromPostgresPayload(ctx context.Context, ref *models.LogIndexRef) (*MainLogDocument, error) {
	objectRef, err := u.repo.GetPostgresPayload(ctx, ref.LogID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMainLogNotFound
	}
	if err != nil {
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
		"product_id": ref.ProductID, "environment_id": ref.EnvironmentID,
		"@timestamp": ref.Timestamp.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
		"payload":    payload, "restored_from": "postgres_payload",
	}
	if ref.SourceID != nil {
		source["source_id"] = *ref.SourceID
	}
	value := document.Map(ref.LogID, source)
	return &value, nil
}

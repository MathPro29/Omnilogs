package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"

	"omnilogs-api/dto"
	"omnilogs-api/internal/audit_logs/repository"
	"omnilogs-api/models"
)

var (
	ErrInvalidAction       = errors.New("invalid audit action")
	ErrInvalidResourceType = errors.New("invalid audit resource type")
	ErrInvalidResult       = errors.New("invalid audit result")
)

type RecordAuditInput struct {
	ActorUserID  *int64
	ProductID    *int64
	Action       string
	ResourceType string
	ResourceID   *string
	RequestID    *string
	TraceID      *string
	Method       *string
	Path         *string
	Result       string
	Metadata     map[string]any
	IPAddress    *string
	UserAgent    *string
}

type Usecase interface {
	List(ctx context.Context, filter dto.AuditLogFilterRequest) ([]models.SystemAuditLog, int64, error)
	FindByID(ctx context.Context, auditID string) (*models.SystemAuditLog, error)
	Record(ctx context.Context, input RecordAuditInput) error
}

type usecase struct {
	repo repository.Repository
}

func NewUsecase(repo repository.Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) List(ctx context.Context, filter dto.AuditLogFilterRequest) ([]models.SystemAuditLog, int64, error) {
	return u.repo.List(ctx, filter)
}

func (u *usecase) FindByID(ctx context.Context, auditID string) (*models.SystemAuditLog, error) {
	return u.repo.FindByID(ctx, auditID)
}

func (u *usecase) Record(ctx context.Context, input RecordAuditInput) error {
	action := strings.ToUpper(strings.TrimSpace(input.Action))
	resourceType := strings.ToUpper(strings.TrimSpace(input.ResourceType))
	result := strings.ToUpper(strings.TrimSpace(input.Result))

	if !isAllowedAction(action) {
		return ErrInvalidAction
	}
	if resourceType != models.AuditResourceTypeMainLog {
		return ErrInvalidResourceType
	}
	if !isAllowedResult(result) {
		return ErrInvalidResult
	}

	metadataJSON, err := json.Marshal(redactMetadata(input.Metadata))
	if err != nil {
		metadataJSON = []byte("{}")
	}

	record := &models.SystemAuditLog{
		AuditID:      newAuditUUID(),
		ActorUserID:  input.ActorUserID,
		ProductID:    input.ProductID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   input.ResourceID,
		RequestID:    input.RequestID,
		TraceID:      input.TraceID,
		Method:       input.Method,
		Path:         input.Path,
		Result:       result,
		Metadata:     metadataJSON,
		IPAddress:    input.IPAddress,
		UserAgent:    input.UserAgent,
	}

	if err := u.repo.Create(ctx, record); err != nil {
		slog.Error("record audit log failed", "action", action, "resource_type", resourceType, "error", err)
		return err
	}
	return nil
}

func redactMetadata(metadata map[string]any) map[string]any {
	if len(metadata) == 0 {
		return map[string]any{}
	}
	cloned := make(map[string]any, len(metadata))
	for key, value := range metadata {
		if looksSensitive(key) {
			cloned[key] = "[REDACTED]"
			continue
		}
		cloned[key] = redactValue(value)
	}
	return cloned
}

func redactValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return redactMetadata(typed)
	case []any:
		result := make([]any, len(typed))
		for i := range typed {
			result[i] = redactValue(typed[i])
		}
		return result
	default:
		return value
	}
}

func looksSensitive(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	for _, token := range []string{"password", "token", "secret", "cookie", "apikey", "api_key", "authorization"} {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}

func isAllowedAction(value string) bool {
	switch value {
	case models.AuditActionViewLogDetail, models.AuditActionSearchLogs, models.AuditActionExportLogs, models.AuditActionViewSensitiveField:
		return true
	default:
		return false
	}
}

func isAllowedResult(value string) bool {
	switch value {
	case models.AuditResultSuccess, models.AuditResultFailed, models.AuditResultDenied:
		return true
	default:
		return false
	}
}

func newAuditUUID() string {
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

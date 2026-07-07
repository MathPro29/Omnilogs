package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"omnilogs-api/dto"
	"omnilogs-api/internal/audit_secret/repository"
	"omnilogs-api/models"
	"omnilogs-api/responses"
	"omnilogs-api/utils"
)

type Actor struct {
	UserID        int
	PlatformAdmin bool
}

type Usecase interface {
	CreateRequest(ctx context.Context, actor Actor, auditID string, req dto.CreateAuditSecretAccessRequest) (*models.AuditSecretAccessRequest, error)
	ReviewRequest(ctx context.Context, actor Actor, auditID string, requestID string, req dto.ReviewAuditSecretAccessRequest) (*models.AuditSecretAccessRequest, error)
	ListRequests(ctx context.Context, auditID string) ([]models.AuditSecretAccessRequest, error)
	RevealValue(ctx context.Context, actor Actor, auditID string, req dto.RevealAuditSecretRequest) (*dto.RevealedAuditSecretResponse, error)
	ListHistory(ctx context.Context, auditID string) ([]models.AuditSecretAccessHistory, error)
}

type usecase struct {
	repo       repository.Repository
	encryptKey []byte
}

func NewUsecase(repo repository.Repository, encryptionKey string) Usecase {
	return &usecase{repo: repo, encryptKey: []byte(encryptionKey)}
}

func (u *usecase) CreateRequest(ctx context.Context, actor Actor, auditID string, req dto.CreateAuditSecretAccessRequest) (*models.AuditSecretAccessRequest, error) {
	if _, err := u.repo.GetAudit(ctx, auditID); err != nil {
		return nil, responses.ErrNotFound
	}
	secret, err := u.repo.GetSecret(ctx, auditID, req.SecretID)
	if err != nil {
		return nil, responses.ErrNotFound
	}
	reason := req.Reason
	value := &models.AuditSecretAccessRequest{
		RequestID:      newUUID(),
		AuditID:        auditID,
		SecretID:       secret.SecretID,
		UserID:         actor.UserID,
		Reason:         &reason,
		ApprovalStatus: "PENDING",
	}
	if err := u.repo.CreateRequest(ctx, value); err != nil {
		return nil, err
	}
	return value, nil
}

func (u *usecase) ReviewRequest(ctx context.Context, actor Actor, auditID string, requestID string, req dto.ReviewAuditSecretAccessRequest) (*models.AuditSecretAccessRequest, error) {
	if !actor.PlatformAdmin {
		return nil, responses.ErrForbidden
	}
	value, err := u.repo.GetRequest(ctx, requestID)
	if err != nil || value.AuditID != auditID {
		return nil, responses.ErrNotFound
	}
	if value.ApprovalStatus != "PENDING" {
		return nil, errors.New("request has already been reviewed")
	}
	now := time.Now()
	value.ApprovalStatus = req.ApprovalStatus
	value.ApprovedBy = &actor.UserID
	value.ApprovedAt = &now
	if req.ExpiresAt != nil {
		value.ExpiresAt = req.ExpiresAt
	} else if req.ApprovalStatus == "APPROVED" {
		exp := now.Add(1 * time.Hour)
		value.ExpiresAt = &exp
	}
	if err := u.repo.UpdateRequest(ctx, value); err != nil {
		return nil, err
	}
	return value, nil
}

func (u *usecase) ListRequests(ctx context.Context, auditID string) ([]models.AuditSecretAccessRequest, error) {
	return u.repo.ListRequests(ctx, auditID)
}

func (u *usecase) RevealValue(ctx context.Context, actor Actor, auditID string, req dto.RevealAuditSecretRequest) (*dto.RevealedAuditSecretResponse, error) {
	accessReq, err := u.repo.GetRequest(ctx, req.RequestID)
	if err != nil || accessReq.AuditID != auditID {
		return nil, responses.ErrNotFound
	}
	if accessReq.UserID != actor.UserID {
		return nil, responses.ErrForbidden
	}
	if accessReq.ApprovalStatus != "APPROVED" {
		return nil, errors.New("request is not approved")
	}
	if accessReq.ExpiresAt != nil && time.Now().After(*accessReq.ExpiresAt) {
		return nil, errors.New("request approval has expired")
	}
	user, err := u.repo.GetUser(ctx, actor.UserID)
	if err != nil {
		return nil, responses.ErrNotFound
	}
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		failedStatus := "FAILED_PASSWORD"
		_ = u.repo.CreateHistory(ctx, &models.AuditSecretAccessHistory{
			AccessID:     newUUID(),
			RequestID:    &req.RequestID,
			AuditID:      auditID,
			SecretID:     accessReq.SecretID,
			UserID:       actor.UserID,
			AccessReason: accessReq.Reason,
			AccessStatus: &failedStatus,
		})
		return nil, errors.New("invalid password")
	}
	secret, err := u.repo.GetSecret(ctx, auditID, accessReq.SecretID)
	if err != nil {
		return nil, responses.ErrNotFound
	}
	decryptedValue, err := utils.DecryptAESGCM(secret.EncryptedValue, u.encryptKey)
	if err != nil {
		return nil, errors.New("failed to decrypt value")
	}
	successStatus := "SUCCESS"
	authMethod := "PASSWORD_RECONFIRM"
	_ = u.repo.CreateHistory(ctx, &models.AuditSecretAccessHistory{
		AccessID:     newUUID(),
		RequestID:    &req.RequestID,
		AuditID:      auditID,
		SecretID:     secret.SecretID,
		UserID:       actor.UserID,
		AccessReason: accessReq.Reason,
		AuthMethod:   &authMethod,
		AccessStatus: &successStatus,
	})
	return &dto.RevealedAuditSecretResponse{
		SecretID:      secret.SecretID,
		AuditID:       auditID,
		FieldKey:      secret.FieldKey,
		FieldPath:     secret.FieldPath,
		SourceSection: secret.SourceSection,
		Value:         decryptedValue,
		ExpiresAt:     accessReq.ExpiresAt,
	}, nil
}

func (u *usecase) ListHistory(ctx context.Context, auditID string) ([]models.AuditSecretAccessHistory, error) {
	return u.repo.ListHistory(ctx, auditID)
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

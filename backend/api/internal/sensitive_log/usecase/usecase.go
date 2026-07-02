package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"omnilogs-api/dto"
	"omnilogs-api/internal/sensitive_log/repository"
	"omnilogs-api/models"
	"omnilogs-api/responses"
	"omnilogs-api/utils"
)

type Actor struct {
	UserID        int
	PlatformAdmin bool
}

type Usecase interface {
	CreateRequest(ctx context.Context, actor Actor, req dto.CreateSensitiveLogAccessRequest) (*models.SensitiveLogAccessRequest, error)
	ReviewRequest(ctx context.Context, actor Actor, requestID string, req dto.ReviewSensitiveLogAccessRequest) (*models.SensitiveLogAccessRequest, error)
	ListRequests(ctx context.Context, actor Actor, productID int) ([]models.SensitiveLogAccessRequest, error)
	RevealValue(ctx context.Context, actor Actor, req dto.RevealSensitiveLogRequest) (*dto.RevealedSensitiveLogResponse, error)
	ListHistory(ctx context.Context, actor Actor, productID int) ([]models.SensitiveLogAccessHistory, error)
}

type usecase struct {
	repo       repository.Repository
	encryptKey []byte
}

func NewUsecase(repo repository.Repository, encryptionKey string) Usecase {
	return &usecase{
		repo:       repo,
		encryptKey: []byte(encryptionKey),
	}
}

func (u *usecase) CreateRequest(ctx context.Context, actor Actor, req dto.CreateSensitiveLogAccessRequest) (*models.SensitiveLogAccessRequest, error) {
	requestID := newUUID()
	
	accessReq := &models.SensitiveLogAccessRequest{
		RequestID:         requestID,
		UserID:            actor.UserID,
		ProductID:         req.ProductID,
		LogID:             req.LogID,
		FieldDefinitionID: req.FieldDefinitionID,
		SecretID:          req.SecretID,
		FieldPath:         req.FieldPath,
		Reason:            &req.Reason,
		ApprovalStatus:    "PENDING",
	}

	if err := u.repo.CreateRequest(ctx, accessReq); err != nil {
		return nil, err
	}
	return accessReq, nil
}

func (u *usecase) ReviewRequest(ctx context.Context, actor Actor, requestID string, req dto.ReviewSensitiveLogAccessRequest) (*models.SensitiveLogAccessRequest, error) {
	// 1. MUST be superadmin, owner, or god
	if !actor.PlatformAdmin {
		return nil, responses.ErrForbidden
	}

	accessReq, err := u.repo.GetRequest(ctx, requestID)
	if err != nil {
		return nil, responses.ErrNotFound
	}

	if accessReq.ApprovalStatus != "PENDING" {
		return nil, errors.New("request has already been reviewed")
	}

	now := time.Now()
	accessReq.ApprovalStatus = req.ApprovalStatus
	accessReq.ApprovedBy = &actor.UserID
	accessReq.ApprovedAt = &now
	
	if req.ExpiresAt != nil {
		accessReq.ExpiresAt = req.ExpiresAt
	} else if req.ApprovalStatus == "APPROVED" {
		// Default to 1 hour expiration
		exp := now.Add(1 * time.Hour)
		accessReq.ExpiresAt = &exp
	}

	if err := u.repo.UpdateRequest(ctx, accessReq); err != nil {
		return nil, err
	}
	return accessReq, nil
}

func (u *usecase) ListRequests(ctx context.Context, actor Actor, productID int) ([]models.SensitiveLogAccessRequest, error) {
	return u.repo.ListRequests(ctx, productID)
}

func (u *usecase) RevealValue(ctx context.Context, actor Actor, req dto.RevealSensitiveLogRequest) (*dto.RevealedSensitiveLogResponse, error) {
	// 1. Verify access request
	accessReq, err := u.repo.GetRequest(ctx, req.RequestID)
	if err != nil {
		return nil, responses.ErrNotFound
	}

	// 2. Requester verification (only original requester can reveal)
	if accessReq.UserID != actor.UserID {
		return nil, responses.ErrForbidden
	}

	// 3. Status checks
	if accessReq.ApprovalStatus != "APPROVED" {
		return nil, errors.New("request is not approved")
	}
	if accessReq.ExpiresAt != nil && time.Now().After(*accessReq.ExpiresAt) {
		return nil, errors.New("request approval has expired")
	}

	// 4. Verify Password for high-security double confirmation
	user, err := u.repo.GetUser(ctx, actor.UserID)
	if err != nil {
		return nil, responses.ErrNotFound
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		// Record failed access attempt
		failedStatus := "FAILED_PASSWORD"
		_ = u.repo.CreateHistory(ctx, &models.SensitiveLogAccessHistory{
			AccessID:     newUUID(),
			RequestID:    &req.RequestID,
			ProductID:    &accessReq.ProductID,
			UserID:       actor.UserID,
			LogID:        accessReq.LogID,
			FieldPath:    accessReq.FieldPath,
			AccessStatus: &failedStatus,
		})
		return nil, errors.New("invalid password")
	}

	// 5. Get and decrypt secret
	var secret *models.LogSensitiveFieldSecret
	if accessReq.SecretID != nil {
		secret, err = u.repo.GetSecret(ctx, *accessReq.SecretID)
	} else if accessReq.LogID != nil && accessReq.FieldPath != nil {
		secret, err = u.repo.GetSecretByLogAndPath(ctx, *accessReq.LogID, *accessReq.FieldPath)
	}
	if err != nil {
		return nil, responses.ErrNotFound
	}

	decryptedValue, err := utils.DecryptAESGCM(secret.EncryptedValue, u.encryptKey)
	if err != nil {
		return nil, errors.New("failed to decrypt value")
	}

	// 6. Record successful access
	successStatus := "SUCCESS"
	authMethod := "PASSWORD_RECONFIRM"
	accessHistory := &models.SensitiveLogAccessHistory{
		AccessID:     newUUID(),
		RequestID:    &req.RequestID,
		ProductID:    &accessReq.ProductID,
		UserID:       actor.UserID,
		LogID:        accessReq.LogID,
		FieldPath:    accessReq.FieldPath,
		AuthMethod:   &authMethod,
		AccessStatus: &successStatus,
	}
	_ = u.repo.CreateHistory(ctx, accessHistory)

	return &dto.RevealedSensitiveLogResponse{
		SecretID:  secret.SecretID,
		FieldKey:  secret.FieldKey,
		FieldPath: secret.FieldPath,
		Value:     decryptedValue,
		ExpiresAt: accessReq.ExpiresAt,
	}, nil
}

func (u *usecase) ListHistory(ctx context.Context, actor Actor, productID int) ([]models.SensitiveLogAccessHistory, error) {
	if !actor.PlatformAdmin {
		return nil, responses.ErrForbidden
	}
	return u.repo.ListHistory(ctx, productID)
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

package usecase

import (
	"context"
	"errors"

	auditusecase "omnilogs-api/internal/audit_logs/usecase"
	"omnilogs-api/internal/main_logs/repository"
	"omnilogs-api/models"
)

var ErrMainLogNotFound = errors.New("main log not found")

type Usecase interface {
	Search(ctx context.Context, input SearchInput, requestID, traceID, ipAddress, userAgent *string) (*SearchResult, error)
	FindByID(ctx context.Context, actorUserID uint, platformAdmin bool, productID int64, logID string, requestID, traceID, ipAddress, userAgent *string) (*MainLogDocument, error)
	FindByAudit(ctx context.Context, actorUserID uint, platformAdmin bool, auditID string, requestID, traceID, ipAddress, userAgent *string) (*models.SystemAuditLog, *MainLogDocument, string, error)
	AuthorizeProductAccess(ctx context.Context, actorUserID uint, platformAdmin bool, productID int64) error
}

type usecase struct {
	repo          repository.Repository
	auditUsecase  auditusecase.Usecase
	encryptionKey string
}

func NewUsecase(repo repository.Repository, auditUsecase auditusecase.Usecase, encryptionKey string) Usecase {
	return &usecase{repo: repo, auditUsecase: auditUsecase, encryptionKey: encryptionKey}
}

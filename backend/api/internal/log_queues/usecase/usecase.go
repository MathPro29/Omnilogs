package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"omnilogs-api/configs"
	"omnilogs-api/dto"
	"omnilogs-api/internal/log_queues/repository"
	"omnilogs-api/internal/queue"
	"omnilogs-api/models"

	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("no pending queue batch found")
	ErrAlreadyExists = errors.New("batch with idempotency key already exists")
	ErrQueueDisabled = errors.New("nats queue is not configured")
)

type Usecase interface {
	Enqueue(ctx context.Context, req dto.IngestLogBatchRequest) (*models.LogQueueBatch, error)
	Consume(ctx context.Context) (*models.LogQueueBatch, error)
	GetItemsByBatch(ctx context.Context, batchID string) ([]models.LogQueueItem, error)
	GetItemByID(ctx context.Context, itemID int64) (*models.LogQueueItem, error)
	GetFailedBatches(ctx context.Context) ([]models.LogQueueBatch, error)
}

type usecase struct {
	repo      repository.Repository
	natsQueue *configs.NATSQueue
}

func NewUsecase(repo repository.Repository, natsQueue *configs.NATSQueue) Usecase {
	return &usecase{repo: repo, natsQueue: natsQueue}
}

func (u *usecase) Enqueue(ctx context.Context, req dto.IngestLogBatchRequest) (*models.LogQueueBatch, error) {
	if u.natsQueue == nil {
		return nil, ErrQueueDisabled
	}

	if req.EnvironmentID != nil {
		if req.ProductID == nil {
			return nil, errors.New("product_id is required when environment_id is provided")
		}

		var count int64
		err := u.repo.DB().Model(&models.ProductEnvironment{}).
			Where("environment_id = ? AND product_id = ? AND deleted_at IS NULL", *req.EnvironmentID, *req.ProductID).
			Count(&count).Error
		if err != nil {
			return nil, fmt.Errorf("validate environment: %w", err)
		}
		if count == 0 {
			return nil, errors.New("environment does not belong to product")
		}
	}

	if req.IdempotencyKey != nil && *req.IdempotencyKey != "" {
		existing, err := u.repo.GetBatchByIdempotencyKey(ctx, *req.IdempotencyKey)
		if err == nil && existing != nil {
			return existing, nil
		}
	}

	now := time.Now().UTC()
	retentionUntil, err := u.resolveBatchPolicy(ctx, req, now)
	if err != nil {
		return nil, err
	}

	batchID := newUUID()
	resolutionStatus := "RESOLVED"
	if req.ProductID == nil {
		resolutionStatus = "PENDING"
	}

	batch := &models.LogQueueBatch{
		BatchID:                 batchID,
		ProductID:               req.ProductID,
		SourceID:                req.SourceID,
		EnvironmentID:           req.EnvironmentID,
		QueueKey:                req.QueueKey,
		SourceType:              req.SourceType,
		SourcePlatform:          req.SourcePlatform,
		IdempotencyKey:          req.IdempotencyKey,
		DetectedProductCode:     req.DetectedProductCode,
		ProductResolutionStatus: resolutionStatus,
		Status:                  "QUEUED",
		Priority:                req.Priority,
		TotalLogs:               len(req.Logs),
		ReceivedAt:              &now,
		RetentionUntil:          retentionUntil,
	}

	if err := u.repo.CreateBatch(ctx, batch); err != nil {
		return nil, err
	}

	for i, logItem := range req.Logs {
		envelope := queue.LogMessage{
			LogID:               newUUID(),
			QueueItemID:         queue.SyntheticQueueItemID(batchID, i+1),
			BatchID:             batchID,
			ProductID:           req.ProductID,
			SourceID:            req.SourceID,
			EnvironmentID:       req.EnvironmentID,
			QueueKey:            req.QueueKey,
			SourceType:          req.SourceType,
			SourcePlatform:      req.SourcePlatform,
			IdempotencyKey:      req.IdempotencyKey,
			DetectedProductCode: chooseDetectedProductCode(logItem.DetectedProductCode, req.DetectedProductCode),
			Priority:            req.Priority,
			SequenceNo:          i + 1,
			InputPayload:        logItem.InputPayload,
			RetentionUntil:      retentionUntil,
			PublishedAt:         now,
		}

		payload, err := queue.MarshalMessage(envelope)
		if err != nil {
			errMsg := err.Error()
			_ = u.repo.UpdateBatchStatus(ctx, batchID, "FAILED", &errMsg)
			return nil, err
		}
		if err := u.natsQueue.Publish(ctx, payload); err != nil {
			errMsg := err.Error()
			_ = u.repo.UpdateBatchStatus(ctx, batchID, "FAILED", &errMsg)
			return nil, err
		}
	}

	return batch, nil
}

func (u *usecase) Consume(ctx context.Context) (*models.LogQueueBatch, error) {
	return nil, ErrNotFound
}

func (u *usecase) GetItemsByBatch(ctx context.Context, batchID string) ([]models.LogQueueItem, error) {
	if batchID == "" {
		return nil, errors.New("batch ID is required")
	}
	return []models.LogQueueItem{}, nil
}

func (u *usecase) GetItemByID(ctx context.Context, itemID int64) (*models.LogQueueItem, error) {
	if itemID <= 0 {
		return nil, errors.New("invalid item ID")
	}
	return nil, gorm.ErrRecordNotFound
}

func (u *usecase) GetFailedBatches(ctx context.Context) ([]models.LogQueueBatch, error) {
	return u.repo.GetFailedBatches(ctx)
}

func (u *usecase) resolveBatchPolicy(ctx context.Context, req dto.IngestLogBatchRequest, now time.Time) (*time.Time, error) {
	retentionDays := 7

	if req.ProductID != nil && req.EnvironmentID != nil {
		var policies []models.LogIngestionPolicy
		err := u.repo.DB().WithContext(ctx).
			Where("product_id = ? AND environment_id = ?", *req.ProductID, *req.EnvironmentID).
			Limit(1).
			Find(&policies).Error
		if err != nil {
			return nil, err
		}
		if len(policies) > 0 {
			policy := policies[0]
			if policy.RetentionDays != nil {
				retentionDays = *policy.RetentionDays
			}
		}
	}

	if retentionDays <= 0 {
		return nil, nil
	}

	retentionUntil := now.AddDate(0, 0, retentionDays)
	return &retentionUntil, nil
}

func chooseDetectedProductCode(itemValue *string, batchValue *string) *string {
	if itemValue != nil && *itemValue != "" {
		return itemValue
	}
	return batchValue
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
	hex.Encode(dst[24:], bytes[10:])
	return string(dst)
}

package repository

import (
	"context"
	"time"

	"omnilogs-api/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	DB() *gorm.DB
	ClaimNextBatch(ctx context.Context, workerID string, lockTimeout time.Duration) (*models.LogQueueBatch, error)
	LoadPendingItems(ctx context.Context, batchID string) ([]models.LogQueueItem, error)
	MarkItemProcessing(ctx context.Context, queueItemID int64, workerID string, now time.Time) error
	UpsertIndexRef(ctx context.Context, indexRef *models.LogIndexRef) error
	MarkItemProcessed(ctx context.Context, queueItemID int64, indexedAt time.Time) error
	CreateFailure(ctx context.Context, failure *models.LogFailure) error
	UpdateItemFailureState(ctx context.Context, queueItemID int64, nextStatus string, retryCount int, nextRetryAt *time.Time, now time.Time, reason string) error
	FinishBatch(ctx context.Context, batchID, status string, message *string) error
	FindIndexPolicy(ctx context.Context, productID int, environmentID *int) (*models.ElasticIndexPolicy, error)
	GetSensitiveFieldDefinitions(ctx context.Context, productID int) ([]models.LogFieldDefinition, error)
	GetLogMaskingRules(ctx context.Context, productID int) ([]models.LogMaskingRule, error)
	CreateSensitiveFieldSecret(ctx context.Context, secret *models.LogSensitiveFieldSecret) error
	DeleteQueueItem(ctx context.Context, queueItemID int64) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) DB() *gorm.DB {
	return r.db
}

func (r *repository) ClaimNextBatch(ctx context.Context, workerID string, lockTimeout time.Duration) (*models.LogQueueBatch, error) {
	var claimed *models.LogQueueBatch
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var batch models.LogQueueBatch
		lockCutoff := time.Now().Add(-lockTimeout)

		query := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status IN ?", []string{"QUEUED", "RETRY_PENDING", "PARTIAL"}).
			Where("(available_at IS NULL OR available_at <= ?)", time.Now()).
			Where("(locked_at IS NULL OR locked_at < ?)", lockCutoff).
			Order("priority DESC").
			Order("received_at ASC NULLS LAST").
			Order("created_at ASC NULLS LAST")

		if err := query.First(&batch).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil
			}
			return err
		}

		now := time.Now()
		updates := map[string]any{
			"status":                "PROCESSING",
			"worker_id":             workerID,
			"locked_by":             workerID,
			"locked_at":             now,
			"processing_started_at": now,
			"processing_attempts":   gorm.Expr("processing_attempts + 1"),
			"error_message":         nil,
		}
		if err := tx.Model(&models.LogQueueBatch{}).
			Where("batch_id = ?", batch.BatchID).
			Updates(updates).Error; err != nil {
			return err
		}

		batch.Status = "PROCESSING"
		batch.WorkerID = stringPtr(workerID)
		batch.LockedBy = stringPtr(workerID)
		batch.LockedAt = &now
		batch.ProcessingStartedAt = &now
		claimed = &batch
		return nil
	})
	return claimed, err
}

func (r *repository) DeleteQueueItem(ctx context.Context, queueItemID int64) error {
	return r.db.WithContext(ctx).Model(&models.LogQueueItem{}).
		Where("queue_item_id = ?", queueItemID).
		Delete(&models.LogQueueItem{}).Error
}

func (r *repository) LoadPendingItems(ctx context.Context, batchID string) ([]models.LogQueueItem, error) {
	var items []models.LogQueueItem
	err := r.db.WithContext(ctx).
		Where("batch_id = ?", batchID).
		Where("status IN ?", []string{"PENDING", "RETRY_PENDING"}).
		Where("(next_retry_at IS NULL OR next_retry_at <= ?)", time.Now()).
		Order("sequence_no ASC").
		Find(&items).Error
	return items, err
}

func (r *repository) MarkItemProcessing(ctx context.Context, queueItemID int64, workerID string, now time.Time) error {
	return r.db.WithContext(ctx).Model(&models.LogQueueItem{}).
		Where("queue_item_id = ?", queueItemID).
		Updates(map[string]any{
			"status":                "PROCESSING",
			"worker_id":             workerID,
			"processing_started_at": now,
			"processing_attempts":   gorm.Expr("processing_attempts + 1"),
			"error_message":         nil,
		}).Error
}

func (r *repository) UpsertIndexRef(ctx context.Context, indexRef *models.LogIndexRef) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "elastic_index"}, {Name: "elastic_document_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"index_status", "indexed_at", "last_sync_at", "sync_error", "response_status_code", "updated_at"}),
	}).Create(indexRef).Error
}

func (r *repository) MarkItemProcessed(ctx context.Context, queueItemID int64, indexedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.LogQueueItem{}).
		Where("queue_item_id = ?", queueItemID).
		Updates(map[string]any{
			"status":        "PROCESSED",
			"processed_at":  indexedAt,
			"last_retry_at": indexedAt,
			"error_message": nil,
		}).Error
}

func (r *repository) CreateFailure(ctx context.Context, failure *models.LogFailure) error {
	return r.db.WithContext(ctx).Create(failure).Error
}

func (r *repository) UpdateItemFailureState(ctx context.Context, queueItemID int64, nextStatus string, retryCount int, nextRetryAt *time.Time, now time.Time, reason string) error {
	return r.db.WithContext(ctx).Model(&models.LogQueueItem{}).
		Where("queue_item_id = ?", queueItemID).
		Updates(map[string]any{
			"status":                nextStatus,
			"retry_count":           retryCount,
			"next_retry_at":         nextRetryAt,
			"last_retry_at":         now,
			"error_message":         reason,
			"processing_started_at": now,
		}).Error
}

func (r *repository) FinishBatch(ctx context.Context, batchID, status string, message *string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.LogQueueBatch{}).
		Where("batch_id = ?", batchID).
		Updates(map[string]any{
			"status":        status,
			"processed_at":  now,
			"locked_by":     nil,
			"locked_at":     nil,
			"error_message": message,
		}).Error
}

func (r *repository) FindIndexPolicy(ctx context.Context, productID int, environmentID *int) (*models.ElasticIndexPolicy, error) {
	var policy models.ElasticIndexPolicy
	query := r.db.WithContext(ctx).Where("product_id = ? AND is_active = TRUE", productID)
	if environmentID != nil {
		query = query.Where("(environment_id IS NULL OR environment_id = ?)", *environmentID).Order("environment_id DESC NULLS LAST")
	} else {
		query = query.Where("environment_id IS NULL")
	}

	if err := query.First(&policy).Error; err != nil {
		return nil, err
	}
	return &policy, nil
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func (r *repository) GetSensitiveFieldDefinitions(ctx context.Context, productID int) ([]models.LogFieldDefinition, error) {
	var list []models.LogFieldDefinition
	err := r.db.WithContext(ctx).Where("product_id = ? AND is_sensitive = TRUE AND is_active = TRUE", productID).Find(&list).Error
	return list, err
}
func (r *repository) GetLogMaskingRules(ctx context.Context, productID int) ([]models.LogMaskingRule, error) {
	var list []models.LogMaskingRule
	err := r.db.WithContext(ctx).Where("product_id = ? AND is_active = TRUE", productID).Find(&list).Error
	return list, err
}
func (r *repository) CreateSensitiveFieldSecret(ctx context.Context, secret *models.LogSensitiveFieldSecret) error {
	return r.db.WithContext(ctx).Create(secret).Error
}

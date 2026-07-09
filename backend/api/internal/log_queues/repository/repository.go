package repository

import (
	"context"
	"time"

	"omnilogs-api/models"

	"gorm.io/gorm"
)

type Repository interface {
	DB() *gorm.DB
	Transaction(fn func(*gorm.DB) error) error
	CreateBatch(ctx context.Context, batch *models.LogQueueBatch) error
	CreateItems(ctx context.Context, items []models.LogQueueItem) error
	GetBatchByIdempotencyKey(ctx context.Context, key string) (*models.LogQueueBatch, error)
	GetPendingBatch(ctx context.Context) (*models.LogQueueBatch, error)
	GetItemsByBatchID(ctx context.Context, batchID string) ([]models.LogQueueItem, error)
	GetItemByID(ctx context.Context, itemID int64) (*models.LogQueueItem, error)
	UpdateBatchStatus(ctx context.Context, batchID string, status string, errorMessage *string) error
	UpdateItemStatus(ctx context.Context, itemID int64, status string, errorMessage *string) error
	GetFailedBatches(ctx context.Context) ([]models.LogQueueBatch, error)
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

func (r *repository) Transaction(fn func(*gorm.DB) error) error {
	return r.db.Transaction(fn)
}

func (r *repository) CreateBatch(ctx context.Context, batch *models.LogQueueBatch) error {
	return r.db.WithContext(ctx).Create(batch).Error
}

func (r *repository) CreateItems(ctx context.Context, items []models.LogQueueItem) error {
	return r.db.WithContext(ctx).Create(&items).Error
}

func (r *repository) GetBatchByIdempotencyKey(ctx context.Context, key string) (*models.LogQueueBatch, error) {
	var batch models.LogQueueBatch
	err := r.db.WithContext(ctx).Where("idempotency_key = ?", key).First(&batch).Error
	if err != nil {
		return nil, err
	}
	return &batch, nil
}

func (r *repository) GetItemByID(ctx context.Context, itemID int64) (*models.LogQueueItem, error) {
	var item models.LogQueueItem
	err := r.db.WithContext(ctx).Where("queue_item_id = ?", itemID).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *repository) GetPendingBatch(ctx context.Context) (*models.LogQueueBatch, error) {
	var batch models.LogQueueBatch
	// ดึง Batch แรกที่มีสถานะ QUEUED โดยเรียงตาม Priority สูงสุดก่อน
	err := r.db.WithContext(ctx).
		Where("status = ?", "QUEUED").
		Order("priority DESC").
		Order("received_at ASC").
		First(&batch).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &batch, nil
}

func (r *repository) GetItemsByBatchID(ctx context.Context, batchID string) ([]models.LogQueueItem, error) {
	var items []models.LogQueueItem
	err := r.db.WithContext(ctx).
		Where("batch_id = ?", batchID).
		Order("sequence_no ASC").
		Find(&items).Error
	return items, err
}

func (r *repository) UpdateBatchStatus(ctx context.Context, batchID string, status string, errorMessage *string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":       status,
		"processed_at": &now,
	}
	if errorMessage != nil {
		updates["error_message"] = *errorMessage
	} else {
		updates["error_message"] = nil
	}
	return r.db.WithContext(ctx).Model(&models.LogQueueBatch{}).
		Where("batch_id = ?", batchID).
		Updates(updates).Error
}

func (r *repository) UpdateItemStatus(ctx context.Context, itemID int64, status string, errorMessage *string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":       status,
		"processed_at": &now,
	}
	if errorMessage != nil {
		updates["error_message"] = *errorMessage
	} else {
		updates["error_message"] = nil
	}
	return r.db.WithContext(ctx).Model(&models.LogQueueItem{}).
		Where("queue_item_id = ?", itemID).
		Updates(updates).Error
}

func (r *repository) GetFailedBatches(ctx context.Context) ([]models.LogQueueBatch, error) {
	var batches []models.LogQueueBatch
	err := r.db.WithContext(ctx).
		Where(`
			EXISTS (
				SELECT 1
				FROM log_failures lf
				WHERE lf.batch_id = log_queue_batches.batch_id
				  AND lf.status = ?
			)
		`, "FAILED").
		Order("received_at ASC").
		Limit(100).
		Find(&batches).Error
	return batches, err
}

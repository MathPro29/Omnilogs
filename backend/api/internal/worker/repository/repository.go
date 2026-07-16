package repository

import (
	"context"

	"omnilogs-api/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	DB() *gorm.DB
	UpsertIndexRef(ctx context.Context, indexRef *models.LogIndexRef) error
	CreateFailure(ctx context.Context, failure *models.LogFailure) error
	FindIndexPolicy(ctx context.Context, productID int, environmentID *int) (*models.ElasticIndexPolicy, error)
	GetActiveFieldDefinitions(ctx context.Context, productID int) ([]models.LogFieldDefinition, error)
	GetLogMaskingRules(ctx context.Context, productID int) ([]models.LogMaskingRule, error)
	GetActiveIndexPolicies(ctx context.Context) ([]models.ElasticIndexPolicy, error)
	CreateLogArchive(ctx context.Context, archive *models.LogArchive) error
	BulkPersistSuccesses(ctx context.Context, indexRefs []models.LogIndexRef, objectRefs []models.LogObjectStorageRef, secrets []models.LogSensitiveFieldSecret) error
	RefreshBatchStatusFromResults(ctx context.Context, batchID string) error
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

func (r *repository) UpsertIndexRef(ctx context.Context, indexRef *models.LogIndexRef) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "elastic_index"}, {Name: "elastic_document_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"index_status", "indexed_at", "last_sync_at", "sync_error", "response_status_code", "updated_at"}),
	}).Create(indexRef).Error
}

func (r *repository) CreateFailure(ctx context.Context, failure *models.LogFailure) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "queue_item_id"}, {Name: "attempt_no"}},
		DoUpdates: clause.AssignmentColumns([]string{"log_id", "sequence_no", "failure_stage", "failure_type", "reason", "error_details", "retry_count", "max_retry_count", "next_retry_at", "last_retry_at", "status", "resolved_at"}),
	}).Create(failure).Error
}

func (r *repository) BulkPersistSuccesses(ctx context.Context, indexRefs []models.LogIndexRef, objectRefs []models.LogObjectStorageRef, secrets []models.LogSensitiveFieldSecret) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(indexRefs) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "elastic_index"}, {Name: "elastic_document_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"log_id", "batch_id", "queue_item_id", "response_status_code", "index_status", "indexed_at", "last_sync_at", "sync_error", "updated_at"}),
			}).CreateInBatches(indexRefs, 500).Error; err != nil {
				return err
			}
		}

		if len(objectRefs) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "log_id"}, {Name: "object_type"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"product_id", "environment_id", "storage_provider", "bucket_name", "object_path", "file_format",
					"size_bytes", "checksum", "encrypted_payload", "encryption_key_ref", "encryption_algorithm",
					"is_encrypted", "retention_until", "purged_at",
				}),
			}).CreateInBatches(objectRefs, 500).Error; err != nil {
				return err
			}
		}

		if len(secrets) > 0 {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(secrets, 500).Error; err != nil {
				return err
			}
		}

		if len(indexRefs) > 0 {
			queueItemIDs := make([]int64, 0, len(indexRefs))
			for _, ref := range indexRefs {
				if ref.QueueItemID == nil || *ref.QueueItemID == 0 {
					continue
				}
				queueItemIDs = append(queueItemIDs, *ref.QueueItemID)
			}
			if len(queueItemIDs) > 0 {
				if err := tx.Model(&models.LogFailure{}).
					Where("queue_item_id IN ?", queueItemIDs).
					Where("status <> ?", "RESOLVED").
					Updates(map[string]any{
						"status":        "RESOLVED",
						"resolved_at":   gorm.Expr("NOW()"),
						"next_retry_at": nil,
					}).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (r *repository) RefreshBatchStatusFromResults(ctx context.Context, batchID string) error {
	type resultCounts struct {
		TotalLogs int
		Indexed   int
		Failed    int
	}

	var counts resultCounts
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			b.total_logs,
			COALESCE((
				SELECT COUNT(DISTINCT queue_item_id)
				FROM log_index_refs
				WHERE batch_id = b.batch_id
			), 0) AS indexed,
			COALESCE((
				SELECT COUNT(DISTINCT queue_item_id)
				FROM log_failures
				WHERE batch_id = b.batch_id AND status = 'FAILED'
			), 0) AS failed
		FROM log_queue_batches b
		WHERE b.batch_id = ?
	`, batchID).Scan(&counts).Error; err != nil {
		return err
	}

	status := "PROCESSING"
	message := ""
	switch {
	case counts.Indexed == counts.TotalLogs && counts.TotalLogs > 0:
		status = "COMPLETED"
		message = "all logs indexed successfully"
	case counts.Indexed == 0 && counts.Failed == counts.TotalLogs && counts.TotalLogs > 0:
		status = "FAILED"
		message = "all logs failed permanently"
	case counts.Indexed+counts.Failed == counts.TotalLogs && counts.Failed > 0 && counts.TotalLogs > 0:
		status = "PARTIAL"
		message = "batch completed with permanent failures"
	default:
		status = "PROCESSING"
		message = "batch still processing"
	}

	updates := map[string]any{
		"status":        status,
		"error_message": message,
	}
	if status != "PROCESSING" {
		updates["processed_at"] = gorm.Expr("NOW()")
		updates["locked_by"] = nil
		updates["locked_at"] = nil
	}

	return r.db.WithContext(ctx).Model(&models.LogQueueBatch{}).
		Where("batch_id = ?", batchID).
		Updates(updates).Error
}

func (r *repository) FindIndexPolicy(ctx context.Context, productID int, environmentID *int) (*models.ElasticIndexPolicy, error) {
	var list []models.ElasticIndexPolicy
	query := r.db.WithContext(ctx).Where("product_id = ? AND is_active = TRUE", productID)
	if environmentID != nil {
		query = query.Where("(environment_id IS NULL OR environment_id = ?)", *environmentID).Order("environment_id DESC NULLS LAST")
	} else {
		query = query.Where("environment_id IS NULL")
	}

	if err := query.Limit(1).Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &list[0], nil
}

func (r *repository) GetActiveFieldDefinitions(ctx context.Context, productID int) ([]models.LogFieldDefinition, error) {
	var list []models.LogFieldDefinition
	err := r.db.WithContext(ctx).Where("(product_id = ? OR product_id IS NULL) AND is_active = TRUE", productID).Find(&list).Error
	return list, err
}

func (r *repository) GetLogMaskingRules(ctx context.Context, productID int) ([]models.LogMaskingRule, error) {
	var list []models.LogMaskingRule
	err := r.db.WithContext(ctx).Where("(product_id = ? OR product_id IS NULL) AND is_active = TRUE", productID).Find(&list).Error
	return list, err
}

func (r *repository) GetActiveIndexPolicies(ctx context.Context) ([]models.ElasticIndexPolicy, error) {
	var list []models.ElasticIndexPolicy
	err := r.db.WithContext(ctx).Where("retention_enabled = TRUE AND retention_days IS NOT NULL AND is_active = TRUE").Find(&list).Error
	return list, err
}

func (r *repository) CreateLogArchive(ctx context.Context, archive *models.LogArchive) error {
	return r.db.WithContext(ctx).Create(archive).Error
}

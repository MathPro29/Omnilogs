package migrate

import "gorm.io/gorm"

var queryPerformanceIndexStatements = []string{
	`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_log_queue_batches_queued_order
		ON log_queue_batches (priority DESC, received_at ASC)
		WHERE status = 'QUEUED'`,
	`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_log_failures_failed_batch_item
		ON log_failures (batch_id, queue_item_id)
		WHERE status = 'FAILED'`,
}

var queryPerformanceIndexRollbackStatements = []string{
	`DROP INDEX CONCURRENTLY IF EXISTS idx_log_queue_batches_queued_order`,
	`DROP INDEX CONCURRENTLY IF EXISTS idx_log_failures_failed_batch_item`,
}

func ensureQueryPerformanceIndexes(db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" {
		return nil
	}
	for _, statement := range queryPerformanceIndexStatements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

// RollbackQueryPerformanceIndexes provides an explicit rollback for the
// non-destructive performance migration. CONCURRENTLY avoids blocking normal
// log writes while the indexes are removed.
func RollbackQueryPerformanceIndexes(db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" {
		return nil
	}
	for _, statement := range queryPerformanceIndexRollbackStatements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

package usecase

import (
	"context"
	"fmt"
	"time"

	"omnilogs-api/models"
)

// handleItemFailure จัดการเมื่อเกิดข้อผิดพลาดในการประมวลผล Item ใดๆ
// โดยจะทำการคำนวณรอบ Retry และบันทึกประวัติความผิดพลาดลง DB (log_failures)
// ถ้าหมดโควต้า Retry (MaxRetryCount) จะถือว่า FAILED ถาวรและลบออกจากคิว
func (u *usecase) handleItemFailure(ctx context.Context, batch *models.LogQueueBatch, item *models.LogQueueItem, stage, failureType, reason string) error {
	now := time.Now()
	retryCount := item.RetryCount + 1

	nextStatus := "FAILED"
	var nextRetryAt *time.Time
	if retryCount <= item.MaxRetryCount {
		nextStatus = "RETRY_PENDING"
		retryAt := now.Add(time.Duration(retryCount) * time.Minute)
		nextRetryAt = &retryAt
	}

	reasonCopy := reason
	itemFailure := models.LogFailure{
		FailureID:     newUUID(),
		AttemptNo:     item.ProcessingAttempts + 1,
		QueueItemID:   item.QueueItemID,
		BatchID:       batch.BatchID,
		ProductID:     batch.ProductID,
		SourceID:      batch.SourceID,
		EnvironmentID: batch.EnvironmentID,
		FailureStage:  stage,
		FailureType:   failureType,
		Reason:        &reasonCopy,
		RetryCount:    retryCount,
		MaxRetryCount: item.MaxRetryCount,
		NextRetryAt:   nextRetryAt,
		LastRetryAt:   &now,
		Status:        nextStatus,
	}

	// 1. บันทึกประวัติความล้มเหลวลงใน LogFailure เสมอ (สำหรับทั้งกรณี Retry และ Failed ถาวร)
	_ = u.repo.CreateFailure(ctx, &itemFailure)

	// 2. จัดการข้อมูลในตาราง Queue ตามสถานะ
	if nextStatus == "FAILED" {
		// ถ้า Fail ถาวร (ไม่มี retry แล้ว) ให้ลบออกจากตาราง queue ไปเลย
		_ = u.repo.DeleteQueueItem(ctx, item.QueueItemID)
	} else {
		// ถ้ายังติด RETRY_PENDING ให้เก็บไว้ใน queue และอัปเดต state
		_ = u.repo.UpdateItemFailureState(ctx, item.QueueItemID, nextStatus, retryCount, nextRetryAt, now, reason)
	}

	if nextStatus == "RETRY_PENDING" {
		return fmt.Errorf("retry scheduled: %s", reason)
	}
	return fmt.Errorf("%s", reason)
}

func (u *usecase) failBatch(ctx context.Context, batch *models.LogQueueBatch, reason string) error {
	return u.repo.FinishBatch(ctx, batch.BatchID, "FAILED", &reason)
}

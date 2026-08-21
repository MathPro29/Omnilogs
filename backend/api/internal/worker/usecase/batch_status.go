package usecase

import (
	"context"
	"log/slog"
	"time"

	"omnilogs-api/dto"
	"omnilogs-api/models"

	"github.com/nats-io/nats.go"
)

func collectBatchIDs(msgs []*nats.Msg) map[string]struct{} {
	batchIDs := make(map[string]struct{})
	for _, msg := range msgs {
		envelope, err := decodeMessage(msg)
		if err != nil || envelope.BatchID == "" {
			continue
		}
		batchIDs[envelope.BatchID] = struct{}{}
	}
	return batchIDs
}

func (u *usecase) refreshBatchStatuses(ctx context.Context, batchIDs map[string]struct{}) {
	for batchID := range batchIDs {
		if err := u.repo.RefreshBatchStatusFromResults(ctx, batchID); err != nil {
			slog.Error("failed to refresh batch status", "batch_id", batchID, "error", err)
		}
	}
}

func (u *usecase) markFetchedBatchesProcessing(ctx context.Context, msgs []*nats.Msg) error {
	batchIDs := make(map[string]struct{})
	for _, msg := range msgs {
		envelope, err := decodeMessage(msg)
		if err != nil || envelope.BatchID == "" {
			continue
		}
		batchIDs[envelope.BatchID] = struct{}{}
	}
	if len(batchIDs) == 0 {
		return nil
	}

	ids := make([]string, 0, len(batchIDs))
	for batchID := range batchIDs {
		ids = append(ids, batchID)
	}

	now := time.Now().UTC()
	if err := u.repo.DB().WithContext(ctx).Model(&models.LogQueueBatch{}).
		Where("batch_id IN ?", ids).
		Updates(map[string]any{
			"status":                "PROCESSING",
			"processing_started_at": now,
			"processed_at":          nil,
			"error_message":         nil,
		}).Error; err != nil {
		slog.Error("failed to mark fetched batches as processing", "error", err)
		return err
	}
	return nil
}

func (u *usecase) failMessage(ctx context.Context, msg *nats.Msg, envelope dto.LogMessage, stage string, failureType string, reason string, retryable bool) error {
	attemptNo := 1
	if metadata, err := msg.Metadata(); err == nil && metadata.NumDelivered > 0 {
		attemptNo = int(metadata.NumDelivered)
	}
	maxRetryCount := envelope.MaxRetryCount
	if maxRetryCount <= 0 {
		maxRetryCount = 3
	}
	retryCount := attemptNo - 1
	status := "FAILED"
	var nextRetryAt *time.Time
	if retryable && retryCount < maxRetryCount {
		delay := time.Second << min(retryCount, 5)
		next := time.Now().UTC().Add(delay)
		nextRetryAt = &next
		status = "RETRYING"
	}
	reasonCopy := reason
	failure := models.LogFailure{
		FailureID:     newUUID(),
		LogID:         envelope.LogID,
		AttemptNo:     attemptNo,
		QueueItemID:   envelope.QueueItemID,
		BatchID:       envelope.BatchID,
		SequenceNo:    envelope.SequenceNo,
		ProductID:     envelope.ProductID,
		SourceID:      envelope.SourceID,
		EnvironmentID: envelope.EnvironmentID,
		FailureStage:  stage,
		FailureType:   failureType,
		Reason:        &reasonCopy,
		ErrorDetails:  envelope.InputPayload, // เก็บ Payload ดิบเอาไว้สำหรับนำไปทำ Retry
		RetryCount:    retryCount,
		MaxRetryCount: maxRetryCount,
		NextRetryAt:   nextRetryAt,
		LastRetryAt:   func() *time.Time { now := time.Now().UTC(); return &now }(),
		Status:        status,
	}

	if err := u.repo.CreateFailure(ctx, &failure); err != nil {
		return err
	}
	if status == "RETRYING" {
		return msg.NakWithDelay(time.Until(*nextRetryAt))
	}

	if err := u.repo.DB().WithContext(ctx).
		Where("queue_item_id = ?", envelope.QueueItemID).
		Delete(&models.LogQueueItem{}).Error; err != nil {
		return err
	}
	if err := msg.Ack(); err != nil {
		return err
	}
	return nil
}

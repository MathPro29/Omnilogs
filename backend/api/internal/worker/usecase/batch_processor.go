package usecase

import (
	"context"

	"omnilogs-api/models"

	"github.com/nats-io/nats.go"
)

func (u *usecase) processFetchedMessages(ctx context.Context, msgs []*nats.Msg) error {
	batchIDs := collectBatchIDs(msgs)
	if len(batchIDs) > 0 {
		defer u.refreshBatchStatuses(ctx, batchIDs)
	}

	if err := u.markFetchedBatchesProcessing(ctx, msgs); err != nil {
		return err
	}

	prepared, err := u.prepareLogs(ctx, msgs)
	if err != nil {
		return err
	}
	if len(prepared) == 0 {
		return nil
	}

	prepared, err = u.bulkIndexDocuments(ctx, prepared)
	if err != nil {
		for _, item := range prepared {
			if retryErr := u.failMessage(ctx, item.natsMsg, item.message, "ELASTICSEARCH", "BULK_REQUEST_FAILED", err.Error(), true); retryErr != nil {
				return retryErr
			}
		}
		return nil
	}
	if len(prepared) == 0 {
		return nil
	}

	indexRefs := make([]models.LogIndexRef, 0, len(prepared))
	objectRefs := make([]models.LogObjectStorageRef, 0, len(prepared))
	secrets := make([]models.LogSensitiveFieldSecret, 0, len(prepared))

	for _, item := range prepared {
		indexRefs = append(indexRefs, item.indexRef)
		objectRefs = append(objectRefs, item.objectRef)
		secrets = append(secrets, item.secrets...)
	}

	if err := u.repo.BulkPersistSuccesses(ctx, indexRefs, objectRefs, secrets); err != nil {
		return err
	}

	for _, item := range prepared {
		if err := item.natsMsg.Ack(); err != nil {
			return err
		}
	}

	return nil
}

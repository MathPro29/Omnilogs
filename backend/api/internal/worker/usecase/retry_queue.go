package usecase

import (
	"context"
	"omnilogs-api/models"
)

// make retry items queue

func (u *usecase) RetryItems(ctx context.Context) error {
	// check in log_queue_batch ว่ามี Retry ไหม

	// ถ้ามีให้ลอง saved ลง queue อีกรอบ 
}



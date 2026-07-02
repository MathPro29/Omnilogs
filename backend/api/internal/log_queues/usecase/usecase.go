package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"omnilogs-api/dto"
	"omnilogs-api/internal/log_queues/repository"
	"omnilogs-api/models"

	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("no pending queue batch found")
	ErrAlreadyExists = errors.New("batch with idempotency key already exists")
)

type Usecase interface {
	Enqueue(ctx context.Context, req dto.IngestLogBatchRequest) (*models.LogQueueBatch, error)
	Consume(ctx context.Context) (*models.LogQueueBatch, error)
	GetItemsByBatch(ctx context.Context, batchID string) ([]models.LogQueueItem, error)
	GetItemByID(ctx context.Context, itemID int64) (*models.LogQueueItem, error)
}

type usecase struct {
	repo repository.Repository
}

func NewUsecase(repo repository.Repository) Usecase {
	return &usecase{repo: repo}
}

// [KEY : Enqueue ตรวจสอบก่อนว่ามี ENV ใน Product และ ตรงกับในตารางไหม]
func (u *usecase) Enqueue(ctx context.Context, req dto.IngestLogBatchRequest) (*models.LogQueueBatch, error) {
	// ส่วนที่ 1: ป้องกัน Cross-product ระหว่าง Product และ Environment
	// ถ้าส่ง environment_id มา ต้องส่ง product_id มาด้วย และ Environment ต้องอยู่ใน Product นั้นจริง
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
	// 1. ตรวจสอบ Idempotency Key ป้องกันการส่งซ้ำ
	if req.IdempotencyKey != nil && *req.IdempotencyKey != "" {
		// ส่วนที่ 2: ถ้าพบ Idempotency Key เดิม ให้คืน Batch เดิมเพื่อไม่สร้างข้อมูลซ้ำ
		existing, err := u.repo.GetBatchByIdempotencyKey(ctx, *req.IdempotencyKey)
		if err == nil && existing != nil {
			return existing, nil // ส่ง Batch ตัวเดิมกลับไป
		}
	}

	// 2. สร้าง Batch ID เป็น UUID
	batchID := newUUID()
	// ส่วนที่ 3: สร้าง Batch ใหม่ และกำหนดสถานะการค้นหา Product
	// หากยังไม่มี product_id สถานะจะเป็น PENDING เพื่อรอการ Resolve
	now := time.Now()

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
	}

	// 3. แมป Logs ใน Request ลงในโมเดล LogQueueItem
	items := make([]models.LogQueueItem, len(req.Logs))
	// ส่วนที่ 4: แปลง Log แต่ละรายการเป็น Queue Item
	// เริ่มด้วยสถานะ PENDING และ Retry ได้สูงสุด 3 ครั้ง
	for i, logItem := range req.Logs {
		sizeBytes := int64(len(logItem.InputPayload))
		items[i] = models.LogQueueItem{
			BatchID:             batchID,
			SequenceNo:          logItem.SequenceNo,
			SourceType:          logItem.SourceType,
			SourcePlatform:      logItem.SourcePlatform,
			InputPayload:        logItem.InputPayload,
			PayloadSizeBytes:    &sizeBytes,
			DetectedProductCode: logItem.DetectedProductCode,
			Status:              "PENDING",
			MaxRetryCount:       3,
		}
	}

	// 4. บันทึกลงฐานข้อมูลภายใต้ Database Transaction
	err := u.repo.Transaction(func(tx *gorm.DB) error {
		// ส่วนที่ 5: บันทึก Batch และ Items ใน Transaction เดียวกัน
		// ถ้าส่วนใดล้มเหลว ฐานข้อมูลจะ Rollback ทั้งชุด
		if err := tx.Create(batch).Error; err != nil {
			return err
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return batch, nil
}

func (u *usecase) Consume(ctx context.Context) (*models.LogQueueBatch, error) {
	// 1. ดึง Batch ที่ค้างอยู่ตัวถัดไป
	batch, err := u.repo.GetPendingBatch(ctx)
	if err != nil {
		return nil, err
	}
	if batch == nil {
		return nil, ErrNotFound
	}

	// 2. ปรับสถานะ Batch เป็น PROCESSING
	err = u.repo.UpdateBatchStatus(ctx, batch.BatchID, "PROCESSING", nil)
	if err != nil {
		return nil, err
	}
	batch.Status = "PROCESSING"

	// 3. ดึง Log Queue Items ของ Batch นี้มาทำงาน
	items, err := u.repo.GetItemsByBatchID(ctx, batch.BatchID)
	if err != nil {
		errMsg := err.Error()
		_ = u.repo.UpdateBatchStatus(ctx, batch.BatchID, "FAILED", &errMsg)
		return nil, err
	}

	// 4. วนลูปประมวลผล (ใน API Handler เราจะประมวลผลจำลองเปลี่ยนสถานะเป็น PROCESSED)
	for _, item := range items {
		err := u.repo.UpdateItemStatus(ctx, item.QueueItemID, "PROCESSED", nil)
		if err != nil {
			errMsg := err.Error()
			_ = u.repo.UpdateItemStatus(ctx, item.QueueItemID, "FAILED", &errMsg)
		}
	}

	// 5. ปรับสถานะ Batch เป็น COMPLETED
	successMsg := "processed successfully"
	err = u.repo.UpdateBatchStatus(ctx, batch.BatchID, "COMPLETED", &successMsg)
	if err != nil {
		return nil, err
	}
	batch.Status = "COMPLETED"
	batch.ErrorMessage = &successMsg

	return batch, nil
}

func (u *usecase) GetItemsByBatch(ctx context.Context, batchID string) ([]models.LogQueueItem, error) {
	if batchID == "" {
		return nil, errors.New("batch ID is required")
	}
	return u.repo.GetItemsByBatchID(ctx, batchID)
}
func (u *usecase) GetItemByID(ctx context.Context, itemID int64) (*models.LogQueueItem, error) {
	if itemID <= 0 {
		return nil, errors.New("invalid item ID")
	}
	return u.repo.GetItemByID(ctx, itemID)
}

// newUUID ฟังก์ชันจำลองการสร้าง UUID V4 ตามมาตรฐาน RFC 4122
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

package usecase

import (
	"context"
	"log/slog"
	"time"

	workerrepo "omnilogs-api/internal/worker/repository"

	"github.com/elastic/go-elasticsearch/v8"
)

const (
	DefaultPollInterval = 5 * time.Second  // ความถี่ในการดึงคิวจาก Database มาประมวลผล
	DefaultLockTimeout  = 15 * time.Minute // ระยะเวลาล็อกงานไม่ให้ Worker ตัวอื่นหยิบซ้ำในระหว่างทำงาน
)

// Usecase กำหนดฟังก์ชันหลักของ Worker ที่ทำหน้าที่ประมวลผลคิว
// - Run: ใช้สำหรับรัน Worker แบบ Background Service (ทำงานวนลูปไปเรื่อยๆ)
// - RunOnce: ใช้สำหรับเรียกทำงานแค่ 1 รอบ (ใช้ตอนที่ API รับข้อมูลแล้วสั่งประมวลผลทันที)
type Usecase interface {
	Run(ctx context.Context) error
	RunOnce(ctx context.Context) (bool, error)
}

type usecase struct {
	repo          workerrepo.Repository
	esClient      *elasticsearch.Client
	workerID      string
	pollInterval  time.Duration
	lockTimeout   time.Duration
	encryptionKey string
}

func NewUsecase(repo workerrepo.Repository, esClient *elasticsearch.Client, encryptionKey string) Usecase {
	return &usecase{
		repo:          repo,
		esClient:      esClient,
		workerID:      "worker-" + newUUID(),
		pollInterval:  DefaultPollInterval,
		lockTimeout:   DefaultLockTimeout,
		encryptionKey: encryptionKey,
	}
}

// Run เริ่มทำงาน Worker แบบลูปอนันต์ (Daemon)
// จะทำงานดึงคิวทุกๆ pollInterval ถ้าไม่มีคิวก็รอจนครบเวลาแล้วดึงใหม่
func (u *usecase) Run(ctx context.Context) error {
	ticker := time.NewTicker(u.pollInterval)
	defer ticker.Stop()

	for {
		// พยายามดึงและประมวลผล Batch หนึ่งตัว
		if err := u.processNextBatch(ctx); err != nil {
			slog.Error("batch processing failed",
				slog.String("worker_id", u.workerID),
				slog.Any("error", err),
			)
		}

		// รอจนกว่าจะครบเวลา หรือ Context ถูกสั่งหยุด (Graceful Shutdown)
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

// RunOnce ดึงคิวมาประมวลผลเพียง 1 Batch
// คืนค่า true ถ้ามีคิวให้ประมวลผล คืนค่า false ถ้าไม่มีคิวเหลืออยู่ในระบบ
func (u *usecase) RunOnce(ctx context.Context) (bool, error) {
	// ค้นหาและ Claim คิวเพื่อป้องกัน Worker ตัวอื่นมาหยิบซ้ำ (Distributed Lock)
	batch, err := u.repo.ClaimNextBatch(ctx, u.workerID, u.lockTimeout)
	if err != nil {
		return false, err
	}
	if batch == nil {
		return false, nil // ไม่มีคิว
	}
	// ประมวลผล Batch ที่ Claim มาได้
	return true, u.processClaimedBatch(ctx, batch)
}

func (u *usecase) processNextBatch(ctx context.Context) error {
	batch, err := u.repo.ClaimNextBatch(ctx, u.workerID, u.lockTimeout)
	if err != nil || batch == nil {
		return err
	}

	return u.processClaimedBatch(ctx, batch)
}

var _ Usecase = (*usecase)(nil)

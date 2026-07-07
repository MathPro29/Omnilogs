package handler

import (
	"errors"
	"fmt"
	"net/http"

	"omnilogs-api/dto"
	"omnilogs-api/internal/log_queues/usecase"
	workerprocessor "omnilogs-api/internal/worker"
	"omnilogs-api/responses"
	"omnilogs-api/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	usecase    usecase.Usecase
	processor *workerprocessor.Processor
}

func NewHandler(usecase usecase.Usecase, processor *workerprocessor.Processor) *Handler {
	return &Handler{usecase: usecase, processor: processor}
}

// QueueHandler จัดการการ Enqueue (รับ log ก้อนใหญ่เข้ามาในคิว)
func (h *Handler) QueueHandler(c *gin.Context) {
	var req dto.IngestLogBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationError(c, err)
		return
	}

	batch, err := h.usecase.Enqueue(c.Request.Context(), req)
	if err != nil {
		if err.Error() == "environment does not belong to product" || err.Error() == "product_id is required when environment_id is provided" {
			responses.BadRequest(c, err.Error())
			return
		}
		responses.Error(c, "INTERNAL_ERROR", "Failed to enqueue log batch", err)
		return
	}

	utils.Success(c, http.StatusCreated, batch)
}

// ConsumeHandler ดึง Batch ล่าสุดที่รออยู่มาประมวลผลบนหน่วยความจำทันที (Manual Trigger)
func (h *Handler) ConsumeHandler(c *gin.Context) {
	if h.processor == nil {
		responses.Error(c, "INTERNAL_ERROR", "Worker processor is not configured", nil)
		return
	}

	processed, err := h.processor.RunOnce(c.Request.Context())
	if err != nil {
		responses.Error(c, "INTERNAL_ERROR", "Failed to consume batch", err)
		return
	}
	if !processed {
		responses.NotFound(c, "No pending batches in queue")
		return
	}

	utils.Success(c, http.StatusOK, gin.H{"status": "processed"})
}

// 1. GetBatchItemsHandler ดึงรายการคิวทั้งหมดใน Batch นั้นๆ
func (h *Handler) GetBatchItemsHandler(c *gin.Context) {
	batchID := c.Param("batchId")
	if batchID == "" {
		responses.BadRequest(c, "batchId is required")
		return
	}

	items, err := h.usecase.GetItemsByBatch(c.Request.Context(), batchID)
	if err != nil {
		responses.Error(c, "INTERNAL_ERROR", "Failed to retrieve batch items", err)
		return
	}

	utils.Success(c, http.StatusOK, items)
}

// 2. GetItemHandler ดึงรายละเอียดของ Queue Item เจาะจงตัวเพื่อตรวจ Raw Payload
func (h *Handler) GetItemHandler(c *gin.Context) {
	// แปลง Param จาก string เป็น int64
	itemIDStr := c.Param("itemId")
	var itemID int64
	_, err := fmt.Sscanf(itemIDStr, "%d", &itemID)
	if err != nil || itemID <= 0 {
		responses.BadRequest(c, "invalid itemId")
		return
	}

	item, err := h.usecase.GetItemByID(c.Request.Context(), itemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			responses.NotFound(c, "Queue item not found")
			return
		}
		responses.Error(c, "INTERNAL_ERROR", "Failed to retrieve queue item", err)
		return
	}

	utils.Success(c, http.StatusOK, item)
}

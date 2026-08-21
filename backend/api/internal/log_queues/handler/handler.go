package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"omnilogs-api/dto"
	"omnilogs-api/internal/log_queues/usecase"
	workerprocessor "omnilogs-api/internal/worker"
	"omnilogs-api/responses"
	"omnilogs-api/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	usecase   usecase.Usecase
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

	if productID, exists := c.Get("service_product_id"); exists {
		resolvedProductID := productID.(int)
		if apiKeyID, ok := c.Get("service_api_key_id"); ok {
			value := apiKeyID.(int)
			req.APIKeyID = &value
		}
		if req.ProductID != nil && *req.ProductID != resolvedProductID {
			responses.Forbidden(c, "API key does not belong to requested product")
			return
		}
		req.ProductID = &resolvedProductID
		if environmentID, scoped := c.Get("service_environment_id"); scoped {
			resolvedEnvironmentID := environmentID.(int)
			if req.EnvironmentID != nil && *req.EnvironmentID != resolvedEnvironmentID {
				responses.Forbidden(c, "API key does not belong to requested environment")
				return
			}
			req.EnvironmentID = &resolvedEnvironmentID
		}
		if req.EnvironmentCode == "" {
			req.EnvironmentCode = c.GetHeader("X-Environment-Code")
		}
		if req.EnvironmentCode == "" {
			req.EnvironmentCode = c.GetHeader("X-Environment")
		}
		req.HeaderProjectID = strings.TrimSpace(c.GetHeader("X-Project-ID"))
		req.HeaderProjectCode = strings.TrimSpace(c.GetHeader("X-Project-Code"))
		req.HeaderEnvironment = strings.TrimSpace(req.EnvironmentCode)
		if sourceID, scoped := c.Get("service_source_id"); scoped {
			resolvedSourceID := sourceID.(int)
			if req.SourceID != nil && *req.SourceID != resolvedSourceID {
				responses.Forbidden(c, "API key does not belong to requested source")
				return
			}
			req.SourceID = &resolvedSourceID
		}
		if err := validatePayloadScopeOptionalEnvironment(req.Logs, resolvedProductID, req.EnvironmentID); err != nil {
			responses.Forbidden(c, err.Error())
			return
		}
	}

	batch, err := h.usecase.Enqueue(c.Request.Context(), req)
	if err != nil {
		var limitErr *usecase.PayloadLimitError
		if errors.As(err, &limitErr) {
			utils.ErrorWithDetails(c, http.StatusRequestEntityTooLarge, limitErr.Code, limitErr.Message, gin.H{"maximum": limitErr.Limit, "received": limitErr.Received})
			return
		}
		if errors.Is(err, usecase.ErrInvalidScope) || errors.Is(err, usecase.ErrAmbiguousScope) || errors.Is(err, usecase.ErrUnmappedScope) {
			responses.BadRequest(c, err.Error())
			return
		}
		responses.Error(c, "INTERNAL_ERROR", "Failed to enqueue log batch", err)
		return
	}

	utils.Success(c, http.StatusCreated, batch)
}

func (h *Handler) GetIngestBatchStatusHandler(c *gin.Context) {
	batch, err := h.usecase.GetBatchByID(c.Request.Context(), c.Param("batchId"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		responses.NotFound(c, "batch not found")
		return
	}
	if err != nil {
		responses.Error(c, "INTERNAL_ERROR", "failed to retrieve batch", err)
		return
	}
	productID, _ := c.Get("service_product_id")
	if productID.(int) != *batch.ProductID {
		responses.Forbidden(c, "batch is outside API key product")
		return
	}
	if environmentID, scoped := c.Get("service_environment_id"); scoped && (batch.EnvironmentID == nil || environmentID.(int) != *batch.EnvironmentID) {
		responses.Forbidden(c, "batch is outside API key environment")
		return
	}
	utils.Success(c, http.StatusOK, batch)
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
	// ถ้าไม่มีคิวค้างให้เป็น 404
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

// GetFailedBatchesHandler ดึงรายการ Failed Batches ทั้งหมด (จำกัดจำนวนตาม Config)
func (h *Handler) GetFailedBatchesHandler(c *gin.Context) {
	batches, err := h.usecase.GetFailedBatches(c.Request.Context())
	if err != nil {
		responses.Error(c, "INTERNAL_ERROR", "Failed to retrieve failed batches", err)
		return
	}

	utils.Success(c, http.StatusOK, batches)
}

func validatePayloadScope(logs []dto.IngestLogItemRequest, productID, environmentID int) error {
	return validatePayloadScopeOptionalEnvironment(logs, productID, &environmentID)
}

func validatePayloadScopeOptionalEnvironment(logs []dto.IngestLogItemRequest, productID int, environmentID *int) error {
	for i, log := range logs {
		raw := log.InputPayload
		if len(raw) == 0 {
			raw = log.Data
		}
		if len(raw) == 0 {
			raw = log.Fields
		}
		var payload map[string]any
		if json.Unmarshal(raw, &payload) != nil {
			continue
		}
		for _, value := range payloadScopeCandidates(payload) {
			if value == nil {
				continue
			}
			if id, ok := payloadScopeID(value["product_id"]); ok && id != productID {
				return fmt.Errorf("logs[%d] product_id does not match API key", i)
			}
			if id, ok := payloadScopeID(value["environment_id"]); ok && environmentID != nil && id != *environmentID {
				return fmt.Errorf("logs[%d] environment_id does not match API key", i)
			}
		}
	}
	return nil
}
func payloadScopeCandidates(payload map[string]any) []map[string]any {
	values := []map[string]any{payload, nestedMetadata(payload)}
	if customFields, ok := payload["custom_fields"].(map[string]any); ok {
		values = append(values, customFields)
	}
	if customFields, ok := payload["customFields"].(map[string]any); ok {
		values = append(values, customFields)
	}
	if routing, ok := payload["routing"].(map[string]any); ok {
		if explicit, ok := routing["explicit_hierarchy"].(map[string]any); ok {
			values = append(values, explicit)
		}
		if explicit, ok := routing["explicitHierarchy"].(map[string]any); ok {
			values = append(values, explicit)
		}
	}
	return values
}
func nestedMetadata(payload map[string]any) map[string]any {
	value, _ := payload["metadata"].(map[string]any)
	return value
}
func payloadScopeID(value any) (int, bool) {
	switch v := value.(type) {
	case float64:
		return int(v), v > 0
	case int:
		return v, v > 0
	case json.Number:
		n, err := v.Int64()
		return int(n), err == nil && n > 0
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(v))
		return n, err == nil && n > 0
	}
	return 0, false
}

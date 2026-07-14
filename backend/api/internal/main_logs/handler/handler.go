package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"omnilogs-api/configs"
	"omnilogs-api/internal/main_logs/usecase"
	"omnilogs-api/middleware"
	"omnilogs-api/responses"
	"omnilogs-api/utils"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
)

type Handler struct {
	usecase   usecase.Usecase
	natsQueue *configs.NATSQueue
}

func NewHandler(usecase usecase.Usecase, natsQueue *configs.NATSQueue) *Handler {
	return &Handler{usecase: usecase, natsQueue: natsQueue}
}

func (h *Handler) Search(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		responses.Unauthorized(c, "unauthorized")
		return
	}

	productID, err := strconv.ParseInt(c.Query("product_id"), 10, 64)
	if err != nil || productID <= 0 {
		responses.BadRequest(c, "product_id is required")
		return
	}

	page := parseIntWithDefault(c.Query("page"), 1)
	perPage := parseIntWithDefault(c.Query("per_page"), 20)
	if perPage > 100 {
		perPage = 100
	}

	input := usecase.SearchInput{
		ActorUserID:      userID,
		PlatformAdmin:    middleware.HasAdminPlatformRole(c),
		ProductID:        productID,
		Page:             page,
		PerPage:          perPage,
		EnvironmentID:    parseOptionalInt64(c.Query("environment_id")),
		ProjectID:        parseOptionalInt64(c.Query("project_id")),
		ProjectIDs:       parseCSVInt64(c.Query("project_ids")),
		CategoryID:       parseOptionalInt64(c.Query("category_id")),
		CategoryIDs:      parseCSVInt64(c.Query("category_ids")),
		Level:            stringParam(c.Query("level")),
		LogType:          stringParam(c.Query("log_type")),
		RequestID:        stringParam(c.Query("request_id")),
		TraceID:          stringParam(c.Query("trace_id")),
		CustomFieldPath:  stringParam(c.Query("custom_field_path")),
		CustomFieldValue: stringParam(c.Query("custom_field_value")),
		Keyword:          stringParam(c.Query("keyword")),
	}

	result, err := h.usecase.Search(c.Request.Context(), input, requestID(c), traceID(c), ipAddress(c), userAgent(c))
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			responses.Error(c, "TIMEOUT", "log search timed out", err)
			return
		}
		if errors.Is(err, usecase.ErrInvalidSearchFilter) {
			responses.BadRequest(c, err.Error())
			return
		}
		if errors.Is(err, responses.ErrForbidden) {
			responses.Forbidden(c, "forbidden")
			return
		}
		responses.Error(c, "INTERNAL_ERROR", "failed to search logs", err)
		return
	}

	responses.SuccessWithMeta(c, http.StatusOK, "LOGS_RETRIEVED", result.Items, utils.NewPaginationMeta(page, perPage, result.Total))
}

func (h *Handler) GetByID(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		responses.Unauthorized(c, "unauthorized")
		return
	}

	productID, err := strconv.ParseInt(c.Query("product_id"), 10, 64)
	if err != nil || productID <= 0 {
		responses.BadRequest(c, "product_id is required")
		return
	}

	value, err := h.usecase.FindByID(c.Request.Context(), userID, middleware.HasAdminPlatformRole(c), productID, c.Param("logId"), requestID(c), traceID(c), ipAddress(c), userAgent(c))
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			responses.Error(c, "TIMEOUT", "log detail query timed out", err)
		case errors.Is(err, responses.ErrForbidden):
			responses.Forbidden(c, "forbidden")
		case errors.Is(err, usecase.ErrMainLogNotFound):
			responses.NotFound(c, "main log not found")
		default:
			responses.Error(c, "INTERNAL_ERROR", "failed to retrieve main log", err)
		}
		return
	}

	responses.Success(c, http.StatusOK, "LOG_RETRIEVED", value)
}

func (h *Handler) GetByAudit(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		responses.Unauthorized(c, "unauthorized")
		return
	}

	audit, value, reason, err := h.usecase.FindByAudit(c.Request.Context(), userID, middleware.HasAdminPlatformRole(c), c.Param("auditId"), requestID(c), traceID(c), ipAddress(c), userAgent(c))
	if err != nil {
		if errors.Is(err, responses.ErrForbidden) {
			responses.Forbidden(c, "forbidden")
			return
		}
		responses.Error(c, "INTERNAL_ERROR", "failed to retrieve linked main log", err)
		return
	}

	data := gin.H{
		"audit": audit,
		"main_log": gin.H{
			"available": value != nil,
		},
	}
	if value != nil {
		data["main_log"] = gin.H{"available": true, "data": value}
	} else if reason != "" {
		data["main_log"] = gin.H{"available": false, "reason": reason}
	}

	responses.Success(c, http.StatusOK, "AUDIT_MAIN_LOG_RETRIEVED", data)
}

func (h *Handler) LiveTail(c *gin.Context) {
	_, ok := middleware.CurrentUserID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	productID, err := strconv.ParseInt(c.Query("product_id"), 10, 64)
	if err != nil || productID <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "product_id is required"})
		return
	}

	environmentID := c.Query("environment_id")

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	subject := fmt.Sprintf("omnilogs.logs.live.%d", productID)
	msgChan := make(chan *nats.Msg, 2000)
	sub, err := h.natsQueue.Conn.ChanSubscribe(subject, msgChan)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to subscribe"})
		return
	}
	defer sub.Unsubscribe()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	buffer := make([]map[string]any, 0, 100)

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case msg := <-msgChan:
			var logData map[string]any
			if err := json.Unmarshal(msg.Data, &logData); err == nil {
				if environmentID != "" {
					if envFloat, ok := logData["environment_id"].(float64); ok {
						if strconv.FormatFloat(envFloat, 'f', -1, 64) != environmentID {
							continue
						}
					}
				}
				mapped := mapESDocToMainLogDocument(logData)
				buffer = append(buffer, mapped)
				if len(buffer) >= 100 {
					flushBuffer(c, &buffer)
				}
			}
		case <-ticker.C:
			if len(buffer) > 0 {
				flushBuffer(c, &buffer)
			}
		}
	}
}

func flushBuffer(c *gin.Context, buffer *[]map[string]any) {
	if len(*buffer) == 0 {
		return
	}
	dataBytes, err := json.Marshal(*buffer)
	if err == nil {
		fmt.Fprintf(c.Writer, "data: %s\n\n", string(dataBytes))
		c.Writer.Flush()
	}
	*buffer = make([]map[string]any, 0, 100)
}

func mapESDocToMainLogDocument(source map[string]any) map[string]any {
	result := map[string]any{
		"log_id": source["log_id"],
		"raw":    source,
	}

	result["product_id"] = getInt64(source["product_id"])
	if env, ok := source["environment_id"]; ok {
		result["environment_id"] = getInt64(env)
	}
	if src, ok := source["source_id"]; ok {
		result["source_id"] = getInt64(src)
	}
	if ts, ok := source["@timestamp"]; ok {
		result["timestamp"] = ts
	}

	payload, _ := source["payload"].(map[string]any)
	if payload != nil {
		result["level"] = payload["log_level"]
		result["log_type"] = payload["event_type"]
		result["message"] = payload["message"]
		result["request_id"] = payload["source_request_id"]
		result["trace_id"] = payload["trace_id"]
		result["method"] = payload["request_method"]
		result["path"] = payload["request_path"]
		result["url"] = payload["url"]
		if sc, ok := payload["status_code"]; ok {
			result["status_code"] = getInt64(sc)
		}
		if lat, ok := payload["duration_ms"]; ok {
			result["latency_ms"] = getInt64(lat)
		}
		result["error_code"] = payload["error_code"]
		result["error_message"] = payload["error_message"]
		result["stack_trace"] = payload["stack_trace"]
		result["request_headers"] = payload["request_headers"]
		result["response_headers"] = payload["response_headers"]
		result["request_payload"] = payload["request_payload"]
		result["response_payload"] = payload["response_payload"]
		result["custom_fields"] = payload["custom_fields"]
	}

	return result
}

func getInt64(val any) int64 {
	switch v := val.(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	default:
		return 0
	}
}

func parseIntWithDefault(raw string, fallback int) int {
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func parseOptionalInt64(raw string) *int64 {
	if raw == "" {
		return nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return nil
	}
	return &value
}

func stringParam(raw string) *string {
	if raw == "" {
		return nil
	}
	return &raw
}

func requestID(c *gin.Context) *string {
	value, ok := c.Get(middleware.ContextRequestID)
	if !ok {
		return nil
	}
	raw, ok := value.(string)
	if !ok || raw == "" {
		return nil
	}
	return &raw
}

func traceID(c *gin.Context) *string {
	raw := c.GetHeader("X-Trace-ID")
	if raw == "" {
		raw = c.GetHeader("Traceparent")
	}
	if raw == "" {
		return nil
	}
	return &raw
}

func ipAddress(c *gin.Context) *string {
	raw := c.ClientIP()
	if raw == "" {
		return nil
	}
	return &raw
}

func userAgent(c *gin.Context) *string {
	raw := c.Request.UserAgent()
	if raw == "" {
		return nil
	}
	return &raw
}

func parseCSVInt64(raw string) []int64 {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]int64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if v, err := strconv.ParseInt(part, 10, 64); err == nil && v > 0 {
			result = append(result, v)
		}
	}
	return result
}

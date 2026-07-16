package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"omnilogs-api/internal/main_logs/document"
	"omnilogs-api/middleware"
	"omnilogs-api/responses"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
)

const (
	liveBufferCapacity  = 100
	liveChannelCapacity = 2000
	liveFlushInterval   = 500 * time.Millisecond
)

func (h *Handler) LiveTail(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	productID, err := strconv.ParseInt(c.Query("product_id"), 10, 64)
	if err != nil || productID <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "product_id is required"})
		return
	}
	if err := h.usecase.AuthorizeProductAccess(c.Request.Context(), userID, middleware.HasAdminPlatformRole(c), productID); err != nil {
		if errors.Is(err, responses.ErrForbidden) {
			responses.Forbidden(c, "forbidden")
			return
		}
		slog.ErrorContext(c.Request.Context(), "failed to authorize live log stream", "error", err, "product_id", productID)
		responses.Error(c, "INTERNAL_ERROR", "failed to authorize live log stream", nil)
		return
	}
	if h.natsQueue == nil || h.natsQueue.Conn == nil {
		slog.ErrorContext(c.Request.Context(), "live log stream is not configured")
		responses.Error(c, "INTERNAL_ERROR", "live log stream is not configured", nil)
		return
	}

	messageChannel := make(chan *nats.Msg, liveChannelCapacity)
	subscription, err := h.natsQueue.Conn.ChanSubscribe(fmt.Sprintf("omnilogs.logs.live.%d", productID), messageChannel)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to subscribe to live log stream", "error", err, "product_id", productID)
		responses.Error(c, "INTERNAL_ERROR", "failed to subscribe", nil)
		return
	}
	defer func() {
		if err := subscription.Unsubscribe(); err != nil {
			slog.WarnContext(c.Request.Context(), "failed to unsubscribe from live log stream", "error", err)
		}
	}()

	setLiveHeaders(c)
	ticker := time.NewTicker(liveFlushInterval)
	defer ticker.Stop()

	environmentID := c.Query("environment_id")
	buffer := make([]map[string]any, 0, liveBufferCapacity)
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case message, open := <-messageChannel:
			if !open {
				return
			}
			mapped, ok := mapLiveMessage(message.Data, environmentID)
			if !ok {
				continue
			}
			buffer = append(buffer, mapped)
			if len(buffer) >= liveBufferCapacity && flushLiveBuffer(c, &buffer) != nil {
				return
			}
		case <-ticker.C:
			if len(buffer) > 0 && flushLiveBuffer(c, &buffer) != nil {
				return
			}
		}
	}
}

func setLiveHeaders(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")
}

func mapLiveMessage(data []byte, environmentFilter string) (map[string]any, bool) {
	var source map[string]any
	if err := json.Unmarshal(data, &source); err != nil {
		return nil, false
	}
	logID, _ := source["log_id"].(string)
	value := document.Map(logID, source)
	if environmentFilter != "" {
		if value.EnvironmentID == nil || strconv.FormatInt(*value.EnvironmentID, 10) != environmentFilter {
			return nil, false
		}
	}

	result := map[string]any{"log_id": source["log_id"], "raw": source, "product_id": value.ProductID}
	if _, exists := source["environment_id"]; exists {
		result["environment_id"] = pointerValue(value.EnvironmentID)
	}
	if _, exists := source["source_id"]; exists {
		result["source_id"] = pointerValue(value.SourceID)
	}
	if timestamp, exists := source["@timestamp"]; exists {
		result["timestamp"] = timestamp
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
		if _, exists := payload["status_code"]; exists {
			result["status_code"] = pointerValue(value.StatusCode)
		}
		if _, exists := payload["duration_ms"]; exists {
			result["latency_ms"] = pointerValue(value.LatencyMs)
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
	return result, true
}

func pointerValue(value *int64) any {
	if value == nil {
		return int64(0)
	}
	return *value
}

func flushLiveBuffer(c *gin.Context, buffer *[]map[string]any) error {
	data, err := json.Marshal(*buffer)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", data); err != nil {
		return err
	}
	c.Writer.Flush()
	*buffer = make([]map[string]any, 0, liveBufferCapacity)
	return nil
}

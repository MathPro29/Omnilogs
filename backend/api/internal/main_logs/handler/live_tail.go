package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"omnilogs-api/middleware"
	"omnilogs-api/responses"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
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
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to authorize log stream"})
		return
	}
	if h.natsQueue == nil || h.natsQueue.Conn == nil {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "live log stream is unavailable"})
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
			if msg == nil {
				continue
			}
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

package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// PerformanceLogger records only slow or failed requests to avoid noisy logs.
func PerformanceLogger(slowThreshold time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		duration := time.Since(startedAt)
		if duration < slowThreshold && c.Writer.Status() < 500 {
			return
		}

		requestID, _ := c.Get(ContextRequestID)
		slog.Warn("slow or failed api request",
			"endpoint", c.FullPath(),
			"method", c.Request.Method,
			"request_id", requestID,
			"status", c.Writer.Status(),
			"duration_ms", duration.Milliseconds(),
			"slow_threshold_ms", slowThreshold.Milliseconds(),
		)
	}
}

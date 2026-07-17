package handler

import (
	"strconv"
	"strings"

	"omnilogs-api/middleware"

	"github.com/gin-gonic/gin"
)

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

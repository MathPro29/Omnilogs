package handler

import (
	"strconv"
	"strings"

	"omnilogs-api/middleware"

	"github.com/gin-gonic/gin"
)

func parseOptionalInt64(raw string) *int64 {
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
	return nonEmptyString(value)
}

func traceID(c *gin.Context) *string {
	value := c.GetHeader("X-Trace-ID")
	if value == "" {
		value = c.GetHeader("Traceparent")
	}
	return nonEmptyString(value)
}

func ipAddress(c *gin.Context) *string { return nonEmptyString(c.ClientIP()) }

func userAgent(c *gin.Context) *string { return nonEmptyString(c.Request.UserAgent()) }

func nonEmptyString(value any) *string {
	text, ok := value.(string)
	if !ok || text == "" {
		return nil
	}
	return &text
}

func parseCSVInt64(raw string) []int64 {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]int64, 0, len(parts))
	seen := make(map[int64]struct{}, len(parts))
	for _, part := range parts {
		value, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil || value <= 0 {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

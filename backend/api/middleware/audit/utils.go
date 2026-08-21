package audit

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"omnilogs-api/middleware"

	"github.com/gin-gonic/gin"
)

func auditUserID(c *gin.Context) *int64 {
	if userID, ok := middleware.CurrentUserID(c); ok {
		value := int64(userID)
		return &value
	}
	return nil
}

func parseStringToInt64(s string) *int64 {
	if value, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64); err == nil && value > 0 {
		return &value
	}
	return nil
}

func auditProductID(c *gin.Context, requestBody []byte) *int64 {
	if id := parseStringToInt64(c.Param("productId")); id != nil {
		return id
	}
	if id := parseStringToInt64(c.Query("product_id")); id != nil {
		return id
	}

	if len(requestBody) == 0 {
		return nil
	}

	var body map[string]any
	if err := json.Unmarshal(requestBody, &body); err != nil {
		return nil
	}

	switch typed := body["product_id"].(type) {
	case float64:
		if value := int64(typed); value > 0 {
			return &value
		}
	case string:
		return parseStringToInt64(typed)
	}

	return nil
}

func stringPointer(value string) *string {
	if value = strings.TrimSpace(value); value != "" {
		return &value
	}
	return nil
}

func requestID(c *gin.Context) *string {
	if value, ok := c.Get(middleware.ContextRequestID); ok {
		if strValue, ok := value.(string); ok {
			return stringPointer(strValue)
		}
	}
	return nil
}

func traceID(c *gin.Context) *string {
	if value := stringPointer(c.GetHeader("X-Trace-ID")); value != nil {
		return value
	}
	return stringPointer(c.GetHeader("Traceparent"))
}

func readRequestBody(c *gin.Context) []byte {
	if c.Request == nil || c.Request.Body == nil {
		return nil
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Request.Body = io.NopCloser(bytes.NewReader(nil))
		return nil
	}

	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	return body
}

func newAuditUUID() string {
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
	hex.Encode(dst[24:36], bytes[10:16])
	return string(dst)
}

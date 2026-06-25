package middleware

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"omnilogs-api/models"
	"omnilogs-api/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var auditRedactedKeys = map[string]struct{}{
	"authorization":   {},
	"password":        {},
	"confirmpassword": {},
	"newpassword":     {},
	"refreshtoken":    {},
	"accesstoken":     {},
	"token":           {},
	"apikey":          {},
	"secret":          {},
	"clientsecret":    {},
	"privatekey":      {},
	"passphrase":      {},
	"credential":      {},
	"credentials":     {},
}

func normalizeKey(key string) string {
	k := strings.ToLower(strings.TrimSpace(key))
	k = strings.ReplaceAll(k, "_", "")
	k = strings.ReplaceAll(k, "-", "")
	return k
}

func isAuditRedactedKey(key string) bool {
	normKey := normalizeKey(key)
	_, ok := auditRedactedKeys[normKey]
	return ok
}

func AuditLogger(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipAudit(c.Request.Method, c.Request.URL.Path) {
			c.Next()
			return
		}

		requestBody := readRequestBody(c)

		c.Next()

		action := auditAction(c)
		if action == "" {
			return
		}

		statusCode := c.Writer.Status()

		var reason *string
		if val, exists := c.Get("audit_reason"); exists {
			if rStr, ok := val.(string); ok {
				reason = stringPointer(rStr)
			}
		}
		if reason == nil {
			reason = stringPointer(http.StatusText(statusCode))
		}

		record := models.SystemAuditLog{
			AuditID:    newAuditUUID(),
			UserID:     auditUserID(c),
			ProductID:  auditProductID(c),
			Action:     action,
			Method:     stringPointer(c.Request.Method),
			URL:        stringPointer(c.Request.URL.RequestURI()),
			IPAddress:  stringPointer(c.ClientIP()),
			QueryJSON:  sanitizeJSONBytes(queryJSON(c.Request.URL.Query())),
			Payload:    sanitizeJSONBytes(requestBody),
			Reason:     reason,
			StatusCode: &statusCode,
		}

		// [KEY:GO ROUTINE] run in background safely using helper function
		utils.SafeGo(func() {
			bgDB := utils.GetBackgroundDB(db)
			if bgDB == nil {
				log.Printf("[ERROR] Failed to save audit log: DB context is nil")
				return
			}
			// Save the audit log using the background DB context
			if err := bgDB.Create(&record).Error; err != nil {
				log.Printf("[DATABASE_FAIL / ฐานข้อมูลล่ม] Failed to save audit log: %v, Data: %+v", err, record)
			}
		})
	}
}

func shouldSkipAudit(method, path string) bool {
	if method == http.MethodOptions {
		return true
	}
	return strings.HasPrefix(path, "/health") || strings.HasPrefix(path, "/swagger")
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

func auditAction(c *gin.Context) string {
	fullPath := c.FullPath()
	if fullPath == "" {
		fullPath = c.Request.URL.Path
	}

	trimmed := strings.Trim(fullPath, "/")
	if trimmed == "" {
		return strings.ToUpper(c.Request.Method)
	}

	parts := strings.Split(trimmed, "/")
	if len(parts) >= 3 && parts[0] == "api" && parts[1] == "v1" {
		parts = parts[2:]
	}

	resource := "SYSTEM"
	for _, part := range parts {
		if part == "" || strings.HasPrefix(part, "{") || strings.HasPrefix(part, ":") {
			continue
		}
		resource = strings.ToUpper(strings.ReplaceAll(part, "-", "_"))
		break
	}

	target := "UNKNOWN"
	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if part == "" || strings.HasPrefix(part, "{") || strings.HasPrefix(part, ":") {
			continue
		}
		target = strings.ToUpper(strings.ReplaceAll(part, "-", "_"))
		break
	}

	action := methodAction(c.Request.Method)
	if target == resource {
		return resource + "." + action
	}
	return resource + "." + target + "." + action
}

func auditUserID(c *gin.Context) *int {
	userID, ok := CurrentUserID(c)
	if !ok {
		return nil
	}
	value := int(userID)
	return &value
}

func auditProductID(c *gin.Context) *int {
	productID := strings.TrimSpace(c.Param("productId"))
	if productID == "" {
		return nil
	}
	value, err := strconv.Atoi(productID)
	if err != nil || value <= 0 {
		return nil
	}
	return &value
}

func stringPointer(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func queryJSON(values map[string][]string) []byte {
	if len(values) == 0 {
		return nil
	}
	payload := make(map[string]any, len(values))
	for key, value := range values {
		if len(value) == 1 {
			payload[key] = value[0]
			continue
		}
		payload[key] = value
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	return raw
}

func sanitizeJSONBytes(raw []byte) []byte {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return nil
	}

	var payload any
	if err := json.Unmarshal(raw, &payload); err != nil {
		fallback, marshalErr := json.Marshal(map[string]string{"raw": string(raw)})
		if marshalErr != nil {
			return nil
		}
		return fallback
	}

	sanitizeJSONValue(&payload)
	sanitized, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	return sanitized
}

func sanitizeJSONValue(value *any) {
	switch typed := (*value).(type) {
	case map[string]any:
		for key, child := range typed {
			if isAuditRedactedKey(key) {
				typed[key] = "[REDACTED]"
				continue
			}
			sanitizeJSONValue(&child)
			typed[key] = child
		}
	case []any:
		for i := range typed {
			child := typed[i]
			sanitizeJSONValue(&child)
			typed[i] = child
		}
	}
}

func methodAction(method string) string {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case http.MethodGet:
		return "READ"
	case http.MethodPost:
		return "CREATE"
	case http.MethodPut, http.MethodPatch:
		return "UPDATE"
	case http.MethodDelete:
		return "DELETE"
	default:
		return strings.ToUpper(strings.TrimSpace(method))
	}
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

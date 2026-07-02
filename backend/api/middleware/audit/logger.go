package audit

import (
	"log"
	"net/http"
	"strings"

	"omnilogs-api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Logger(db *gorm.DB, encryptionKey string) gin.HandlerFunc {
	_ = encryptionKey

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
		result := models.AuditResultSuccess
		if statusCode >= 400 && statusCode < 500 {
			result = models.AuditResultDenied
		} else if statusCode >= 500 {
			result = models.AuditResultFailed
		}

		resourceType, resourceID := auditResource(c)
		metadataJSON := buildAuditMetadata(c, requestBody, statusCode)
		record := models.SystemAuditLog{
			AuditID:      newAuditUUID(),
			ActorUserID:  auditUserID(c),
			ProductID:    auditProductID(c, requestBody),
			Action:       action,
			ResourceType: resourceType,
			ResourceID:   resourceID,
			RequestID:    requestID(c),
			TraceID:      traceID(c),
			Method:       stringPointer(c.Request.Method),
			Path:         stringPointer(c.Request.URL.Path),
			Result:       result,
			Metadata:     metadataJSON,
			IPAddress:    stringPointer(c.ClientIP()),
			UserAgent:    stringPointer(c.Request.UserAgent()),
		}

		if err := db.WithContext(c.Request.Context()).Create(&record).Error; err != nil {
			log.Printf("[ERROR] Failed to save audit log: %v", err)
		}
	}
}

func shouldSkipAudit(method, path string) bool {
	if method == http.MethodOptions {
		return true
	}
	return strings.HasPrefix(path, "/health") || strings.HasPrefix(path, "/swagger") || strings.HasPrefix(path, "/api/v1/queues")
}

package audit

import (
	"context"
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

		if !shouldRecordAuditEvent(c) {
			return
		}

		action := auditAction(c)
		if action == "" {
			return
		}

		auditID := newAuditUUID()
		statusCode := c.Writer.Status()
		result := models.AuditResultSuccess
		if statusCode >= 400 && statusCode < 500 {
			result = models.AuditResultDenied
		} else if statusCode >= 500 {
			result = models.AuditResultFailed
		}

		resourceType, resourceID := auditResource(c)
		metadataResult := buildAuditMetadata(c, requestBody, statusCode, auditID, encryptionKey)
		record := models.SystemAuditLog{
			AuditID:      auditID,
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
			Metadata:     metadataResult.Metadata,
			IPAddress:    stringPointer(c.ClientIP()),
			UserAgent:    stringPointer(c.Request.UserAgent()),
		}

		auditCtx := context.WithoutCancel(c.Request.Context())
		if err := db.WithContext(auditCtx).Create(&record).Error; err != nil {
			log.Printf("[ERROR] Failed to save audit log: %v", err)
			return
		}
		for _, secret := range metadataResult.Secrets {
			if secret.EncryptedValue == "" {
				continue
			}
			if err := db.WithContext(auditCtx).Create(&secret).Error; err != nil {
				log.Printf("[ERROR] Failed to save audit secret: %v", err)
			}
		}
	}
}

func shouldSkipAudit(method, path string) bool {
	if method == http.MethodOptions {
		return true
	}
	return strings.HasPrefix(path, "/health") || strings.HasPrefix(path, "/swagger")
}

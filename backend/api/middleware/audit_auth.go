package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"omnilogs-api/models"
	"omnilogs-api/responses"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuditVisibilityMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := CurrentUserID(c)
		if !ok {
			responses.Unauthorized(c, "unauthorized")
			c.Abort()
			return
		}

		if HasAdminPlatformRole(c) {
			c.Set("audit_scope_global", true)
			c.Next()
			return
		}

		var productIDs []int
		err := db.Model(&models.ProductMembership{}).
			Where("user_id = ? AND is_active = true AND (expires_at IS NULL OR expires_at > ?)", userID, time.Now()).
			Pluck("product_id", &productIDs).Error
		if err != nil {
			responses.Error(c, "INTERNAL_ERROR", "failed to check product memberships", err)
			c.Abort()
			return
		}

		c.Set("audit_scope_global", false)
		c.Set("audit_scope_products", productIDs)
		c.Next()
	}
}

func AuditScope(c *gin.Context) (bool, []int) {
	globalVal, existsGlobal := c.Get("audit_scope_global")
	productsVal, existsProducts := c.Get("audit_scope_products")

	var global bool
	if existsGlobal {
		global, _ = globalVal.(bool)
	} else {
		global = HasAdminPlatformRole(c)
	}

	var productIDs []int
	if existsProducts {
		productIDs, _ = productsVal.([]int)
	}

	return global, productIDs
}

func CanReadAuditProduct(c *gin.Context, productID *int64) bool {
	global, allowedProductIDs := AuditScope(c)
	if global {
		return true
	}
	if productID == nil {
		return false
	}
	pid := int(*productID)
	for _, id := range allowedProductIDs {
		if id == pid {
			return true
		}
	}
	return false
}

func ProductAPIKeyAuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKeyStr := c.GetHeader("X-API-Key")
		if apiKeyStr == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
					apiKeyStr = parts[1]
				} else {
					apiKeyStr = authHeader
				}
			}
		}

		apiKeyStr = strings.TrimSpace(apiKeyStr)
		if apiKeyStr == "" {
			responses.Unauthorized(c, "API key is required")
			c.Abort()
			return
		}

		if !strings.HasPrefix(apiKeyStr, "omni_") {
			responses.Unauthorized(c, "invalid API key prefix")
			c.Abort()
			return
		}

		// Compute hash of full key
		sum := sha256.Sum256([]byte(apiKeyStr))
		keyHash := hex.EncodeToString(sum[:])

		var apiKey models.ProductAPIKey
		err := db.Where("key_hash = ? AND is_active = true", keyHash).First(&apiKey).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				responses.Unauthorized(c, "invalid API key")
				c.Abort()
				return
			}
			responses.Error(c, "INTERNAL_ERROR", "failed to authenticate API key", err)
			c.Abort()
			return
		}

		// Check expiration
		now := time.Now()
		if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(now) {
			responses.Unauthorized(c, "API key has expired")
			c.Abort()
			return
		}

		// Check revoked
		if apiKey.RevokedAt != nil && apiKey.RevokedAt.Before(now) {
			responses.Unauthorized(c, "API key has been revoked")
			c.Abort()
			return
		}

		// Update last used time
		db.Model(&apiKey).Update("last_used_at", now)

		c.Set("service_product_id", apiKey.ProductID)
		if apiKey.EnvironmentID != nil {
			c.Set("service_environment_id", *apiKey.EnvironmentID)
		}
		c.Next()
	}
}

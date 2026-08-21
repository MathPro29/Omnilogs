package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.
// It logs the panic error and stack trace using slog.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Retrieve Request ID if present
				var requestID string
				if val, exists := c.Get(ContextRequestID); exists {
					if idStr, ok := val.(string); ok {
						requestID = idStr
					}
				}

				slog.Error("panic recovered",
					"error", err,
					"request_id", requestID,
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"query", c.Request.URL.RawQuery,
					"ip", c.ClientIP(),
					"stack", string(debug.Stack()),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal Server Error",
				})
			}
		}()
		c.Next()
	}
}

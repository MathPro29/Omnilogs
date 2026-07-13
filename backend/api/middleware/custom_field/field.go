package customfield

import "github.com/gin-gonic/gin"

func CustomFieldMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

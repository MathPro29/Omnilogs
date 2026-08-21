package routes

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealthRoutes(router *gin.Engine, readyCheck func(context.Context) error) {
	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"status": "live",
			},
		})
	})

	router.GET("/health/ready", func(c *gin.Context) {
		if err := readyCheck(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "DATABASE_UNAVAILABLE",
					"message": "database ping failed",
					"status":  http.StatusServiceUnavailable,
					"details": err.Error(),
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"status": "ready",
			},
		})
	})
}

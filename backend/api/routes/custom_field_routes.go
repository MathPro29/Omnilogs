package routes

import (
	"time"

	"omnilogs-api/configs"
	"omnilogs-api/internal/custom_fields/handler"
	"omnilogs-api/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CustomFieldRoutes(router *gin.Engine, db *gorm.DB, env *configs.Env) {
	handler := handler.NewHandler(db, env)

	customFields := router.Group("/api/v1/custom-fields")
	customFields.Use(middleware.UserAuthMiddleware(env.JWTSecret))
	customFields.Use(middleware.RequestTimeout(time.Duration(env.APIRequestTimeoutSeconds) * time.Second))

	customFields.POST("/parse-json", handler.ParseJSON)
	customFields.GET("/favorites", handler.ListFavoriteJSONFields)
	customFields.POST("/favorites", handler.FavoriteJSONField)
	customFields.PUT("/favorites/reorder", handler.ReorderFavoriteJSONFields)
	customFields.DELETE("/favorites/:id", handler.RemoveFavoriteJSONField)
	customFields.DELETE("/bulk/delete", handler.BulkDeleteCustomFields)
	customFields.GET("/", handler.GetAllCustomFields)
	customFields.POST("/", handler.CreateCustomField)
	customFields.GET("/:id", handler.GetCustomFieldByID)
	customFields.PUT("/:id", handler.UpdateCustomField)
	customFields.DELETE("/:id", handler.DeleteCustomField)
	customFields.GET("/:id/options", handler.ListOptions)
	customFields.POST("/:id/options", handler.CreateOption)
	customFields.PUT("/:id/options/:optionId", handler.UpdateOption)
	customFields.DELETE("/:id/options/:optionId", handler.DeleteOption)
	customFields.GET("/:id/value-sources", handler.ListValueSources)
	customFields.POST("/:id/value-sources", handler.CreateValueSource)
	customFields.PUT("/:id/value-sources/:sourceId", handler.UpdateValueSource)
}

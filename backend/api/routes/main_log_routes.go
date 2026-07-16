package routes

import (
	"omnilogs-api/configs"
	mainlogmodule "omnilogs-api/internal/main_logs/module"
	"omnilogs-api/middleware"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func MainLogRoutes(router *gin.Engine, db *gorm.DB, esClient *elasticsearch.Client, env *configs.Env, natsQueue *configs.NATSQueue) {
	handler := mainlogmodule.NewHandler(db, esClient, env.DataEncryptionKey, natsQueue)

	group := router.Group("/api/v1/logs")
	group.Use(middleware.UserAuthMiddleware(env.JWTSecret))

	group.GET("", handler.Search)
	group.GET("/live", handler.LiveTail)
	group.GET("/:logId", handler.GetByID)
}

package routes

import (
	"time"

	"omnilogs-api/configs"
	mainlogmodule "omnilogs-api/internal/main_logs/module"
	"omnilogs-api/middleware"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func MainLogRoutes(router *gin.Engine, db *gorm.DB, esClient *elasticsearch.Client, env *configs.Env, natsQueue *configs.NATSQueue) {
	handler := mainlogmodule.NewHandler(
		db,
		esClient,
		env.DataEncryptionKey,
		time.Duration(env.ElasticsearchQueryTimeoutSeconds)*time.Second,
		time.Duration(env.ElasticsearchSlowQueryThresholdMS)*time.Millisecond,
	)

	group := router.Group("/api/v1/logs")
	group.Use(middleware.UserAuthMiddleware(env.JWTSecret))

	timed := group.Group("")
	timed.Use(middleware.RequestTimeout(time.Duration(env.APIRequestTimeoutSeconds) * time.Second))
	timed.GET("", handler.Search)
	timed.POST("/search", handler.SearchV2)
	timed.GET("/:logId", handler.GetByID)
}

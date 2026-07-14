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

func MainLogRoutes(router *gin.Engine, db *gorm.DB, esClient *elasticsearch.Client, env *configs.Env) {
	natsQueue, err := configs.ConnectNATS(env)
	if err != nil {
		panic(err)
	}

	handler := mainlogmodule.NewHandler(
		db,
		esClient,
		env.DataEncryptionKey,
		natsQueue,
		time.Duration(env.ElasticsearchQueryTimeoutSeconds)*time.Second,
		time.Duration(env.ElasticsearchSlowQueryThresholdMS)*time.Millisecond,
	)

	group := router.Group("/api/v1/logs")
	group.Use(middleware.UserAuthMiddleware(env.JWTSecret))

	group.GET("/live", handler.LiveTail)

	timed := group.Group("")
	timed.Use(middleware.RequestTimeout(time.Duration(env.APIRequestTimeoutSeconds) * time.Second))
	timed.GET("", handler.Search)
	timed.GET("/:logId", handler.GetByID)
}

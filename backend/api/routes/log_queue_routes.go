package routes

import (
	"omnilogs-api/configs"
	logqueueshandler "omnilogs-api/internal/log_queues/handler"
	logqueuesrepo "omnilogs-api/internal/log_queues/repository"
	logqueuesusecase "omnilogs-api/internal/log_queues/usecase"
	workerprocessor "omnilogs-api/internal/worker"
	"omnilogs-api/middleware"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func LogQueueRoutes(router *gin.Engine, db *gorm.DB, env *configs.Env, esClient *elasticsearch.Client, natsQueue *configs.NATSQueue) {
	repo := logqueuesrepo.NewRepository(db)
	use := logqueuesusecase.NewUsecase(repo, natsQueue)

	processor := workerprocessor.NewProcessor(db, esClient, env.DataEncryptionKey, natsQueue)
	handler := logqueueshandler.NewHandler(use, processor)

	queues := router.Group("/api/v1/queues")
	queues.Use(middleware.UserAuthMiddleware(env.JWTSecret))

	queues.POST("", handler.QueueHandler)

	// Product and service backends ingest with their scoped API key and do not
	// need an interactive user session.
	router.POST("/api/v1/ingest/logs", middleware.ProductAPIKeyAuthMiddleware(db), handler.QueueHandler)
	queues.POST("/consume", handler.ConsumeHandler)
	queues.GET("/batches/:batchId/items", handler.GetBatchItemsHandler)
	queues.GET("/items/:itemId", handler.GetItemHandler)
	queues.GET("/failed", handler.GetFailedBatchesHandler)
}

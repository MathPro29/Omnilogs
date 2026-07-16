package routes

import (
	"omnilogs-api/configs"
	auditmodule "omnilogs-api/internal/audit_logs/module"
	auditsecretmodule "omnilogs-api/internal/audit_secret/module"
	mainlogmodule "omnilogs-api/internal/main_logs/module"
	"omnilogs-api/middleware"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuditRoutes(router *gin.Engine, db *gorm.DB, esClient *elasticsearch.Client, env *configs.Env) {
	auditHandler := auditmodule.NewHandler(db)
	auditSecretHandler := auditsecretmodule.NewHandler(db, env.DataEncryptionKey)
	mainLogHandler := mainlogmodule.NewHandler(db, esClient, env.DataEncryptionKey, nil)

	group := router.Group("/api/v1/audit-logs")
	group.Use(middleware.UserAuthMiddleware(env.JWTSecret))

	group.GET("", auditHandler.List)
	group.GET("/secrets/requests", auditSecretHandler.ListPendingRequests)
	group.GET("/:auditId", auditHandler.GetByID)
	group.GET("/:auditId/main-log", mainLogHandler.GetByAudit)
	group.GET("/:auditId/secrets/requests", auditSecretHandler.ListRequests)
	group.GET("/:auditId/secrets", auditSecretHandler.ListSecrets)
	group.POST("/:auditId/secrets/requests", auditSecretHandler.CreateRequest)
	group.POST("/:auditId/secrets/requests/:requestId/review", auditSecretHandler.ReviewRequest)
	group.GET("/:auditId/secrets/history", auditSecretHandler.ListHistory)
	group.POST("/:auditId/secrets/reveal", auditSecretHandler.RevealValue)
}

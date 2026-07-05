package routes

import (
	"omnilogs-api/configs"
	auditmodule "omnilogs-api/internal/audit_logs/module"
	mainlogmodule "omnilogs-api/internal/main_logs/module"
	"omnilogs-api/middleware"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuditRoutes(router *gin.Engine, db *gorm.DB, esClient *elasticsearch.Client, env *configs.Env) {
	auditHandler := auditmodule.NewHandler(db)
	mainLogHandler := mainlogmodule.NewHandler(db, esClient, env.DataEncryptionKey)

	group := router.Group("/api/v1/audit-logs")
	group.Use(middleware.UserAuthMiddleware(env.JWTSecret))

	group.GET("", auditHandler.List)
	group.GET("/:auditId", auditHandler.GetByID)
	group.GET("/:auditId/main-log", mainLogHandler.GetByAudit)
}

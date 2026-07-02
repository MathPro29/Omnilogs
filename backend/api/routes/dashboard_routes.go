package routes

import (
	"omnilogs-api/configs"
	"omnilogs-api/internal/dashboard/handler"
	"omnilogs-api/internal/dashboard/repository"
	"omnilogs-api/internal/dashboard/usecase"
	"omnilogs-api/middleware"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DashboardRoutes registers endpoints for log monitoring
func DashboardRoutes(router *gin.Engine, db *gorm.DB, esClient *elasticsearch.Client, env *configs.Env) {
	repo := repository.NewRepository(db, esClient)
	uc := usecase.NewUsecase(repo)
	dashHandler := handler.NewHandler(uc)

	group := router.Group("/api/v1/dashboard")
	group.Use(middleware.UserAuthMiddleware(env.JWTSecret))

	group.GET("/logs", dashHandler.GetLogs)
	group.GET("/stats", dashHandler.GetLogStats)
	group.GET("/logs/:index/:logId", dashHandler.GetLogDetail)
	group.GET("/audit-logs", dashHandler.GetAuditLogs)
}

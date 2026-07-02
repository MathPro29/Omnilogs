package routes

import (
	"omnilogs-api/configs"
	envhandler "omnilogs-api/internal/environment/handler"
	envrepo "omnilogs-api/internal/environment/repository"
	envusecase "omnilogs-api/internal/environment/usecase"
	"omnilogs-api/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func EnvironmentRoutes(router *gin.Engine, db *gorm.DB, env *configs.Env) {
	handler := envhandler.NewHandler(envusecase.NewUsecase(envrepo.NewRepository(db)))

	products := router.Group("/api/v1/products")
	products.Use(middleware.UserAuthMiddleware(env.JWTSecret))

	products.POST("/:productId/environments", handler.CreateEnvironment)
	products.GET("/:productId/environments", handler.ListEnvironments)
	products.PATCH("/:productId/environments/:environmentId", handler.UpdateEnvironment)
	products.DELETE("/:productId/environments/:environmentId", handler.DeleteEnvironment)
}

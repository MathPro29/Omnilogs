package routes

import (
	"omnilogs-api/configs"
	authhandler "omnilogs-api/internal/auth/handler"
	authrepo "omnilogs-api/internal/auth/repository"
	authusecase "omnilogs-api/internal/auth/usecase"
	"omnilogs-api/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthRoutes(router *gin.Engine, db *gorm.DB, env *configs.Env) {
	repository := authrepo.NewRepository(db)
	usecase := authusecase.NewUsecase(
		repository,
		env.JWTSecret,
		env.AccessTokenExpireSeconds,
		env.RefreshTokenExpireSeconds,
	)
	handler := authhandler.NewHandler(usecase)

	// This is the shared authentication entry point for every platform role.
	auth := router.Group("/auth")
	auth.Use(middleware.AuthRateLimit(env))

	auth.POST("/register", handler.Register)
	auth.POST("/login", handler.Login)
	auth.POST("/refresh-token", handler.RefreshToken)
	auth.GET("/me", middleware.UserAuthMiddleware(env.JWTSecret), handler.Me)

}

func RegisterAuthRoutes(router *gin.Engine, db *gorm.DB, env *configs.Env) {
	AuthRoutes(router, db, env)
	AdminAuthRoutes(router, db, env)
	ProductRoutes(router, db, env)
}

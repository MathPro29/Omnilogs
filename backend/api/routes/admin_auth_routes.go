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

func AdminAuthRoutes(router *gin.Engine, db *gorm.DB, env *configs.Env) {
	repository := authrepo.NewRepository(db)
	usecase := authusecase.NewUsecase(
		repository,
		env.JWTSecret,
		env.AccessTokenExpireSeconds,
		env.RefreshTokenExpireSeconds,
	)
	handler := authhandler.NewHandler(usecase)

	// Authentication is shared through /auth/login. Only privileged platform
	// roles may enter routes under /admin.
	adminProtected := router.Group("/admin")
	adminProtected.Use(middleware.AuthRateLimit(env))
	adminProtected.Use(middleware.UserAuthMiddleware(env.JWTSecret))
	adminProtected.Use(middleware.RequireAdminPlatformRole())

	adminProtected.PUT("/give-admin", handler.GiveAdminAccess)
	adminProtected.GET("/show-users", handler.ListAllUsers)
	// ERROR
	// adminProtected.PUT("/edit-user", handler.EditUserRole)
}

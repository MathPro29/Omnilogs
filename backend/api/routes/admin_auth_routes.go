package routes

import (
	"omnilogs-api/configs"
	authmodule "omnilogs-api/internal/auth/module"
	"omnilogs-api/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AdminAuthRoutes(router *gin.Engine, db *gorm.DB, env *configs.Env) {
	handler := authmodule.NewHandler(db, env)

	// Authentication is shared through /auth/login. Only privileged platform
	// roles may enter routes under /admin.
	adminProtected := router.Group("/api/v1/admin")
	adminProtected.Use(middleware.AuthRateLimit(env))
	adminProtected.Use(middleware.UserAuthMiddleware(env.JWTSecret))
	adminProtected.Use(middleware.RequireAdminPlatformRole())

	adminProtected.PUT("/give-admin", handler.GiveAdminAccess)
	adminProtected.GET("/show-users", handler.ListAllUsers)
	// ERROR
	// adminProtected.PUT("/edit-user", handler.EditUserRole)
}

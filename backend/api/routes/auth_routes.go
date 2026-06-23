package routes

import (
	"omnilogs-api/configs"
	authmodule "omnilogs-api/internal/auth/module"
	"omnilogs-api/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthRoutes(router *gin.Engine, db *gorm.DB, env *configs.Env) {
	handler := authmodule.NewHandler(db, env)

	// This is the shared authentication entry point for every platform role.
	auth := router.Group("/api/v1/auth")
	auth.Use(middleware.AuthRateLimit(env))

	auth.POST("/register", handler.Register)
	auth.POST("/login", handler.Login)
	auth.POST("/refresh-token", handler.RefreshToken)
	auth.POST("/logout", handler.Logout)
	auth.GET("/me", middleware.UserAuthMiddleware(env.JWTSecret), handler.Me)
	auth.POST("/forgot-password", handler.ForgotPassword)
	auth.POST("/reset-password", handler.ResetPassword)

}

func RegisterAuthRoutes(router *gin.Engine, db *gorm.DB, env *configs.Env) {
	AuthRoutes(router, db, env)
	AdminAuthRoutes(router, db, env)
	ProductRoutes(router, db, env)
	EnvironmentRoutes(router, db, env)
	ScopeRoutes(router, db, env)
}

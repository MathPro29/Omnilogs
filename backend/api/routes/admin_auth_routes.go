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
	//Edit user
	adminProtected.PUT("/edit-user", handler.EditUserRole)
	//See Details
	adminProtected.GET("/user/:id", handler.GetUserByID)
	//Delete User
	adminProtected.DELETE("/user/:id", handler.DeleteUserByID)

	// Standard CRUD routes (matching frontend client calls)
	usersProtected := router.Group("/api/v1/users")
	usersProtected.Use(middleware.AuthRateLimit(env))
	usersProtected.Use(middleware.UserAuthMiddleware(env.JWTSecret))
	usersProtected.Use(middleware.RequireAdminPlatformRole())

	usersProtected.POST("", handler.CreateUser)
	usersProtected.GET("/:id", handler.GetUserByID)
	usersProtected.PUT("/:id", handler.UpdateUser)
	usersProtected.DELETE("/:id", handler.DeleteUserByID)
}

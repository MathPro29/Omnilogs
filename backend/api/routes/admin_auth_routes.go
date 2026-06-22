package routes

import (

	authhandler "omnilogs-api/internal/auth/handler"
	authrepo "omnilogs-api/internal/auth/repository"
	authusecase "omnilogs-api/internal/auth/usecase"
	

	"omnilogs-api/middleware"
	"omnilogs-api/configs"

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


	adminAuthRoutes := router.Group("/admin")
	adminAuthRoutes.Use(middleware.AuthRateLimit(env))

	adminAuthRoutes.POST("/login", handler.LoginAdmin)

	// Admin Dashboard
	dashboardrepo := dashrepo.NewRepository(db)
	dashboardusecase := dashboardusecase.NewUsecase(dashboardrepo)
	dashboardhandler := dashboardhandler.NewHandler(dashboardusecase)

	adminDashboard := router.Group("/admin/dashboard")
	adminDashboard.Use(middleware.AuthRateLimit(env))
	adminDashboard.Use(middleware.UserAuthMiddleware(env.JWTSecret))
	adminDashboard.Use(middleware.RequireAdminOrSuperadmin())
	adminDashboard.GET("/", dashboardhandler.AdminDashboard)

	adminProtected := adminAuthRoutes.Group("")
	adminProtected.Use(middleware.UserAuthMiddleware(env.JWTSecret))
	adminProtected.Use(middleware.RequireAdminOrSuperadmin())

	adminProtected.GET("/me", handler.Me)
	adminProtected.PUT("/give-admin", handler.GiveAdminAccess)
	adminProtected.GET("/show-users", handler.ListAllUsers)
	// ERROR
	// adminProtected.PUT("/edit-user", handler.EditUserRole)
}

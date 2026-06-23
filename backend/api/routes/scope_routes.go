package routes

import (
	"omnilogs-api/configs"
	scopehandler "omnilogs-api/internal/scopes/handler"
	scoperepo "omnilogs-api/internal/scopes/repository"
	scopeusecase "omnilogs-api/internal/scopes/usecase"
	"omnilogs-api/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ScopeRoutes(router *gin.Engine, db *gorm.DB, env *configs.Env) {
	handler := scopehandler.NewHandler(scopeusecase.NewUsecase(scoperepo.New(db), db))

	products := router.Group("/api/v1/products")
	products.Use(middleware.UserAuthMiddleware(env.JWTSecret))

	products.POST("/:productId/memberships/:membershipId/scopes", handler.CreateScope)
	products.GET("/:productId/memberships/:membershipId/scopes", handler.ListScopes)
	products.PATCH("/:productId/memberships/:membershipId/scopes/:scopeId", handler.UpdateScope)
	products.DELETE("/:productId/memberships/:membershipId/scopes/:scopeId", handler.DeleteScope)
}

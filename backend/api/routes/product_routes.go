package routes

import (
	"omnilogs-api/configs"
	producthandler "omnilogs-api/internal/product/handler"
	productrepo "omnilogs-api/internal/product/repository"
	productusecase "omnilogs-api/internal/product/usecase"
	"omnilogs-api/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ProductRoutes(router *gin.Engine, db *gorm.DB, env *configs.Env) {
	handler := producthandler.NewHandler(productusecase.NewUsecase(productrepo.NewRepository(db)))

	products := router.Group("/products")
	products.Use(middleware.UserAuthMiddleware(env.JWTSecret))

	products.POST("", handler.CreateProduct)
	products.GET("", handler.ListProducts)
	products.GET("/:productId", handler.GetProduct)
	products.PATCH("/:productId", handler.UpdateProduct)

	products.POST("/:productId/environments", handler.CreateEnvironment)
	products.GET("/:productId/environments", handler.ListEnvironments)
	products.PATCH("/:productId/environments/:environmentId", handler.UpdateEnvironment)

	products.POST("/:productId/api-keys", handler.CreateAPIKey)
	products.GET("/:productId/api-keys", handler.ListAPIKeys)
	products.PATCH("/:productId/api-keys/:keyId", handler.UpdateAPIKey)
	products.DELETE("/:productId/api-keys/:keyId", handler.RevokeAPIKey)

	products.POST("/:productId/projects", handler.CreateProject)
	products.GET("/:productId/projects", handler.ListProjects)
	products.GET("/:productId/projects/:projectId", handler.GetProject)
	products.PATCH("/:productId/projects/:projectId", handler.UpdateProject)
	products.POST("/:productId/projects/:projectId/features", handler.CreateFeature)
	products.GET("/:productId/projects/:projectId/features", handler.ListFeatures)
	products.PATCH("/:productId/projects/:projectId/features/:featureId", handler.UpdateFeature)

	products.POST("/:productId/roles", handler.CreateRole)
	products.GET("/:productId/roles", handler.ListRoles)
	products.PATCH("/:productId/roles/:roleId", handler.UpdateRole)

	products.POST("/:productId/memberships", handler.CreateMembership)
	products.GET("/:productId/memberships", handler.ListMemberships)
	products.PATCH("/:productId/memberships/:membershipId", handler.UpdateMembership)
	products.POST("/:productId/memberships/:membershipId/scopes", handler.CreateScope)
	products.GET("/:productId/memberships/:membershipId/scopes", handler.ListScopes)
	products.PATCH("/:productId/memberships/:membershipId/scopes/:scopeId", handler.UpdateScope)

	products.POST("/:productId/permission-rules", handler.CreatePermissionRule)
	products.GET("/:productId/permission-rules", handler.ListPermissionRules)
	products.PATCH("/:productId/permission-rules/:ruleId", handler.UpdatePermissionRule)

	permissions := router.Group("/permissions")
	permissions.Use(middleware.UserAuthMiddleware(env.JWTSecret))
	permissions.POST("/check", handler.CheckPermission)
}

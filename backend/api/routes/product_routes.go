package routes

import (
	"omnilogs-api/configs"
	apikeyshandler "omnilogs-api/internal/api_keys/handler"
	apikeysrepo "omnilogs-api/internal/api_keys/repository"
	apikeysusecase "omnilogs-api/internal/api_keys/usecase"
	featurehandler "omnilogs-api/internal/feature/handler"
	featurerepo "omnilogs-api/internal/feature/repository"
	featureusecase "omnilogs-api/internal/feature/usecase"
	globalauthhandler "omnilogs-api/internal/global_auth/handler"
	globalauthrepo "omnilogs-api/internal/global_auth/repository"
	globalauthusecase "omnilogs-api/internal/global_auth/usecase"
	producthandler "omnilogs-api/internal/product/handler"
	productrepo "omnilogs-api/internal/product/repository"
	productusecase "omnilogs-api/internal/product/usecase"
	projecthandler "omnilogs-api/internal/project/handler"
	projectrepo "omnilogs-api/internal/project/repository"
	projectusecase "omnilogs-api/internal/project/usecase"
	"omnilogs-api/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ProductRoutes(router *gin.Engine, db *gorm.DB, env *configs.Env) {
	productHandler := producthandler.NewHandler(productusecase.NewUsecase(productrepo.NewRepository(db)))
	apiKeysHandler := apikeyshandler.NewHandler(apikeysusecase.NewUsecase(apikeysrepo.NewRepository(db)))
	projectHandler := projecthandler.NewHandler(projectusecase.NewUsecase(projectrepo.NewRepository(db)))
	featureHandler := featurehandler.NewHandler(featureusecase.NewUsecase(featurerepo.NewRepository(db)))
	accessHandler := globalauthhandler.NewHandler(globalauthusecase.NewUsecase(globalauthrepo.NewRepository(db)))

	products := router.Group("/api/v1/products")
	products.Use(middleware.UserAuthMiddleware(env.JWTSecret))

	products.POST("", productHandler.CreateProduct)
	products.GET("", productHandler.ListProducts)
	products.GET("/:productId", productHandler.GetProduct)
	products.PATCH("/:productId", productHandler.UpdateProduct)
	products.DELETE("/:productId", productHandler.DeleteProduct)

	products.POST("/:productId/api-keys", apiKeysHandler.CreateAPIKey)
	products.GET("/:productId/api-keys", apiKeysHandler.ListAPIKeys)
	products.PATCH("/:productId/api-keys/:keyId", apiKeysHandler.UpdateAPIKey)
	products.DELETE("/:productId/api-keys/:keyId", apiKeysHandler.RevokeAPIKey)

	products.POST("/:productId/projects", projectHandler.CreateProject)
	products.GET("/:productId/projects", projectHandler.ListProjects)
	products.GET("/:productId/projects/:projectId", projectHandler.GetProject)
	products.PATCH("/:productId/projects/:projectId", projectHandler.UpdateProject)
	products.POST("/:productId/projects/:projectId/features", featureHandler.CreateFeature)
	products.GET("/:productId/projects/:projectId/features", featureHandler.ListFeatures)
	products.PATCH("/:productId/projects/:projectId/features/:featureId", featureHandler.UpdateFeature)
	products.DELETE("/:productId/projects/:projectId/features/:featureId", featureHandler.DeleteFeature)

	products.POST("/:productId/roles", accessHandler.CreateRole)
	products.GET("/:productId/roles", accessHandler.ListRoles)
	products.PATCH("/:productId/roles/:roleId", accessHandler.UpdateRole)
	products.DELETE("/:productId/roles/:roleId", accessHandler.DeleteRole)

	products.POST("/:productId/memberships", accessHandler.CreateMembership)
	products.GET("/:productId/memberships", accessHandler.ListMemberships)
	products.PATCH("/:productId/memberships/:membershipId", accessHandler.UpdateMembership)
	products.DELETE("/:productId/memberships/:membershipId", accessHandler.DeleteMembership)

	products.POST("/:productId/permission-rules", accessHandler.CreatePermissionRule)
	products.GET("/:productId/permission-rules", accessHandler.ListPermissionRules)
	products.PATCH("/:productId/permission-rules/:ruleId", accessHandler.UpdatePermissionRule)

	permissions := router.Group("/api/v1/authorization")
	permissions.Use(middleware.UserAuthMiddleware(env.JWTSecret))
	permissions.POST("/check", accessHandler.CheckPermission)
}

package routes

import (
	"omnilogs-api/configs"
	apikeyshandler "omnilogs-api/internal/api_keys/handler"
	apikeysrepo "omnilogs-api/internal/api_keys/repository"
	apikeysusecase "omnilogs-api/internal/api_keys/usecase"
	elasticIndexPolicyHandler "omnilogs-api/internal/elastic_index_policy/handler"
	elasticIndexPolicyRepo "omnilogs-api/internal/elastic_index_policy/repository"
	elasticIndexPolicyUsecase "omnilogs-api/internal/elastic_index_policy/usecase"
	featurehandler "omnilogs-api/internal/feature/handler"
	featurerepo "omnilogs-api/internal/feature/repository"
	featureusecase "omnilogs-api/internal/feature/usecase"
	globalauthhandler "omnilogs-api/internal/global_auth/handler"
	globalauthrepo "omnilogs-api/internal/global_auth/repository"
	globalauthusecase "omnilogs-api/internal/global_auth/usecase"
	logArchiveHandler "omnilogs-api/internal/log_archive/handler"
	producthandler "omnilogs-api/internal/product/handler"
	productrepo "omnilogs-api/internal/product/repository"
	productusecase "omnilogs-api/internal/product/usecase"
	projecthandler "omnilogs-api/internal/project/handler"
	projectrepo "omnilogs-api/internal/project/repository"
	projectusecase "omnilogs-api/internal/project/usecase"
	sensitiveLogHandler "omnilogs-api/internal/sensitive_log/handler"
	sensitiveLogRepo "omnilogs-api/internal/sensitive_log/repository"
	sensitiveLogUsecase "omnilogs-api/internal/sensitive_log/usecase"
	"omnilogs-api/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ProductRoutes(router *gin.Engine, db *gorm.DB, env *configs.Env) {
	esClient, err := configs.ConnectElasticsearch(env)
	if err != nil {
		panic(err)
	}

	productHandler := producthandler.NewHandler(productusecase.NewUsecase(productrepo.NewRepository(db)))
	apiKeysHandler := apikeyshandler.NewHandler(apikeysusecase.NewUsecase(apikeysrepo.NewRepository(db)))
	projectHandler := projecthandler.NewHandler(projectusecase.NewUsecase(projectrepo.NewRepository(db)))
	featureHandler := featurehandler.NewHandler(featureusecase.NewUsecase(featurerepo.NewRepository(db)))
	accessHandler := globalauthhandler.NewHandler(globalauthusecase.NewUsecase(globalauthrepo.NewRepository(db)))
	sensitiveHandler := sensitiveLogHandler.NewHandler(
		sensitiveLogUsecase.NewUsecase(
			sensitiveLogRepo.NewRepository(db),
			env.DataEncryptionKey,
		),
	)

	products := router.Group("/api/v1/products")
	products.Use(middleware.UserAuthMiddleware(env.JWTSecret))

	products.POST("", productHandler.CreateProduct)
	products.GET("", productHandler.ListProducts)
	products.GET("/:productId", productHandler.GetProduct)
	products.PATCH("/:productId", productHandler.UpdateProduct)
	products.DELETE("/:productId", productHandler.DeleteProduct)
	products.DELETE("/product/bulk-delete", productHandler.BulkDeleteProducts)

	archiveHandler := logArchiveHandler.NewHandler(db, esClient)
	products.GET("/:productId/log-archives", archiveHandler.List)
	products.POST("/:productId/log-archives/:archiveId/restore", archiveHandler.Restore)

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
	products.POST("/:productId/memberships/bulk", accessHandler.CreateMemberships)
	products.GET("/:productId/memberships", accessHandler.ListMemberships)
	products.PATCH("/:productId/memberships/:membershipId", accessHandler.UpdateMembership)
	products.DELETE("/:productId/memberships/:membershipId", accessHandler.DeleteMembership)
	products.GET("/:productId/access-overview", accessHandler.GetProductAccessOverview)
	products.PUT("/:productId/access/members/:userId", accessHandler.UpsertProductAccess)

	products.POST("/:productId/permission-rules", accessHandler.CreatePermissionRule)
	products.GET("/:productId/permission-rules", accessHandler.ListPermissionRules)
	products.PATCH("/:productId/permission-rules/:ruleId", accessHandler.UpdatePermissionRule)

	products.POST("/:productId/sensitive-logs/requests", sensitiveHandler.CreateAccessRequest)
	products.POST("/:productId/sensitive-logs/requests/:requestId/review", sensitiveHandler.ReviewAccessRequest)
	products.GET("/:productId/sensitive-logs/requests", sensitiveHandler.ListAccessRequests)
	products.GET("/:productId/sensitive-logs/secrets", sensitiveHandler.ListSecrets)
	products.POST("/:productId/sensitive-logs/reveal", sensitiveHandler.RevealSensitiveValue)
	products.GET("/:productId/sensitive-logs/main-logs/:logId/raw", sensitiveHandler.RevealRawMainLog)
	products.GET("/:productId/sensitive-logs/history", sensitiveHandler.ListAccessHistory)

	permissions := router.Group("/api/v1/authorization")
	permissions.Use(middleware.UserAuthMiddleware(env.JWTSecret))
	permissions.POST("/check", accessHandler.CheckPermission)

	elasticIndexPolicyHandler := elasticIndexPolicyHandler.NewHandler(elasticIndexPolicyUsecase.NewUsecase(elasticIndexPolicyRepo.NewRepository(db, esClient)))
	elasticIndexPolicies := products.Group("/:productId/elastic-index-policies")
	elasticIndexPolicies.Use(middleware.UserAuthMiddleware(env.JWTSecret))
	elasticIndexPolicies.POST("", elasticIndexPolicyHandler.Create)
	elasticIndexPolicies.GET("", elasticIndexPolicyHandler.List)
	elasticIndexPolicies.GET("/:elasticPolicyId", elasticIndexPolicyHandler.GetByID)
	elasticIndexPolicies.PATCH("/:elasticPolicyId", elasticIndexPolicyHandler.Update)
	elasticIndexPolicies.DELETE("/:elasticPolicyId", elasticIndexPolicyHandler.Delete)
	// push to archives
	elasticIndexPolicies.POST("/:elasticPolicyId/push-to-archives", elasticIndexPolicyHandler.PushToArchives)
	elasticIndexPolicies.DELETE("/clear-logs", elasticIndexPolicyHandler.ClearAllLogs)
}

package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"omnilogs-api/configs"
	workerprocessor "omnilogs-api/internal/worker"
	"omnilogs-api/middleware"
	"omnilogs-api/middleware/audit"
	"omnilogs-api/routes"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RunAPIServer(env *configs.Env, db *gorm.DB) error {
	esClient, err := configs.ConnectElasticsearch(env)
	if err != nil {
		return err
	}
	natsQueue, err := configs.ConnectNATS(env)
	if err != nil {
		return err
	}
	defer natsQueue.Close()

	processor := workerprocessor.NewProcessor(db, esClient, env.DataEncryptionKey, natsQueue)
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	go func() {
		slog.Info("background worker engine started for API server")
		_ = processor.Run(workerCtx)
	}()

	router := NewRouter(env, db, esClient, natsQueue)
	server := &http.Server{
		Addr:              ":" + env.AppPort,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("omnilogs api listening", "addr", server.Addr)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return err
	}

	slog.Info("server shutdown completed")
	return nil
}

func NewRouter(env *configs.Env, db *gorm.DB, esClient *elasticsearch.Client, natsQueue *configs.NATSQueue) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(middleware.Recovery())

	config := cors.DefaultConfig()
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "X-Request-ID", "X-API-Key", "X-Product-Code", "X-Environment-Code"}
	config.AllowCredentials = true
	config.AllowOriginFunc = func(origin string) bool {
		return isAllowedOrigin(origin, env.CORSAllowedOrigins)
	}
	router.Use(cors.New(config))
	router.Use(middleware.RequestID())
	router.Use(middleware.PerformanceLogger(time.Duration(env.APISlowRequestThresholdMS) * time.Millisecond))
	router.Use(middleware.GlobalRateLimit(env))
	router.Use(audit.Logger(db, env.DataEncryptionKey))

	routes.HealthRoutes(router, func(ctx context.Context) error {
		if err := pingDatabase(ctx, db); err != nil {
			return err
		}
		return pingElasticsearch(ctx, esClient)
	})
	routes.SwaggerRoutes(router, env)
	routes.RegisterAuthRoutes(router, db, env)
	routes.DashboardRoutes(router, db, esClient, env)
	routes.MainLogRoutes(router, db, esClient, env, natsQueue)
	routes.AuditRoutes(router, db, esClient, env)
	routes.LogQueueRoutes(router, db, env, esClient, natsQueue)

	return router
}

func isAllowedOrigin(origin string, allowedOriginsConfig string) bool {
	if origin == "" || origin == "null" {
		return true
	}

	if allowedOriginsConfig != "" {
		for _, allowed := range strings.Split(allowedOriginsConfig, ",") {
			allowed = strings.TrimSpace(allowed)
			if allowed == "*" || allowed == origin || strings.TrimSuffix(allowed, "/") == strings.TrimSuffix(origin, "/") {
				return true
			}
		}
	}

	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	hostname := u.Hostname()

	if hostname == "" || hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1" {
		return true
	}

	ip := net.ParseIP(hostname)
	if ip != nil {
		return ip.IsPrivate() || ip.IsLoopback()
	}

	return false
}

func pingDatabase(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func pingElasticsearch(ctx context.Context, esClient *elasticsearch.Client) error {
	res, err := esClient.Info(esClient.Info.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch returned status %d", res.StatusCode)
	}

	return nil
}

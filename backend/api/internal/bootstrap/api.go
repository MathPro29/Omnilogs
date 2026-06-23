package bootstrap

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"omnilogs-api/configs"
	"omnilogs-api/middleware"
	"omnilogs-api/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RunAPIServer(env *configs.Env, db *gorm.DB) error {
	router := NewRouter(env, db)
	server := &http.Server{
		Addr:              ":" + env.AppPort,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server stopped unexpectedly: %v", err)
		}
	}()

	log.Printf("omnilogs api listening on %s", server.Addr)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return err
	}

	log.Println("server shutdown completed")
	return nil
}

func NewRouter(env *configs.Env, db *gorm.DB) *gin.Engine {
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"}
	config.AllowCredentials = true
	config.AllowOriginFunc = func(origin string) bool {
		return isAllowedOrigin(origin)
	}
	router.Use(cors.New(config))
	router.Use(middleware.RequestID())
	router.Use(middleware.GlobalRateLimit(env))
	router.Use(middleware.AuditLogger(db))

	routes.HealthRoutes(router, func(ctx context.Context) error {
		if err := pingDatabase(ctx, db); err != nil {
			return err
		}
		return pingElasticsearch(ctx, env.ElasticURL)
	})
	routes.SwaggerRoutes(router, env)
	routes.RegisterAuthRoutes(router, db, env)

	return router
}

func isAllowedOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	hostname := u.Hostname()

	if hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1" {
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

func pingElasticsearch(ctx context.Context, elasticURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, elasticURL, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("elasticsearch returned status %d", resp.StatusCode)
	}

	return nil
}

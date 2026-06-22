package main

import (
	"log"
	"net"
	"net/url"
	"omnilogs-api/configs"
	"omnilogs-api/middleware"
	"omnilogs-api/migrate"
	"omnilogs-api/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	env := configs.LoadEnv()

	db, err := configs.ConnectDB(env)
	if err != nil {
		log.Fatalf("connect database failed: %v", err)
	}
	// call migrate function
	migrate.Migrate(db, env)

	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	config.AllowCredentials = true
	config.AllowOriginFunc = func(origin string) bool {
		return isAllowedOrigin(origin)
	}
	router.Use(cors.New(config))

	router.Use(middleware.RequestID())
	router.Use(middleware.GlobalRateLimit(env))

	routes.RegisterAuthRoutes(router, db, env)

	if err := router.Run(":" + env.AppPort); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func isAllowedOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	hostname := u.Hostname()

	// Allow localhost, loopback, and local network IPs
	if hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1" {
		return true
	}

	ip := net.ParseIP(hostname)
	if ip != nil {
		return ip.IsPrivate() || ip.IsLoopback()
	}

	return false
}

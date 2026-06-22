package main

import (
	"log"

	"omnilogs-api/configs"
	"omnilogs-api/middleware"

	// "omnilogs-api/models"
	"omnilogs-api/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	env := configs.LoadEnv()

	db, err := configs.ConnectDB(env)
	if err != nil {
		log.Fatalf("connect database failed: %v", err)
	}
	// migrate db
	// if err := db.AutoMigrate(&models.User{}, &models.Todo{}); err != nil {
	// 	log.Fatalf("auto migrate failed: %v", err)
	// }

	router := gin.Default()
	router.Use(middleware.RequestID())
	router.Use(middleware.GlobalRateLimit(env))

	routes.RegisterAuthRoutes(router, db, env)
	// Category
	

	if err := router.Run(":" + env.AppPort); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

//go:build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"omnilogs-api/configs"
	"omnilogs-api/dto"
	workerrepo "omnilogs-api/internal/worker/repository"
	workerusecase "omnilogs-api/internal/worker/usecase"
	"omnilogs-api/models"
)

func main() {
	os.Setenv("ENV_PATH", "../.env")
	env := configs.LoadEnv()
	db, err := configs.ConnectDB(env)
	if err != nil {
		log.Fatalf("connect DB failed: %v", err)
	}

	ctx := context.Background()
	productID := 17
	environmentID := 29

	payload := map[string]any{
		"level":         "INFO",
		"message":       "Direct worker routing test",
		"route_key":     "testproduct.event",
		"facility_name": "สนามเทนนิส",
	}

	// Process routing
	fmt.Println("Before:", payload)
	// Query log_routes directly using repo DB to see if DB connection returns the row
	var route models.LogRoute
	err = db.WithContext(ctx).Where("product_id = ? AND environment_id = ? AND is_active = TRUE AND LOWER(BTRIM(route_key)) = LOWER(BTRIM(?))", productID, environmentID, "testproduct.event").First(&route).Error
	if err != nil {
		fmt.Printf("DB query error: %v\n", err)
	} else {
		fmt.Printf("DB route found: ID=%d, ProjectID=%d, CatID=%v\n", route.RouteID, route.ProjectID, route.CategoryID)
	}
}

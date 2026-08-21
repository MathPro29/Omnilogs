//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"omnilogs-api/configs"
	"omnilogs-api/models"
)

func main() {
	os.Setenv("ENV_PATH", "../.env")
	env := configs.LoadEnv()
	db, err := configs.ConnectDB(env)
	if err != nil {
		log.Fatalf("connect DB failed: %v", err)
	}

	routeKey := "testproduct.event"
	productID := 17
	environmentID := 29

	var route models.LogRoute
	err = db.Where("product_id = ? AND environment_id = ? AND is_active = TRUE AND LOWER(BTRIM(route_key)) = LOWER(BTRIM(?))", productID, environmentID, routeKey).
		Order("priority DESC, route_id").First(&route).Error

	if err != nil {
		log.Fatalf("no route matched: %v", err)
	}

	fmt.Printf("✅ ROUTE MATCHED SUCCESSFULLY!\n")
	fmt.Printf("RouteID: %d\n", route.RouteID)
	fmt.Printf("ProductID: %d\n", route.ProductID)
	fmt.Printf("EnvironmentID: %d\n", route.EnvironmentID)
	fmt.Printf("RouteKey: %s\n", route.RouteKey)
	fmt.Printf("ProjectID: %d\n", route.ProjectID)
	if route.CategoryID != nil {
		fmt.Printf("CategoryID: %d\n", *route.CategoryID)
	} else {
		fmt.Printf("CategoryID: nil\n")
	}
}

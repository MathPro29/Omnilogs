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

	var routes []models.LogRoute
	db.Where("product_id = ? AND environment_id = ?", 17, 29).Find(&routes)
	fmt.Printf("Routes count: %d\n", len(routes))
	for _, r := range routes {
		fmt.Printf("RouteID: %d, ProductID: %d, EnvID: %d, RouteKey: '%s', ProjectID: %d, CatID: %v, IsActive: %v\n",
			r.RouteID, r.ProductID, r.EnvironmentID, r.RouteKey, r.ProjectID, r.CategoryID, r.IsActive)
	}
}

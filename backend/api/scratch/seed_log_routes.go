//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"omnilogs-api/configs"
	"omnilogs-api/models"

	"gorm.io/gorm"
)

func main() {
	os.Setenv("ENV_PATH", "../.env")
	env := configs.LoadEnv()
	db, err := configs.ConnectDB(env)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}

	if err := SeedLogRoutes(db); err != nil {
		log.Fatalf("failed to seed log routes: %v", err)
	}
	fmt.Println("All Log routes seeded successfully!")
}

func SeedLogRoutes(db *gorm.DB) error {
	var products []models.Product
	if err := db.Find(&products).Error; err != nil {
		return err
	}

	for _, p := range products {
		var envs []models.ProductEnvironment
		db.Where("product_id = ?", p.ProductID).Find(&envs)
		if len(envs) == 0 {
			continue
		}

		var projects []models.Project
		db.Where("product_id = ?", p.ProductID).Find(&projects)
		if len(projects) == 0 {
			continue
		}

		firstProject := projects[0]
		var categories []models.ProjectFeature
		db.Where("product_id = ? AND project_id = ?", p.ProductID, firstProject.ProjectID).Find(&categories)
		var firstCategoryID *int
		if len(categories) > 0 {
			firstCategoryID = &categories[0].CategoryID
		}

		for _, env := range envs {
			routeKeys := []string{
				"testproduct.event",
				"testproduct.event.created",
				fmt.Sprintf("%s.event", p.ProductCode),
				fmt.Sprintf("%s.%s", p.ProductCode, firstProject.ProjectCode),
			}

			for idx, key := range routeKeys {
				var existing models.LogRoute
				err := db.Where("product_id = ? AND environment_id = ? AND LOWER(BTRIM(route_key)) = LOWER(BTRIM(?))", p.ProductID, env.EnvironmentID, key).First(&existing).Error
				if err != nil {
					newRoute := models.LogRoute{
						ProductID:     p.ProductID,
						EnvironmentID: env.EnvironmentID,
						RouteKey:      key,
						ProjectID:     firstProject.ProjectID,
						CategoryID:    firstCategoryID,
						Priority:      10 - idx,
						IsActive:      true,
					}
					if err := db.Create(&newRoute).Error; err != nil {
						return fmt.Errorf("failed to create log route for product %d, route_key %s: %w", p.ProductID, key, err)
					}
					fmt.Printf("[SEED ROUTE] Product=%s(%d) | Env=%s(%d) | RouteKey=%s -> ProjectID=%d, CatID=%v\n",
						p.ProductName, p.ProductID, env.EnvironmentName, env.EnvironmentID, key, firstProject.ProjectID, firstCategoryID)
				}
			}
		}
	}
	return nil
}

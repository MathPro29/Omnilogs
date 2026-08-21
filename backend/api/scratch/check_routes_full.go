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
		log.Fatalf("failed to connect db: %v", err)
	}

	var products []models.Product
	db.Find(&products)
	fmt.Println("=== Products ===")
	for _, p := range products {
		fmt.Printf("ProductID: %d, Name: %s, Code: %s\n", p.ProductID, p.ProductName, p.ProductCode)
	}

	var envs []models.ProductEnvironment
	db.Find(&envs)
	fmt.Println("=== Environments ===")
	for _, e := range envs {
		fmt.Printf("EnvID: %d, ProductID: %d, Name: %s, Code: %s\n", e.EnvironmentID, e.ProductID, e.EnvironmentName, e.EnvironmentCode)
	}

	var projects []models.Project
	db.Find(&projects)
	fmt.Println("=== Projects ===")
	for _, pr := range projects {
		fmt.Printf("ProjectID: %d, ProductID: %d, Name: %s, Code: %s\n", pr.ProjectID, pr.ProductID, pr.ProjectName, pr.ProjectCode)
	}

	var features []models.ProjectFeature
	db.Find(&features)
	fmt.Println("=== Features/Categories ===")
	for _, f := range features {
		fmt.Printf("CategoryID: %d, ProductID: %d, ProjectID: %d, Name: %s, Code: %s, Path: %v\n", f.CategoryID, f.ProductID, f.ProjectID, f.CategoryName, f.CategoryCode, f.FullPath)
	}

	var apiKeys []models.ProductAPIKey
	db.Find(&apiKeys)
	fmt.Println("=== API Keys ===")
	for _, k := range apiKeys {
		fmt.Printf("KeyID: %d, ProductID: %d, EnvID: %v, Name: %s, KeyPrefix: %s, DefProject: %v, DefCat: %v\n", k.KeyID, k.ProductID, k.EnvironmentID, k.KeyName, k.KeyPrefix, k.DefaultProjectID, k.DefaultCategoryID)
	}

	var routes []models.LogRoute
	db.Find(&routes)
	fmt.Println("=== Log Routes ===")
	for _, r := range routes {
		fmt.Printf("RouteID: %d, ProductID: %d, EnvID: %d, RouteKey: %s, ProjectID: %d, CategoryID: %v, Priority: %d\n", r.RouteID, r.ProductID, r.EnvironmentID, r.RouteKey, r.ProjectID, r.CategoryID, r.Priority)
	}
}

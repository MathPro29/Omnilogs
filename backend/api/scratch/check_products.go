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
		fmt.Printf("EnvID: %d, ProductID: %d, Name: %s\n", e.EnvironmentID, e.ProductID, e.EnvironmentName)
	}
}

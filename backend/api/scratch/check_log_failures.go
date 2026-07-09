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

	fmt.Println("=== Elastic Index Policies ===")
	var policies []models.ElasticIndexPolicy
	db.Find(&policies)
	for _, p := range policies {
		prefixStr := "nil"
		if p.IndexPrefix != "" {
			prefixStr = p.IndexPrefix
		}
		fmt.Printf("PolicyID: %s, ProductID: %d, EnvID: %v, Prefix: %s, Active: %v\n",
			p.ElasticPolicyID, p.ProductID, p.EnvironmentID, prefixStr, p.IsActive)
	}

	fmt.Println("\n=== Product Environments ===")
	var envs []models.ProductEnvironment
	db.Find(&envs)
	for _, e := range envs {
		fmt.Printf("EnvID: %d, ProductID: %d, EnvName: %s, EnvCode: %s\n",
			e.EnvironmentID, e.ProductID, e.EnvironmentName, e.EnvironmentCode)
	}
}

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

	var secrets []models.LogSensitiveFieldSecret
	db.Find(&secrets)
	fmt.Println("=== Sensitive Secrets ===")
	for _, s := range secrets {
		fmt.Printf("SecretID: %s, LogID: %s, ProductID: %d, Key: %s, Path: %s, SourceSection: %s\n",
			s.SecretID, s.LogID, s.ProductID, s.FieldKey, s.FieldPath, s.SourceSection)
	}
}

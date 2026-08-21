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

	var fields []models.LogFieldDefinition
	db.Find(&fields)
	fmt.Println("=== Custom Fields ===")
	for _, f := range fields {
		fmt.Printf("FieldID: %d, Key: %s, DisplayName: %s, DataType: %s, IsActive: %v, IsFilterable: %v, FieldPath: %s\n",
			f.FieldDefinitionID, f.FieldKey, *f.DisplayName, f.DataType, f.IsActive, f.IsFilterable, *f.FieldPath)
	}
}

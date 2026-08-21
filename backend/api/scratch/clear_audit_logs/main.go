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
	// Set dotenv path to root of backend api
	os.Setenv("ENV_PATH", "../../.env")

	env := configs.LoadEnv()
	if err := env.Validate(); err != nil {
		log.Fatalf("invalid environment configuration: %v", err)
	}

	// Connect to Postgres
	db, err := configs.ConnectDB(env)
	if err != nil {
		log.Fatalf("connect database failed: %v", err)
	}

	fmt.Println("--- Clearing PostgreSQL system_audit_logs Table ---")

	// Delete all rows from system_audit_logs
	result := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.SystemAuditLog{})
	if result.Error != nil {
		log.Fatalf("Failed to clear system_audit_logs: %v\n", result.Error)
	}

	fmt.Printf("Successfully cleared system_audit_logs (deleted %d rows).\n", result.RowsAffected)
}

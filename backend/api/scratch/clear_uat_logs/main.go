package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"omnilogs-api/configs"
	"omnilogs-api/models"

	"github.com/elastic/go-elasticsearch/v8"
	"gorm.io/gorm"
)

func main() {
	os.Setenv("ENV_PATH", "../../.env")

	env := configs.LoadEnv()
	if err := env.Validate(); err != nil {
		log.Fatalf("invalid environment configuration: %v", err)
	}

	// 1. Clear Postgres system_audit_logs
	fmt.Println("--- 1. Clearing PostgreSQL system_audit_logs Table ---")
	db, err := configs.ConnectDB(env)
	if err != nil {
		log.Fatalf("connect database failed: %v", err)
	}

	result := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.SystemAuditLog{})
	if result.Error != nil {
		log.Fatalf("Failed to clear system_audit_logs: %v\n", result.Error)
	}
	fmt.Printf("Successfully cleared system_audit_logs (deleted %d rows).\n", result.RowsAffected)

	// 2. Clear Log Explorer documents in Elasticsearch / OpenSearch
	fmt.Println("\n--- 2. Clearing Elasticsearch/OpenSearch Log Indices ---")
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{env.ElasticURL},
	})
	if err != nil {
		log.Fatalf("failed to create elasticsearch client: %v", err)
	}

	// Target index explicitly or use _all/_match_all without wildcard index pattern that fails
	res, err := es.DeleteByQuery(
		[]string{"omnilogs-log*"},
		nil,
		es.DeleteByQuery.WithQuery(`{"query":{"match_all":{}}}`),
		es.DeleteByQuery.WithContext(context.Background()),
		es.DeleteByQuery.WithConflicts("proceed"),
	)
	if err != nil {
		log.Printf("Error during DeleteByQuery: %v\n", err)
	} else {
		defer res.Body.Close()
		fmt.Printf("Elasticsearch DeleteByQuery Response Status: %s\n", res.Status())
	}

	fmt.Println("\n=== Clean-up Completed Successfully ===")
}

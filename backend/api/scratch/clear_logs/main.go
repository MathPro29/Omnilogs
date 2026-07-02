package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

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

	// 1. Connect to Postgres
	db, err := configs.ConnectDB(env)
	if err != nil {
		log.Fatalf("connect database failed: %v", err)
	}

	// 2. Connect to Elasticsearch
	esClient, err := configs.ConnectElasticsearch(env)
	if err != nil {
		log.Fatalf("connect elasticsearch failed: %v", err)
	}

	ctx := context.Background()

	// 3. Clear Postgres Logs References
	fmt.Println("--- Clearing Postgres Log Tables ---")
	
	clearTable := func(model interface{}, tableName string) {
		result := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(model)
		if result.Error != nil {
			fmt.Printf("❌ Failed to clear table %s: %v\n", tableName, result.Error)
		} else {
			fmt.Printf("✅ Cleared table %s (deleted %d rows)\n", tableName, result.RowsAffected)
		}
	}

	clearTable(&models.LogIndexRef{}, "log_index_refs")
	clearTable(&models.LogFailure{}, "log_failures")
	clearTable(&models.LogQueueItem{}, "log_queue_items")
	clearTable(&models.LogQueueBatch{}, "log_queue_batches")
	clearTable(&models.LogSensitiveFieldSecret{}, "log_sensitive_field_secrets")

	// 4. Clear Elasticsearch indices
	fmt.Println("\n--- Clearing Elasticsearch Indices ---")

	// Query dynamic index prefixes from policies to make sure we clear custom indices as well
	var policies []models.ElasticIndexPolicy
	if err := db.Find(&policies).Error; err != nil {
		fmt.Printf("⚠️ Failed to load index policies: %v\n", err)
	}

	targets := []string{"omnilogs-*"}
	for _, policy := range policies {
		if strings.TrimSpace(policy.IndexPrefix) != "" {
			pattern := fmt.Sprintf("%s-*", strings.ToLower(strings.TrimSpace(policy.IndexPrefix)))
			// Avoid duplicates
			found := false
			for _, t := range targets {
				if t == pattern {
					found = true
					break
				}
			}
			if !found {
				targets = append(targets, pattern)
			}
		}
	}

	fmt.Printf("Deleting Elasticsearch indices matching: %s\n", strings.Join(targets, ", "))

	// Enable wildcard deletes temporarily
	disableWildcardCheckUrl := fmt.Sprintf("%s/_cluster/settings", env.ElasticURL)
	disableReq, err := http.NewRequest("PUT", disableWildcardCheckUrl, strings.NewReader(`{"transient":{"action.destructive_requires_name":false}}`))
	if err == nil {
		disableReq.Header.Set("Content-Type", "application/json")
		if disableResp, err := http.DefaultClient.Do(disableReq); err == nil {
			disableResp.Body.Close()
		}
	}

	res, err := esClient.Indices.Delete(
		targets,
		esClient.Indices.Delete.WithContext(ctx),
	)

	// Re-enable wildcard protection (good practice)
	enableReq, err := http.NewRequest("PUT", disableWildcardCheckUrl, strings.NewReader(`{"transient":{"action.destructive_requires_name":true}}`))
	if err == nil {
		enableReq.Header.Set("Content-Type", "application/json")
		if enableResp, err := http.DefaultClient.Do(enableReq); err == nil {
			enableResp.Body.Close()
		}
	}

	if err != nil {
		fmt.Printf("❌ Failed to delete indices: %v\n", err)
	} else {
		defer res.Body.Close()
		if res.IsError() {
			if res.StatusCode == 404 {
				fmt.Println("✅ No matching indices found in Elasticsearch to delete.")
			} else {
				fmt.Printf("❌ Elasticsearch delete returned error (Status %d): %s\n", res.StatusCode, res.String())
			}
		} else {
			fmt.Println("✅ Successfully deleted indices from Elasticsearch.")
		}
	}

	fmt.Println("\n🎉 Clear Logs Completed!")
}

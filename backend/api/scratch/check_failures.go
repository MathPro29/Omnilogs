package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"omnilogs-api/configs"
	"omnilogs-api/models"
)

func main() {
	os.Setenv("ENV_PATH", ".env")
	env := configs.LoadEnv()
	db, err := configs.ConnectDB(env)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}

	var totalBatches int64
	db.Model(&models.LogQueueBatch{}).Count(&totalBatches)

	var sumLogs struct {
		Total int64
	}
	db.Model(&models.LogQueueBatch{}).Select("COALESCE(SUM(total_logs), 0) as total").Scan(&sumLogs)

	var statusCounts []struct {
		Status string
		Count  int64
		Logs   int64
	}
	db.Model(&models.LogQueueBatch{}).Select("status, COUNT(*) as count, COALESCE(SUM(total_logs), 0) as logs").Group("status").Scan(&statusCounts)

	fmt.Printf("=== Total Batches: %d | Total Logs Received across all batches: %d ===\n", totalBatches, sumLogs.Total)
	for _, sc := range statusCounts {
		fmt.Printf("Status: %-10s | Batches: %d | Total Logs: %d\n", sc.Status, sc.Count, sc.Logs)
	}

	var indexRefs int64
	db.Model(&models.LogIndexRef{}).Count(&indexRefs)
	fmt.Printf("=== Total Indexed in Postgres (log_index_refs): %d ===\n", indexRefs)

	esClient, err := configs.ConnectElasticsearch(env)
	if err == nil {
		res, err := esClient.Search(
			esClient.Search.WithIndex("omnilogs-*", "test-*"),
			esClient.Search.WithSize(0),
			esClient.Search.WithTrackTotalHits(true),
		)
		if err == nil {
			defer res.Body.Close()
			var r map[string]any
			if json.NewDecoder(res.Body).Decode(&r) == nil {
				if hits, ok := r["hits"].(map[string]any); ok {
					if total, ok := hits["total"].(map[string]any); ok {
						fmt.Printf("=== Elasticsearch total hits: %v ===\n", total["value"])
					}
				}
			}
		}
	}
}

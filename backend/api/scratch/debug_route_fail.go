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
		log.Fatalf("connect DB failed: %v", err)
	}

	// Document 52259b7b-9cc0-4f8b-a5a5-d9e137aa06a2 had:
	// ProductID: 17, EnvID: 29, route_key: "testproduct.event"
	routeKey := "testproduct.event"

	// Let's check API Key used by batch ed9d0636-9ed3-4777-befd-d0796bd76179
	var batch models.LogQueueBatch
	if err := db.Where("batch_id = ?", "ed9d0636-9ed3-4777-befd-d0796bd76179").First(&batch).Error; err == nil {
		fmt.Printf("Batch found! ProductID: %v, EnvID: %v, SourceID: %v\n",
			batch.ProductID, batch.EnvironmentID, batch.SourceID)

		var route models.LogRoute
		query := db.Where("product_id = ? AND environment_id = ? AND is_active = TRUE AND LOWER(BTRIM(route_key)) = LOWER(BTRIM(?))", *batch.ProductID, *batch.EnvironmentID, routeKey).
			Where("source_id IS NULL")
		if batch.SourceID != nil {
			query = db.Where("product_id = ? AND environment_id = ? AND is_active = TRUE AND LOWER(BTRIM(route_key)) = LOWER(BTRIM(?))", *batch.ProductID, *batch.EnvironmentID, routeKey).
				Where("source_id IS NULL OR source_id = ?", *batch.SourceID).
				Order("CASE WHEN source_id IS NULL THEN 1 ELSE 0 END")
		}
		err := query.Order("priority DESC, route_id").First(&route).Error
		fmt.Printf("Query result: err=%v, route=%+v\n", err, route)
	} else {
		fmt.Printf("Batch query err: %v\n", err)
	}
}

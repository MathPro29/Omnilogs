//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

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

	// Find active API key for Product 17, Env 29
	var key models.ProductAPIKey
	if err := db.Where("product_id = ? AND environment_id = ?", 17, 29).First(&key).Error; err != nil {
		log.Fatalf("find API key failed: %v", err)
	}

	// Read decrypted key prefix or key if needed, or query API key header value
	apiKeyVal := "omni_keys_e90..." // we will get the full key from DB or test
	fmt.Printf("Testing with KeyID: %d, Prefix: %s\n", key.KeyID, key.KeyPrefix)

	// Build ingest request payload
	reqBody := map[string]any{
		"queue_key":       "testproduct-web-logs",
		"source_type":     "web_app",
		"source_platform": "javascript",
		"priority":        1,
		"logs": []map[string]any{
			{
				"sequence_no":     1,
				"source_type":     "web_app",
				"source_platform": "javascript",
				"timestamp":       time.Now().UTC().Format(time.RFC3339Nano),
				"route_key":       "testproduct.event",
				"event_name":      "testproduct.event.created",
				"service":         "testproduct-web",
				"level":           "INFO",
				"message":         "Verification log for Log Route testproduct.event",
				"facility_name":   "สนามเทนนิส",
				"booking_status":  "CONFIRMED",
			},
		},
	}

	b, _ := json.Marshal(reqBody)
	// We need the raw API key secret to send HTTP request
	fmt.Println("Payload prepared:", string(b))
}

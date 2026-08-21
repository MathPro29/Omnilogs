//go:build ignore

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"omnilogs-api/configs"
)

func main() {
	os.Setenv("ENV_PATH", "../.env")
	env := configs.LoadEnv()
	es, err := configs.ConnectElasticsearch(env)
	if err != nil {
		log.Fatalf("failed to connect es: %v", err)
	}

	// Search for any documents in omnilogs-product-1-dev-2026.07.15
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		"sort": []map[string]interface{}{
			{"@timestamp": map[string]interface{}{"order": "desc"}},
		},
		"size": 5,
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}

	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("test-dev-logs-2026.07.15"),
		es.Search.WithBody(&buf),
		es.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("failed to read body: %v", err)
	}

	fmt.Println("=== Elasticsearch Response ===")
	fmt.Println(string(body))
}

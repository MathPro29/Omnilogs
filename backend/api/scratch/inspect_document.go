//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"omnilogs-api/configs"

	"github.com/elastic/go-elasticsearch/v8"
)

func main() {
	os.Setenv("ENV_PATH", "../.env")
	env := configs.LoadEnv()

	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{env.ElasticURL},
	})
	if err != nil {
		log.Fatalf("connect ES failed: %v", err)
	}

	docID := "52259b7b-9cc0-4f8b-a5a5-d9e137aa06a2"
	res, err := esClient.Get("omnilogs-product-17-env-29-2026.07.27", docID)
	if err != nil {
		log.Fatalf("get ES doc failed: %v", err)
	}
	defer res.Body.Close()

	var payload struct {
		Found  bool           `json:"found"`
		Source map[string]any `json:"_source"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		log.Fatalf("decode ES doc failed: %v", err)
	}

	fmt.Printf("Found: %v\n", payload.Found)
	b, _ := json.MarshalIndent(payload.Source, "", "  ")
	fmt.Println("Source:\n", string(b))
}

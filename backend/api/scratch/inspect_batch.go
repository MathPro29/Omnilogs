//go:build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"omnilogs-api/configs"
	"omnilogs-api/models"

	"github.com/elastic/go-elasticsearch/v8"
)

func main() {
	os.Setenv("ENV_PATH", "../.env")
	env := configs.LoadEnv()
	db, err := configs.ConnectDB(env)
	if err != nil {
		log.Fatalf("connect DB failed: %v", err)
	}

	batchID := "ed9d0636-9ed3-4777-befd-d0796bd76179"
	var refs []models.LogIndexRef
	if err := db.Where("batch_id = ?", batchID).Find(&refs).Error; err != nil {
		log.Fatalf("query index refs failed: %v", err)
	}

	fmt.Printf("=== Index Refs for Batch %s ===\n", batchID)
	for _, ref := range refs {
		fmt.Printf("LogID: %s, ProductID: %d, EnvID: %d, ProjectID: %v, CategoryID: %v, Index: %s, Timestamp: %v\n",
			ref.LogID, ref.ProductID, ref.EnvironmentID, ref.ProjectID, ref.CategoryID, ref.ElasticIndex, ref.Timestamp)
	}

	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{env.ElasticURL},
	})
	if err != nil {
		log.Fatalf("connect ES failed: %v", err)
	}

	query := fmt.Sprintf(`{"query": {"term": {"batch_id.keyword": "%s"}}}`, batchID)
	if len(refs) > 0 {
		query = fmt.Sprintf(`{"query": {"term": {"_id": "%s"}}}`, refs[0].LogID)
	}

	res, err := esClient.Search(
		esClient.Search.WithContext(context.Background()),
		esClient.Search.WithIndex("logs-17-29*"),
		esClient.Search.WithBody(jsonReader(query)),
	)
	if err != nil {
		log.Fatalf("search ES failed: %v", err)
	}
	defer res.Body.Close()

	var payload struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				ID     string         `json:"_id"`
				Source map[string]any `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		log.Fatalf("decode ES response failed: %v", err)
	}

	fmt.Printf("\n=== Elasticsearch Documents (Total: %d) ===\n", payload.Hits.Total.Value)
	for _, hit := range payload.Hits.Hits {
		docJSON, _ := json.MarshalIndent(hit.Source, "", "  ")
		fmt.Printf("ID: %s\nDocument:\n%s\n", hit.ID, string(docJSON))
	}
}

func jsonReader(s string) *os.File {
	f, _ := os.CreateTemp("", "es-query-*.json")
	f.WriteString(s)
	f.Seek(0, 0)
	return f
}

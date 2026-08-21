//go:build ignore

package main

import (
	"context"
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

	res, err := es.Cat.Indices(
		es.Cat.Indices.WithContext(context.Background()),
		es.Cat.Indices.WithFormat("json"),
	)
	if err != nil {
		log.Fatalf("Error getting indices: %s", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("failed to read body: %v", err)
	}

	fmt.Println("=== Elasticsearch Indices ===")
	fmt.Println(string(body))
}

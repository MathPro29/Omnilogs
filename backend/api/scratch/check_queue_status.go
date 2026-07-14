package main

import (
	"fmt"
	"log"
	"os"

	"omnilogs-api/configs"
)

func main() {
	os.Setenv("ENV_PATH", "../.env")
	env := configs.LoadEnv()
	db, err := configs.ConnectDB(env)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}

	type BatchCount struct {
		Status string
		Count  int64
	}

	var results []BatchCount
	db.Table("log_queue_batches").Select("status, count(*) as count").Group("status").Scan(&results)

	fmt.Println("=== Log Queue Batches Status Count ===")
	for _, r := range results {
		fmt.Printf("Status: %s, Count: %d\n", r.Status, r.Count)
	}
}

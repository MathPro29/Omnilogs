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

	var count int64
	db.Table("log_index_refs").Count(&count)

	fmt.Printf("=== Total Log Index Refs in DB: %d ===\n", count)
}

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
	db.Table("log_failures").Count(&count)

	fmt.Printf("=== Total Log Failures: %d ===\n", count)
}

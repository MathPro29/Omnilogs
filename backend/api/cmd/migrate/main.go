package main

import (
	"log"

	"omnilogs-api/configs"
	"omnilogs-api/migrate"
)

func main() {
	env := configs.LoadEnv()
	if err := env.Validate(); err != nil {
		log.Fatalf("invalid environment configuration: %v", err)
	}

	db, err := configs.ConnectDB(env)
	if err != nil {
		log.Fatalf("connect database failed: %v", err)
	}

	migrate.Migrate(db, env)
	log.Println("migration completed")
}

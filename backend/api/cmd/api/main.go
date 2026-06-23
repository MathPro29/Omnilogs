package main

import (
	"log"

	"omnilogs-api/configs"
	"omnilogs-api/internal/bootstrap"
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

	if err := bootstrap.RunAPIServer(env, db); err != nil {
		log.Fatalf("api server shutdown failed: %v", err)
	}
}

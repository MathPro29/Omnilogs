package main

import (
	"log"
	"log/slog"
	"os"

	"omnilogs-api/configs"
	"omnilogs-api/internal/bootstrap"
	"omnilogs-api/migrate"
)

func main() {
	env := configs.LoadEnv()
	if err := env.Validate(); err != nil {
		log.Fatalf("invalid environment configuration: %v", err)
	}

	configs.InitLogger(env)

	db, err := configs.ConnectDB(env)
	if err != nil {
		slog.Error("connect database failed", "error", err)
		os.Exit(1)
	}

	migrate.Migrate(db, env)

	if err := bootstrap.RunAPIServer(env, db); err != nil {
		slog.Error("api server shutdown failed", "error", err)
		os.Exit(1)
	}
}

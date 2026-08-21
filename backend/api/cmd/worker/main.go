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

	// Local worker runs do not always have the compose migration job in front
	// of them. Keep the worker safe to start on a fresh database as well.
	migrate.Migrate(db, env)

	if err := bootstrap.RunWorker(env, db); err != nil {
		slog.Error("worker shutdown failed", "error", err)
		os.Exit(1)
	}
}

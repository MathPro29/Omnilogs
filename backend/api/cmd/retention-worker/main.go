package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"omnilogs-api/configs"
	"omnilogs-api/internal/retention_worker/usecase"
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
	es, err := configs.ConnectElasticsearch(env)
	if err != nil {
		slog.Error("connect Elasticsearch failed", "error", err)
		os.Exit(1)
	}
	poll := time.Duration(env.RetentionWorkerPollSeconds) * time.Second
	lead := time.Duration(env.RetentionWorkerLeadSeconds) * time.Second
	worker := usecase.New(db, es, poll, lead)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := worker.Run(ctx); err != nil {
		slog.Error("retention worker stopped", "error", err)
		os.Exit(1)
	}
}

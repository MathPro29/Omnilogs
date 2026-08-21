package bootstrap

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"omnilogs-api/configs"
	retentionworker "omnilogs-api/internal/retention_worker/usecase"

	"gorm.io/gorm"
)

func RunRetentionWorker(env *configs.Env, db *gorm.DB) error {
	slog.Info("omnilogs retention worker starting", "env", env.AppEnv)
	es, err := configs.ConnectElasticsearch(env)
	if err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	worker := retentionworker.New(db, es,
		time.Duration(env.RetentionWorkerPollSeconds)*time.Second,
		time.Duration(env.RetentionWorkerLeadSeconds)*time.Second,
	)
	return worker.Run(ctx)
}

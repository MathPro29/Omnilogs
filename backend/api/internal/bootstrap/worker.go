package bootstrap

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"omnilogs-api/configs"
	workerprocessor "omnilogs-api/internal/worker"

	"gorm.io/gorm"
)

func RunWorker(env *configs.Env, db *gorm.DB) error {
	slog.Info("omnilogs worker starting", "env", env.AppEnv)
	slog.Info("worker pulls pending logs from PostgreSQL queue and indexes them into Elasticsearch")

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	esClient, err := configs.ConnectElasticsearch(env)
	if err != nil {
		return err
	}

	processor := workerprocessor.NewProcessor(db, esClient, env.DataEncryptionKey)
	done := make(chan error, 1)
	go func() {
		// ลูปนี้จะคอยหยิบ batch จากคิวมาประมวลผลและส่งเข้า pipeline ของการ index
		// ไปเรื่อย ๆ จนกว่าจะมีการสั่ง shutdown
		done <- processor.Run(ctx)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-stop:
		slog.Info("worker received shutdown signal", "signal", sig.String())
		cancel()
	case err := <-done:
		cancel()
		if err != nil {
			return err
		}
	}

	select {
	case err := <-done:
		if err != nil {
			return err
		}
	default:
	}

	if err := sqlDB.Close(); err != nil {
		return err
	}

	slog.Info("worker shutdown completed")
	return nil
}

package bootstrap

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"omnilogs-api/configs"
	workerprocessor "omnilogs-api/internal/worker"

	"gorm.io/gorm"
)

func RunWorker(env *configs.Env, db *gorm.DB) error {
	log.Printf("omnilogs worker starting in %s", env.AppEnv)
	log.Println("worker ตัวนี้มีหน้าที่ดึง log ที่ค้างในคิวจาก PostgreSQL แล้วส่งต่อไป index ใน Elasticsearch")

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	processor := workerprocessor.NewProcessor(db, env.ElasticURL)
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
		log.Printf("worker received shutdown signal: %s", sig)
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

	log.Println("worker shutdown completed")
	return nil
}

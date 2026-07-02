package configs

import (
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB(env *Env) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Bangkok",
		env.DBHost,
		env.DBUsername,
		env.DBPassword,
		env.DBName,
		env.DBPort,
	)

	var db *gorm.DB
	var err error
	maxRetries := 15
	retryInterval := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			// Successfully connected
			break
		}

		slog.Warn("failed to connect to database, retrying...", 
			"attempt", i+1, 
			"max_retries", maxRetries, 
			"error", err,
		)
		time.Sleep(retryInterval)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(env.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(env.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(env.DBConnMaxLifetimeSeconds) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(env.DBConnMaxIdleTimeSeconds) * time.Second)

	return db, nil
}

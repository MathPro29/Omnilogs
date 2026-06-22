package configs

import (
	"fmt"

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
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})

}

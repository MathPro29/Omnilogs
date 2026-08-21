package main

import (
	"log"

	"log/slog"
	"omnilogs-api/configs"
	"omnilogs-api/internal/bootstrap"
)

func main() {
	env := configs.LoadEnv()
	if err := env.Validate(); err != nil {
		log.Fatalf("invalid environment configuration: %v", err)
	}

	// ตั้งค่า Logger ของระบบ
	configs.InitLogger(env)

	// ทิ้งข้อความแรกให้เห็นว่า Worker กำลังจะเริ่มทำงานแล้ว
	slog.Info("worker starting",
		"env", env.AppEnv,
		"port", env.AppPort,
		"version", "1.0.0", // ใส่เวอร์ชันโปรแกรมตรงนี้ได้เลย
	)

	db, err := configs.ConnectDB(env)
	if err != nil {
		log.Fatalf("connect database failed: %v", err)
	}

	if err := bootstrap.RunAPIServer(env, db); err != nil {
		log.Fatalf("api server shutdown failed: %v", err)
	}
}

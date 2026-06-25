package configs

import (
	"log/slog"
	"os"
)

func InitLogger(env *Env) {
	var handler slog.Handler

	if env.AppEnv == "production" {
		// พ่นเป็น JSON สำหรับเซิร์ฟเวอร์
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo, // เก็บล็อกระดับ Info ขึ้นไป
		})
	} else {
		// พ่นเป็น Text ที่อ่านง่ายสำหรับตอนรันในเครื่อง (Local)
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug, // แสดงสัญลักษณ์และคำอธิบายสีต่างๆ (ถ้ามี)
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger) // ตั้งให้เป็น Logger หลักของระบบ
}

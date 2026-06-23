package scratch

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type SystemAuditLog struct {
	AuditID   string          `json:"audit_id"`
	UserID    *int            `json:"user_id"`
	ProductID *int            `json:"product_id"`
	Action    string          `json:"action"`
	Method    *string         `json:"method"`
	URL       *string         `json:"url"`
	IPAddress *string         `json:"ip_address"`
	QueryJSON json.RawMessage `json:"query_json"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt string          `json:"created_at"`
}

func DumpAuditLogs() {
	_ = godotenv.Load("../.env") // load from root/parent if present
	_ = godotenv.Load(".env")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Bangkok",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "admin"),
		getEnv("DB_NAME", "omnilogs"),
		getEnv("DB_PORT", "5499"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	var logs []SystemAuditLog
	if err := db.Table("system_audit_logs").Order("created_at desc").Limit(100).Find(&logs).Error; err != nil {
		log.Fatalf("failed to query logs: %v", err)
	}

	fmt.Println("Latest 10 Audit Logs:")
	for _, l := range logs {
		payloadStr := ""
		if len(l.Payload) > 0 {
			payloadStr = string(l.Payload)
		}
		userIdVal := "NULL"
		if l.UserID != nil {
			userIdVal = fmt.Sprintf("%d", *l.UserID)
		}
		methodVal := ""
		if l.Method != nil {
			methodVal = *l.Method
		}
		urlVal := ""
		if l.URL != nil {
			urlVal = *l.URL
		}
		fmt.Printf("AuditID: %s, UserID: %s, Action: %s, Method: %s, URL: %s, Payload: %s, CreatedAt: %s\n",
			l.AuditID, userIdVal, l.Action, methodVal, urlVal, payloadStr, l.CreatedAt)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

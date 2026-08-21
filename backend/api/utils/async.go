package utils

import (
	"context"
	"log"

	"gorm.io/gorm"
)

// SafeGo runs a function in a background goroutine safely, catching and logging any panic.
func SafeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[SYSTEM PANIC / ระบบล่ม] Recovered from panic in background goroutine: %v", r)
			}
		}()
		fn()
	}()
}

// GetBackgroundDB returns a copy of GORM DB instance configured with a fresh background context.
// This is critical when calling DB queries in background goroutines (like saving audit logs or sending emails)
// because the HTTP request context (c.Request.Context()) will be cancelled immediately after the response is sent.
func GetBackgroundDB(db *gorm.DB) *gorm.DB {
	if db == nil {
		return nil
	}
	// We use context.Background() so that GORM operations do not get cancelled when the HTTP request finishes.
	return db.WithContext(context.Background())
}

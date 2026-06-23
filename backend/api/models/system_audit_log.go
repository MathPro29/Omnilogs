package models

import (
	"encoding/json"
	"time"
)

type SystemAuditLog struct {
	AuditID   string          `gorm:"type:uuid;primaryKey" json:"audit_id"`
	UserID    *int            `json:"user_id,omitempty"`
	ProductID *int            `json:"product_id,omitempty"`
	Action    string          `gorm:"not null" json:"action"`
	Method    *string         `json:"method,omitempty"`
	URL       *string         `gorm:"type:text" json:"url,omitempty"`
	IPAddress *string         `json:"ip_address,omitempty"`
	QueryJSON  json.RawMessage `gorm:"type:jsonb" json:"query_json,omitempty"`
	Payload    json.RawMessage `gorm:"type:jsonb" json:"payload,omitempty"`
	StatusCode *int            `json:"status_code,omitempty"`
	CreatedAt  *time.Time      `gorm:"type:timestamptz" json:"created_at,omitempty"`
}

package models

import (
	"encoding/json"
	"time"
)

type ProductAPIKey struct {
	KeyID         int             `gorm:"primaryKey;autoIncrement" json:"key_id"`
	ProductID     int             `gorm:"not null;index:idx_api_key_product_environment_active,priority:1" json:"product_id"`
	EnvironmentID *int            `gorm:"index:idx_api_key_product_environment_active,priority:2" json:"environment_id,omitempty"`
	KeyName       *string         `json:"key_name,omitempty"`
	KeyPrefix     string          `gorm:"not null;index" json:"key_prefix"`
	KeyHash       string          `gorm:"not null;unique" json:"-"`
	Permissions   json.RawMessage `gorm:"type:jsonb;not null" json:"permissions"`
	IsActive      bool            `gorm:"not null;default:true;index:idx_api_key_product_environment_active,priority:3" json:"is_active"`
	LastUsedAt    *time.Time      `gorm:"type:timestamptz" json:"last_used_at,omitempty"`
	ExpiresAt     *time.Time      `gorm:"type:timestamptz" json:"expires_at,omitempty"`
	RevokedAt     *time.Time      `gorm:"type:timestamptz" json:"revoked_at,omitempty"`
	Timestamps
}

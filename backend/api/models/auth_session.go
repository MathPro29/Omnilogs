package models

import "time"

type AuthSession struct {
	SessionID        int        `gorm:"primaryKey;autoIncrement" json:"session_id"`
	UserID           int        `gorm:"not null;index" json:"user_id"`
	RefreshTokenHash string     `gorm:"not null;uniqueIndex" json:"-"`
	UserAgent        *string    `gorm:"type:text" json:"user_agent,omitempty"`
	IPAddress        *string    `gorm:"type:text" json:"ip_address,omitempty"`
	ExpiresAt        time.Time  `gorm:"type:timestamptz;not null;index" json:"expires_at"`
	RevokedAt        *time.Time `gorm:"type:timestamptz" json:"revoked_at,omitempty"`
	LastUsedAt       *time.Time `gorm:"type:timestamptz" json:"last_used_at,omitempty"`
	Timestamps
}

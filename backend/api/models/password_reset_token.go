package models

import "time"

type PasswordResetToken struct {
	PasswordResetTokenID int        `gorm:"primaryKey;autoIncrement" json:"password_reset_token_id"`
	UserID               int        `gorm:"not null;index" json:"user_id"`
	TokenHash            string     `gorm:"not null;unique" json:"-"`
	ExpiresAt            time.Time  `gorm:"type:timestamptz;not null;index" json:"expires_at"`
	UsedAt               *time.Time `gorm:"type:timestamptz" json:"used_at,omitempty"`
	Timestamps
}

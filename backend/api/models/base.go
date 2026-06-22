package models

import "time"

type Timestamps struct {
	CreatedAt *time.Time `gorm:"type:timestamptz" json:"created_at,omitempty"`
	UpdatedAt *time.Time `gorm:"type:timestamptz" json:"updated_at,omitempty"`
}

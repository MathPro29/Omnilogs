package models

import "time"

type PlatformMembership struct {
	PlatformMembershipID int        `gorm:"primaryKey;autoIncrement" json:"platform_membership_id"`
	UserID               int        `gorm:"not null;uniqueIndex:uq_platform_membership_user" json:"user_id"`
	PlatformRoleID       int        `gorm:"not null" json:"platform_role_id"`
	IsActive             bool       `gorm:"not null;default:true" json:"is_active"`
	ExpiresAt            *time.Time `gorm:"type:timestamptz" json:"expires_at,omitempty"`
	Timestamps
}

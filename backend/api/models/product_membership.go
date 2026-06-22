package models

import "time"

type ProductMembership struct {
	MembershipID int        `gorm:"primaryKey;autoIncrement" json:"membership_id"`
	UserID       int        `gorm:"not null;uniqueIndex:uq_product_membership,priority:1" json:"user_id"`
	ProductID    int        `gorm:"not null;uniqueIndex:uq_product_membership,priority:2;index:idx_membership_product_active,priority:1" json:"product_id"`
	RoleID       int        `gorm:"not null;uniqueIndex:uq_product_membership,priority:3" json:"role_id"`
	ExpiresAt    *time.Time `gorm:"type:timestamptz" json:"expires_at,omitempty"`
	IsActive     bool       `gorm:"not null;default:true;index:idx_membership_product_active,priority:2" json:"is_active"`
	Timestamps
}

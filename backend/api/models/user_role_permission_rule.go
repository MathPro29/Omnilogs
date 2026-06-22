package models

import "time"

type UserRolePermissionRule struct {
	PermissionRuleID int        `gorm:"primaryKey;autoIncrement" json:"permission_rule_id"`
	UserID           int        `gorm:"not null;uniqueIndex:uq_permission_rule,priority:1;index:idx_rule_user_active,priority:1;index:idx_rule_product_user,priority:2;index:idx_rule_role_user,priority:2" json:"user_id"`
	ProductID        *int       `gorm:"uniqueIndex:uq_permission_rule,priority:2;index:idx_rule_product_user,priority:1" json:"product_id,omitempty"`
	RoleID           *int       `gorm:"uniqueIndex:uq_permission_rule,priority:3;index:idx_rule_role_user,priority:1" json:"role_id,omitempty"`
	ProjectID        *int       `gorm:"uniqueIndex:uq_permission_rule,priority:4" json:"project_id,omitempty"`
	CategoryID       *int       `gorm:"uniqueIndex:uq_permission_rule,priority:5" json:"category_id,omitempty"`
	ResourceType     string     `gorm:"not null;uniqueIndex:uq_permission_rule,priority:6;index:idx_rule_resource_action,priority:1" json:"resource_type"`
	Action           string     `gorm:"not null;uniqueIndex:uq_permission_rule,priority:7;index:idx_rule_resource_action,priority:2" json:"action"`
	Effect           string     `gorm:"not null;index" json:"effect"`
	ScopeLevel       string     `gorm:"not null" json:"scope_level"`
	GrantedBy        *int       `json:"granted_by,omitempty"`
	IsActive         bool       `gorm:"not null;default:true;index:idx_rule_user_active,priority:2" json:"is_active"`
	ExpiresAt        *time.Time `gorm:"type:timestamptz" json:"expires_at,omitempty"`
	Timestamps
}
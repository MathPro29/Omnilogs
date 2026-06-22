package models

import "encoding/json"

type PlatformRole struct {
	PlatformRoleID int             `gorm:"primaryKey;autoIncrement" json:"platform_role_id"`
	RoleCode       string          `gorm:"not null;uniqueIndex:platform_roles_role_code_key" json:"role_code"`
	RoleName       string          `gorm:"not null" json:"role_name"`
	Description    *string         `gorm:"type:text" json:"description,omitempty"`
	Permissions    json.RawMessage `gorm:"type:jsonb;not null" json:"permissions"`
	IsSystemRole   bool            `gorm:"not null;default:false" json:"is_system_role"`
	IsActive       bool            `gorm:"not null;default:true" json:"is_active"`
	Timestamps
}

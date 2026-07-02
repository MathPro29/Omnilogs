package models

import "encoding/json"

type ProductRole struct {
	RoleID            int                     `gorm:"primaryKey;autoIncrement;uniqueIndex:uq_role_product,priority:1" json:"role_id"`
	ProductID         int                     `gorm:"not null;uniqueIndex:uq_role_product,priority:2;uniqueIndex:uq_product_role_code,priority:1" json:"product_id"`
	TemplateID        *int                    `json:"template_id,omitempty"`
	RoleCode          string                  `gorm:"not null;uniqueIndex:uq_product_role_code,priority:2" json:"role_code"`
	RoleName          string                  `gorm:"not null" json:"role_name"`
	LegacyPermissions json.RawMessage         `gorm:"column:permissions;type:jsonb;not null;default:'[]'" json:"-"`
	Permissions       []ProductRolePermission `gorm:"foreignKey:RoleID;references:RoleID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"permissions"`
	IsDefault         bool                    `gorm:"not null;default:false" json:"is_default"`
	IsActive          bool                    `gorm:"not null;default:true" json:"is_active"`
	Timestamps
}

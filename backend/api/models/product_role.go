package models

import "encoding/json"

type ProductRole struct {
	RoleID      int             `gorm:"primaryKey;autoIncrement;uniqueIndex:uq_role_product,priority:1" json:"role_id"`
	ProductID   int             `gorm:"not null;uniqueIndex:uq_role_product,priority:2;uniqueIndex:uq_product_role_code,priority:1" json:"product_id"`
	TemplateID  *int            `json:"template_id,omitempty"`
	RoleCode    string          `gorm:"not null;uniqueIndex:uq_product_role_code,priority:2" json:"role_code"`
	RoleName    string          `gorm:"not null" json:"role_name"`
	Permissions json.RawMessage `gorm:"type:jsonb;not null" json:"permissions"`
	Timestamps
}

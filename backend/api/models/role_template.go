package models

type RoleTemplate struct {
	TemplateID       int     `gorm:"primaryKey;autoIncrement" json:"template_id"`
	RoleCode         string  `gorm:"not null;uniqueIndex" json:"role_code"`
	RoleName         string  `gorm:"not null" json:"role_name"`
	Description      *string `gorm:"type:text" json:"description,omitempty"`
	IsSystemTemplate bool    `gorm:"not null;default:false" json:"is_system_template"`
	IsActive         bool    `gorm:"not null;default:true" json:"is_active"`
	Timestamps
}

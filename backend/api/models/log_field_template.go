package models

type LogFieldTemplate struct {
	TemplateID    int    `gorm:"primaryKey;autoIncrement" json:"template_id"`
	TemplateName  string `gorm:"not null;uniqueIndex" json:"template_name"`
	Description   *string `json:"description,omitempty"`
	Category      string `gorm:"index" json:"category"` // e.g., HTTP, AUTH, PAYMENT
	FieldsJSON    []byte `gorm:"type:jsonb" json:"fields_json"` // Store the array of field configs
	IsSystem      bool   `gorm:"not null;default:false" json:"is_system"`
	IsActive      bool   `gorm:"not null;default:true" json:"is_active"`
	Timestamps
}



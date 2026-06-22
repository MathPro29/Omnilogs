package models

type LogFieldEnumOption struct {
	OptionID          int    `gorm:"primaryKey;autoIncrement" json:"option_id"`
	FieldDefinitionID int    `gorm:"not null;uniqueIndex:uq_field_option,priority:1;index:idx_option_active,priority:1;index:idx_option_order,priority:1" json:"field_definition_id"`
	OptionKey         string `gorm:"not null;uniqueIndex:uq_field_option,priority:2" json:"option_key"`
	OptionLabel       string `gorm:"not null" json:"option_label"`
	OptionValue       string `gorm:"not null" json:"option_value"`
	DisplayOrder      *int   `gorm:"index:idx_option_order,priority:2" json:"display_order,omitempty"`
	IsDefault         bool   `gorm:"not null;default:false" json:"is_default"`
	IsActive          bool   `gorm:"not null;default:true;index:idx_option_active,priority:2" json:"is_active"`
	Timestamps
}
package models

import "encoding/json"

type LogFieldValueSource struct {
	ValueSourceID     int             `gorm:"primaryKey;autoIncrement" json:"value_source_id"`
	FieldDefinitionID int             `gorm:"not null;uniqueIndex:uq_field_value_source,priority:1;index:idx_value_source_active,priority:1" json:"field_definition_id"`
	SourceType        string          `gorm:"not null;index:idx_source_type_key,priority:1" json:"source_type"`
	SourceKey         string          `gorm:"not null;uniqueIndex:uq_field_value_source,priority:2;index:idx_source_type_key,priority:2" json:"source_key"`
	SourceConfig      json.RawMessage `gorm:"type:jsonb" json:"source_config,omitempty"`
	LabelField        *string         `json:"label_field,omitempty"`
	ValueField        *string         `json:"value_field,omitempty"`
	IsRequired        bool            `gorm:"not null;default:false" json:"is_required"`
	TimeoutSeconds    *int            `json:"timeout_seconds,omitempty"`
	RefreshMode       string          `gorm:"not null;default:ON_DEMAND" json:"refresh_mode"`
	IsActive          bool            `gorm:"not null;default:true;index;index:idx_value_source_active,priority:2" json:"is_active"`
	Timestamps
}

package models

type LogFieldDefinition struct {
	FieldDefinitionID    int     `gorm:"primaryKey;autoIncrement" json:"field_definition_id"`
	ProductID            *int    `gorm:"uniqueIndex:uq_field_definition,priority:1;index:idx_field_product_active,priority:1" json:"product_id,omitempty"`
	ProjectID            *int    `gorm:"uniqueIndex:uq_field_definition,priority:2" json:"project_id,omitempty"`
	CategoryID           *int    `gorm:"uniqueIndex:uq_field_definition,priority:3" json:"category_id,omitempty"`
	FieldKey             string  `gorm:"not null;uniqueIndex:uq_field_definition,priority:4" json:"field_key"`
	DisplayName          *string `json:"display_name,omitempty"`
	Description          *string `gorm:"type:text" json:"description,omitempty"`
	SourceSection        string  `gorm:"not null;index:idx_source_field_path,priority:1" json:"source_section"`
	FieldPath            *string `gorm:"index:idx_source_field_path,priority:2" json:"field_path,omitempty"`
	ElasticFieldName     string  `gorm:"not null;index" json:"elastic_field_name"`
	DataType             string  `gorm:"not null" json:"data_type"`
	ValueSourceType      string  `gorm:"not null;default:NONE;index:idx_value_source,priority:1" json:"value_source_type"`
	ValueSourceKey       *string `gorm:"index:idx_value_source,priority:2" json:"value_source_key,omitempty"`
	IsRequired           bool    `gorm:"not null;default:false" json:"is_required"`
	IsSensitive          bool    `gorm:"not null;default:false" json:"is_sensitive"`
	MaskBeforeIndex      bool    `gorm:"not null;default:false" json:"mask_before_index"`
	EncryptBeforeArchive bool    `gorm:"not null;default:false" json:"encrypt_before_archive"`
	IsVisible            bool    `gorm:"not null;default:true" json:"is_visible"`
	IsSearchable         bool    `gorm:"not null;default:false;index" json:"is_searchable"`
	IsFilterable         bool    `gorm:"not null;default:false;index" json:"is_filterable"`
	IsSortable           bool    `gorm:"not null;default:false" json:"is_sortable"`
	IsAggregatable       bool    `gorm:"not null;default:false;index" json:"is_aggregatable"`
	DisplayOrder         *int    `json:"display_order,omitempty"`
	DefaultValue         *string `gorm:"type:text" json:"default_value,omitempty"`
	IsActive             bool    `gorm:"not null;default:true;index:idx_field_product_active,priority:2" json:"is_active"`
	Timestamps
}
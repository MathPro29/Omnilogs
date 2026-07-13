package dto

import (
	"encoding/json"
	"omnilogs-api/models"
)

type CreateLogFieldDefinitionRequest struct {
	ProductID            *int    `json:"product_id,omitempty" binding:"omitempty,gt=0"`
	ProjectID            *int    `json:"project_id,omitempty" binding:"omitempty,gt=0"`
	CategoryID           *int    `json:"category_id,omitempty" binding:"omitempty,gt=0"`
	FieldKey             string  `json:"field_key" binding:"required"`
	DisplayName          *string `json:"display_name,omitempty"`
	Description          *string `json:"description,omitempty"`
	SourceSection        string  `json:"source_section" binding:"required"`
	FieldPath            *string `json:"field_path,omitempty"`
	ElasticFieldName     string  `json:"elastic_field_name" binding:"required"`
	DataType             string  `json:"data_type" binding:"required"`
	ValueSourceType      string  `json:"value_source_type" binding:"required"`
	ValueSourceKey       *string `json:"value_source_key,omitempty"`
	IsRequired           bool    `json:"is_required"`
	IsSensitive          bool    `json:"is_sensitive"`
	MaskBeforeIndex      bool    `json:"mask_before_index"`
	EncryptBeforeArchive bool    `json:"encrypt_before_archive"`
	IsVisible            *bool   `json:"is_visible,omitempty"`
	IsSearchable         bool    `json:"is_searchable"`
	IsFilterable         bool    `json:"is_filterable"`
	IsSortable           bool    `json:"is_sortable"`
	IsAggregatable       bool    `json:"is_aggregatable"`
	DisplayOrder         *int                    `json:"display_order,omitempty" binding:"omitempty,gte=0"`
	DefaultValue         *string                 `json:"default_value,omitempty"`
	FieldType            *string                 `json:"field_type,omitempty"`
	ConfigJSON           *models.FieldConfigJSON `json:"config_json,omitempty"`
}

type UpdateLogFieldDefinitionRequest struct {
	DisplayName          *string `json:"display_name,omitempty"`
	Description          *string `json:"description,omitempty"`
	FieldPath            *string `json:"field_path,omitempty"`
	ElasticFieldName     *string `json:"elastic_field_name,omitempty"`
	DataType             *string `json:"data_type,omitempty"`
	ValueSourceType      *string `json:"value_source_type,omitempty"`
	ValueSourceKey       *string `json:"value_source_key,omitempty"`
	IsRequired           *bool   `json:"is_required,omitempty"`
	IsSensitive          *bool   `json:"is_sensitive,omitempty"`
	MaskBeforeIndex      *bool   `json:"mask_before_index,omitempty"`
	EncryptBeforeArchive *bool   `json:"encrypt_before_archive,omitempty"`
	IsVisible            *bool   `json:"is_visible,omitempty"`
	IsSearchable         *bool   `json:"is_searchable,omitempty"`
	IsFilterable         *bool   `json:"is_filterable,omitempty"`
	IsSortable           *bool   `json:"is_sortable,omitempty"`
	IsAggregatable       *bool   `json:"is_aggregatable,omitempty"`
	DisplayOrder         *int                    `json:"display_order,omitempty" binding:"omitempty,gte=0"`
	DefaultValue         *string                 `json:"default_value,omitempty"`
	IsActive             *bool                   `json:"is_active,omitempty"`
	FieldType            *string                 `json:"field_type,omitempty"`
	ConfigJSON           *models.FieldConfigJSON `json:"config_json,omitempty"`
}

type LogFieldDefinitionResponse struct {
	FieldDefinitionID    int     `json:"field_definition_id"`
	ProductID            *int    `json:"product_id,omitempty"`
	ProjectID            *int    `json:"project_id,omitempty"`
	CategoryID           *int    `json:"category_id,omitempty"`
	FieldKey             string  `json:"field_key"`
	DisplayName          *string `json:"display_name,omitempty"`
	Description          *string `json:"description,omitempty"`
	SourceSection        string  `json:"source_section"`
	FieldPath            *string `json:"field_path,omitempty"`
	ElasticFieldName     string  `json:"elastic_field_name"`
	DataType             string  `json:"data_type"`
	ValueSourceType      string  `json:"value_source_type"`
	ValueSourceKey       *string `json:"value_source_key,omitempty"`
	IsRequired           bool    `json:"is_required"`
	IsSensitive          bool    `json:"is_sensitive"`
	MaskBeforeIndex      bool    `json:"mask_before_index"`
	EncryptBeforeArchive bool    `json:"encrypt_before_archive"`
	IsVisible            bool    `json:"is_visible"`
	IsSearchable         bool    `json:"is_searchable"`
	IsFilterable         bool    `json:"is_filterable"`
	IsSortable           bool    `json:"is_sortable"`
	IsAggregatable       bool    `json:"is_aggregatable"`
	DisplayOrder         *int                    `json:"display_order,omitempty"`
	DefaultValue         *string                 `json:"default_value,omitempty"`
	IsActive             bool                    `json:"is_active"`
	SchemaVersion        int                     `json:"schema_version"`
	FieldType            *string                 `json:"field_type,omitempty"`
	ConfigJSON           *models.FieldConfigJSON `json:"config_json,omitempty"`
	TimestampResponse
}

type CreateLogFieldEnumOptionRequest struct {
	FieldDefinitionID int    `json:"field_definition_id,omitempty"`
	OptionKey         string `json:"option_key" binding:"required"`
	OptionLabel       string `json:"option_label" binding:"required"`
	OptionValue       string  `json:"option_value" binding:"required"`
	DisplayOrder      *int    `json:"display_order,omitempty" binding:"omitempty,gte=0"`
	ColorCode         *string `json:"color_code,omitempty"`
	Description       *string `json:"description,omitempty"`
	IsDefault         bool    `json:"is_default"`
}

type UpdateLogFieldEnumOptionRequest struct {
	OptionKey    *string `json:"option_key,omitempty"`
	OptionLabel  *string `json:"option_label,omitempty"`
	OptionValue  *string `json:"option_value,omitempty"`
	DisplayOrder *int    `json:"display_order,omitempty" binding:"omitempty,gte=0"`
	ColorCode    *string `json:"color_code,omitempty"`
	Description  *string `json:"description,omitempty"`
	IsDefault    *bool   `json:"is_default,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
}

type LogFieldEnumOptionResponse struct {
	OptionID int `json:"option_id"`
	CreateLogFieldEnumOptionRequest
	IsActive bool `json:"is_active"`
	TimestampResponse
}

type CreateLogFieldValueSourceRequest struct {
	FieldDefinitionID int             `json:"field_definition_id,omitempty"`
	SourceType        string          `json:"source_type" binding:"required"`
	SourceKey         string          `json:"source_key" binding:"required"`
	SourceConfig      json.RawMessage `json:"source_config,omitempty"`
	LabelField        *string         `json:"label_field,omitempty"`
	ValueField        *string         `json:"value_field,omitempty"`
	IsRequired        bool            `json:"is_required"`
	TimeoutSeconds    *int            `json:"timeout_seconds,omitempty" binding:"omitempty,gt=0"`
	RefreshMode       string          `json:"refresh_mode" binding:"required"`
}

type UpdateLogFieldValueSourceRequest struct {
	SourceType     *string         `json:"source_type,omitempty"`
	SourceKey      *string         `json:"source_key,omitempty"`
	SourceConfig   json.RawMessage `json:"source_config,omitempty"`
	LabelField     *string         `json:"label_field,omitempty"`
	ValueField     *string         `json:"value_field,omitempty"`
	IsRequired     *bool           `json:"is_required,omitempty"`
	TimeoutSeconds *int            `json:"timeout_seconds,omitempty" binding:"omitempty,gt=0"`
	RefreshMode    *string         `json:"refresh_mode,omitempty"`
	IsActive       *bool           `json:"is_active,omitempty"`
}

type LogFieldValueSourceResponse struct {
	ValueSourceID int `json:"value_source_id"`
	CreateLogFieldValueSourceRequest
	IsActive bool `json:"is_active"`
	TimestampResponse
}

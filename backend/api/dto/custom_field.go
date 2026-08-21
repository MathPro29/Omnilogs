package dto

import (
	"encoding/json"
	"omnilogs-api/models"
)

type CustomFieldResponse struct {
	models.LogFieldDefinition
	EnumOptions []models.LogFieldEnumOption `json:"enum_options"`
}

type ParseJSONRequest struct {
	SampleJSON json.RawMessage `json:"sample_json" binding:"required"`
}

type FavoriteJSONFieldRequest struct {
	ProductID    int             `json:"product_id" binding:"required,gt=0"`
	ProjectID    *int            `json:"project_id,omitempty"`
	CategoryID   *int            `json:"category_id,omitempty"`
	FieldPath    string          `json:"field_path" binding:"required"`
	DisplayName  *string         `json:"display_name,omitempty"`
	SampleValue  json.RawMessage `json:"sample_value,omitempty"`
	DetectedType string          `json:"detected_type" binding:"required"`
}

type FavoriteReorderRequest struct {
	ProductID int             `json:"product_id" binding:"required,gt=0"`
	Items     []FavoriteOrder `json:"items" binding:"required,min=1"`
}

type FavoriteOrder struct {
	FieldDefinitionID int `json:"field_definition_id" binding:"required,gt=0"`
	DisplayOrder      int `json:"display_order" binding:"gte=0"`
}

type BulkDeleteCustomFieldRequest struct {
	FieldDefinitionIDs []int `json:"field_definition_ids" binding:"required,min=1"`
}

package dto

import (
	"encoding/json"
	"time"
)

type CreateAPIKeyRequest struct {
	ProductID         int             `json:"product_id,omitempty" binding:"omitempty,gt=0"`
	EnvironmentID     *int            `json:"environment_id,omitempty" binding:"omitempty,gt=0"`
	SourceID          *int            `json:"source_id,omitempty" binding:"omitempty,gt=0"`
	DefaultProjectID  *int            `json:"default_project_id,omitempty" binding:"omitempty,gt=0"`
	DefaultCategoryID *int            `json:"default_category_id,omitempty" binding:"omitempty,gt=0"`
	EnvironmentCode   string          `json:"environment_code,omitempty"`
	KeyName           *string         `json:"key_name,omitempty"`
	Permissions       json.RawMessage `json:"permissions,omitempty"`
	ExpiresAt         *time.Time      `json:"expires_at,omitempty"`
}
type UpdateAPIKeyRequest struct {
	KeyName     *string         `json:"key_name,omitempty"`
	Permissions json.RawMessage `json:"permissions,omitempty"`
	IsActive    *bool           `json:"is_active,omitempty"`
	ExpiresAt   *time.Time      `json:"expires_at,omitempty"`
}
type APIKeyResponse struct {
	KeyID             int             `json:"key_id"`
	ProductID         int             `json:"product_id"`
	EnvironmentID     *int            `json:"environment_id,omitempty"`
	SourceID          *int            `json:"source_id,omitempty" binding:"omitempty,gt=0"`
	DefaultProjectID  *int            `json:"default_project_id,omitempty" binding:"omitempty,gt=0"`
	DefaultCategoryID *int            `json:"default_category_id,omitempty" binding:"omitempty,gt=0"`
	KeyName           *string         `json:"key_name,omitempty"`
	KeyPrefix         string          `json:"key_prefix"`
	Permissions       json.RawMessage `json:"permissions"`
	IsActive          bool            `json:"is_active"`
	ExpiresAt         *time.Time      `json:"expires_at,omitempty"`
	TimestampResponse
}
type CreateAPIKeyResponse struct {
	APIKeyResponse
	APIKey string `json:"api_key"`
}

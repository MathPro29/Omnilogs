package dto

import (
	"encoding/json"
	"time"
)

type CreateProductRequest struct {
	ProductName  string                     `json:"product_name" binding:"required"`
	ProductCode  string                     `json:"product_code" binding:"required"`
	Environments []CreateEnvironmentRequest `json:"environments,omitempty" binding:"omitempty"`
}
type UpdateProductRequest struct {
	ProductName *string `json:"product_name,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}
type ProductResponse struct {
	ProductID           int                   `json:"product_id"`
	ProductName         string                `json:"product_name"`
	ProductCode         string                `json:"product_code"`
	ProductEnvironments []EnvironmentResponse `json:"product_environments,omitempty"`
	IsActive            bool                  `json:"is_active"`
	TimestampResponse
}

type CreateEnvironmentRequest struct {
	ProductID       *int   `json:"product_id,omitempty" binding:"omitempty,gt=0"`
	EnvironmentCode string `json:"environment_code" binding:"required"`
	EnvironmentName string `json:"environment_name" binding:"required"`
}
type UpdateEnvironmentRequest struct {
	EnvironmentName *string `json:"environment_name,omitempty"`
}
type EnvironmentResponse struct {
	EnvironmentID   int        `json:"environment_id"`
	ProductID       int        `json:"product_id"`
	EnvironmentCode string     `json:"environment_code"`
	EnvironmentName string     `json:"environment_name"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
}

type CreateAPIKeyRequest struct {
	ProductID     int             `json:"product_id,omitempty" binding:"omitempty,gt=0"`
	EnvironmentID *int            `json:"environment_id,omitempty" binding:"omitempty,gt=0"`
	KeyName       *string         `json:"key_name,omitempty"`
	Permissions   json.RawMessage `json:"permissions" binding:"required"`
	ExpiresAt     *time.Time      `json:"expires_at,omitempty"`
}
type UpdateAPIKeyRequest struct {
	KeyName     *string         `json:"key_name,omitempty"`
	Permissions json.RawMessage `json:"permissions,omitempty"`
	IsActive    *bool           `json:"is_active,omitempty"`
	ExpiresAt   *time.Time      `json:"expires_at,omitempty"`
}
type APIKeyResponse struct {
	KeyID         int             `json:"key_id"`
	ProductID     int             `json:"product_id"`
	EnvironmentID *int            `json:"environment_id,omitempty"`
	KeyName       *string         `json:"key_name,omitempty"`
	KeyPrefix     string          `json:"key_prefix"`
	Permissions   json.RawMessage `json:"permissions"`
	IsActive      bool            `json:"is_active"`
	ExpiresAt     *time.Time      `json:"expires_at,omitempty"`
	TimestampResponse
}
type CreateAPIKeyResponse struct {
	APIKeyResponse
	APIKey string `json:"api_key"`
}

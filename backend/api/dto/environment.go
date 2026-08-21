package dto

import (
	"time"
)

type CreateEnvironmentRequest struct {
	ProductID       *int   `json:"product_id,omitempty" binding:"omitempty,gt=0"`
	EnvironmentCode string `json:"environment_code" binding:"required"`
	EnvironmentName string `json:"environment_name" binding:"required"`
	EnvironmentType string `json:"environment_type,omitempty"`
}
type UpdateEnvironmentRequest struct {
	EnvironmentName *string `json:"environment_name,omitempty"`
}
type DeleteEnvironmentRequest struct {
	EnvironmentID int `json:"environment_id" binding:"required"`
}

type EnvironmentResponse struct {
	EnvironmentID   int        `json:"environment_id"`
	ProductID       int        `json:"product_id"`
	EnvironmentCode string     `json:"environment_code"`
	EnvironmentName string     `json:"environment_name"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
}

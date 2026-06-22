package dto

import "time"

type PaginationRequest struct {
	Page int `form:"page" binding:"omitempty,min=1"`
	PerPage int `form:"per_page" binding:"omitempty,min=1,max=100"`
}

type PaginationResponse struct {
	Page int `json:"page"`
	PerPage int `json:"per_page"`
	TotalItems int64 `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

type TimestampResponse struct {
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type IDListRequest struct {
	IDs []int `json:"ids" binding:"required,min=1,dive,gt=0"`
}

type StatusResponse struct {
	ID any `json:"id"`
	Status string `json:"status"`
	Message string `json:"message,omitempty"`
}

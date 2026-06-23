package dto

import (
	"encoding/json"
	"time"
)

type AuditLogFilterRequest struct {
	PaginationRequest
	UserID    *int       `form:"user_id" binding:"omitempty,gt=0"`
	ProductID *int       `form:"product_id" binding:"omitempty,gt=0"`
	Action    *string    `form:"action"`
	DateFrom  *time.Time `form:"date_from" time_format:"2006-01-02T15:04:05Z07:00"`
	DateTo    *time.Time `form:"date_to" time_format:"2006-01-02T15:04:05Z07:00"`
}

type SystemAuditLogResponse struct {
	AuditID   string          `json:"audit_id"`
	UserID    *int            `json:"user_id,omitempty"`
	ProductID *int            `json:"product_id,omitempty"`
	Action    string          `json:"action"`
	Method    *string         `json:"method,omitempty"`
	URL       *string         `json:"url,omitempty"`
	IPAddress *string         `json:"ip_address,omitempty"`
	QueryJSON  json.RawMessage `json:"query_json,omitempty"`
	Payload    json.RawMessage `json:"payload,omitempty"`
	StatusCode *int            `json:"status_code,omitempty"`
	CreatedAt  *time.Time      `json:"created_at,omitempty"`
}

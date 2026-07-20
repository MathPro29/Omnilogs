package dto

import (
	"encoding/json"
	"time"
)

type AuditLogFilterRequest struct {
	PaginationRequest
	ActorUserID       *int64     `form:"actor_user_id" binding:"omitempty,gt=0"`
	ProductID         *int64     `form:"product_id" binding:"omitempty,gt=0"`
	AllowedProductIDs []int      `form:"-" json:"-"`
	Action            *string    `form:"action"`
	ResourceType      *string    `form:"resource_type"`
	ResourceID        *string    `form:"resource_id"`
	Result            *string    `form:"result"`
	Keyword           *string    `form:"keyword"`
	DateFrom          *time.Time `form:"date_from" time_format:"2006-01-02T15:04:05Z07:00"`
	DateTo            *time.Time `form:"date_to" time_format:"2006-01-02T15:04:05Z07:00"`
}

type SystemAuditLogResponse struct {
	AuditID      string          `json:"audit_id"`
	ActorUserID  *int64          `json:"actor_user_id,omitempty"`
	ProductID    *int64          `json:"product_id,omitempty"`
	Action       string          `json:"action"`
	ResourceType string          `json:"resource_type"`
	ResourceID   *string         `json:"resource_id,omitempty"`
	RequestID    *string         `json:"request_id,omitempty"`
	TraceID      *string         `json:"trace_id,omitempty"`
	Method       *string         `json:"method,omitempty"`
	Path         *string         `json:"path,omitempty"`
	Result       string          `json:"result"`
	Metadata     json.RawMessage `json:"metadata"`
	IPAddress    *string         `json:"ip_address,omitempty"`
	UserAgent    *string         `json:"user_agent,omitempty"`
	CreatedAt    *time.Time      `json:"created_at,omitempty"`
}

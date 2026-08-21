package dto

import "time"

type CreateLogSourceRequest struct {
	ProductID      *int    `json:"product_id,omitempty" binding:"omitempty,gt=0"`
	SourceCode     string  `json:"source_code" binding:"required"`
	SourceName     string  `json:"source_name" binding:"required"`
	SourceType     string  `json:"source_type" binding:"required"`
	SourcePlatform string  `json:"source_platform" binding:"required"`
	RuntimeName    *string `json:"runtime_name,omitempty"`
	SDKName        *string `json:"sdk_name,omitempty"`
	SDKVersion     *string `json:"sdk_version,omitempty"`
	Description    *string `json:"description,omitempty"`
}

type UpdateLogSourceRequest struct {
	SourceName      *string `json:"source_name,omitempty"`
	SourceType      *string `json:"source_type,omitempty"`
	SourcePlatform  *string `json:"source_platform,omitempty"`
	RuntimeName     *string `json:"runtime_name,omitempty"`
	SDKName         *string `json:"sdk_name,omitempty"`
	SDKVersion      *string `json:"sdk_version,omitempty"`
	DiscoveryStatus *string `json:"discovery_status,omitempty"`
	Description     *string `json:"description,omitempty"`
	IsActive        *bool   `json:"is_active,omitempty"`
}

type LogSourceResponse struct {
	SourceID        int        `json:"source_id"`
	ProductID       *int       `json:"product_id,omitempty"`
	SourceCode      string     `json:"source_code"`
	SourceName      string     `json:"source_name"`
	SourceType      string     `json:"source_type"`
	SourcePlatform  string     `json:"source_platform"`
	RuntimeName     *string    `json:"runtime_name,omitempty"`
	SDKName         *string    `json:"sdk_name,omitempty"`
	SDKVersion      *string    `json:"sdk_version,omitempty"`
	DiscoveryStatus string     `json:"discovery_status"`
	CreatedBySource string     `json:"created_by_source"`
	FirstSeenAt     *time.Time `json:"first_seen_at,omitempty"`
	LastSeenAt      *time.Time `json:"last_seen_at,omitempty"`
	Description     *string    `json:"description,omitempty"`
	IsActive        bool       `json:"is_active"`
	TimestampResponse
}

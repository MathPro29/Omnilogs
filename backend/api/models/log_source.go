package models

import "time"

type LogSource struct {
	SourceID          int        `gorm:"primaryKey;autoIncrement" json:"source_id"`
	ProductID         *int       `gorm:"index:idx_source_type,priority:1;index:idx_source_platform,priority:1;index:idx_source_active,priority:1" json:"product_id,omitempty"`
	EnvironmentID     *int       `gorm:"index" json:"environment_id,omitempty"`
	DefaultProjectID  *int       `gorm:"index" json:"default_project_id,omitempty"`
	DefaultCategoryID *int       `gorm:"index" json:"default_category_id,omitempty"`
	SourceCode        string     `gorm:"not null" json:"source_code"`
	SourceName        string     `gorm:"not null" json:"source_name"`
	SourceType        string     `gorm:"not null;default:UNKNOWN;index:idx_source_type,priority:2" json:"source_type"`
	SourcePlatform    string     `gorm:"not null;default:UNKNOWN;index:idx_source_platform,priority:2" json:"source_platform"`
	RuntimeName       *string    `json:"runtime_name,omitempty"`
	SDKName           *string    `json:"sdk_name,omitempty"`
	SDKVersion        *string    `json:"sdk_version,omitempty"`
	DiscoveryStatus   string     `gorm:"not null;default:AUTO_DISCOVERED;index" json:"discovery_status"`
	CreatedBySource   string     `gorm:"not null;default:SYSTEM" json:"created_by_source"`
	FirstSeenAt       *time.Time `gorm:"type:timestamptz" json:"first_seen_at,omitempty"`
	LastSeenAt        *time.Time `gorm:"type:timestamptz" json:"last_seen_at,omitempty"`
	Description       *string    `gorm:"type:text" json:"description,omitempty"`
	IsActive          bool       `gorm:"not null;default:true;index:idx_source_active,priority:2" json:"is_active"`
	Timestamps
}

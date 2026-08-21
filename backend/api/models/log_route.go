package models

type LogRoute struct {
	RouteID       int    `gorm:"primaryKey;autoIncrement" json:"route_id"`
	ProductID     int    `gorm:"not null;index:idx_log_route_lookup,priority:1" json:"product_id"`
	EnvironmentID int    `gorm:"not null;index:idx_log_route_lookup,priority:2" json:"environment_id"`
	SourceID      *int   `gorm:"index:idx_log_route_lookup,priority:3" json:"source_id,omitempty"`
	RouteKey      string `gorm:"not null" json:"route_key"`
	ProjectID     int    `gorm:"not null;index" json:"project_id"`
	CategoryID    *int   `gorm:"index" json:"category_id,omitempty"`
	Priority      int    `gorm:"not null;default:0" json:"priority"`
	IsActive      bool   `gorm:"not null;default:true;index" json:"is_active"`
	Timestamps
}

package models

import "encoding/json"

type LogRoutingRule struct {
	RuleID           int             `gorm:"primaryKey;autoIncrement" json:"rule_id"`
	ProductID        int             `gorm:"not null;index:idx_log_routing_rule_lookup,priority:1" json:"product_id"`
	EnvironmentID    *int            `gorm:"index:idx_log_routing_rule_lookup,priority:2" json:"environment_id,omitempty"`
	SourceID         *int            `gorm:"index:idx_log_routing_rule_lookup,priority:3" json:"source_id,omitempty"`
	RuleName         string          `gorm:"not null" json:"rule_name"`
	Priority         int             `gorm:"not null;default:0;index" json:"priority"`
	Conditions       json.RawMessage `gorm:"type:jsonb;not null" json:"conditions"`
	TargetProjectID  int             `gorm:"not null;index" json:"target_project_id"`
	TargetCategoryID *int            `gorm:"index" json:"target_category_id,omitempty"`
	IsActive         bool            `gorm:"not null;default:true;index" json:"is_active"`
	Timestamps
}

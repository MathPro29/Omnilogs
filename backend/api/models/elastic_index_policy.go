package models

type ElasticIndexPolicy struct {
	ElasticPolicyID  int     `gorm:"primaryKey;autoIncrement" json:"elastic_policy_id"`
	ProductID        int     `gorm:"not null;uniqueIndex:uq_elastic_policy,priority:1" json:"product_id"`
	EnvironmentID    *int    `gorm:"uniqueIndex:uq_elastic_policy,priority:2" json:"environment_id,omitempty"`
	ProjectID        *int    `gorm:"uniqueIndex:uq_elastic_policy,priority:3" json:"project_id,omitempty"`
	CategoryID       *int    `gorm:"uniqueIndex:uq_elastic_policy,priority:4" json:"category_id,omitempty"`
	IndexPrefix      string  `gorm:"not null;uniqueIndex" json:"index_prefix"`
	IndexPattern     *string `json:"index_pattern,omitempty"`
	WriteAlias       *string `gorm:"uniqueIndex" json:"write_alias,omitempty"`
	ILMPolicyName    *string `gorm:"uniqueIndex" json:"ilm_policy_name,omitempty"`
	RolloverType     string  `gorm:"not null;default:MONTHLY" json:"rollover_type"`
	NumberOfShards   int     `gorm:"not null;default:1" json:"number_of_shards"`
	NumberOfReplicas int     `gorm:"not null;default:1" json:"number_of_replicas"`
	RetentionEnabled bool    `gorm:"not null;default:true" json:"retention_enabled"`
	RetentionDays    *int    `json:"retention_days,omitempty"`
	SchemaVersion    string  `gorm:"not null;default:v1" json:"schema_version"`
	IsActive         bool    `gorm:"not null;default:true;index" json:"is_active"`
	Timestamps
}

package models

type ElasticIndexPolicy struct {
	ElasticPolicyID  int     `gorm:"primaryKey;autoIncrement" json:"elastic_policy_id"`
	ProductID        int     `gorm:"not null;uniqueIndex:uq_elastic_policy,priority:1" json:"product_id"`
	EnvironmentID    *int    `gorm:"uniqueIndex:uq_elastic_policy,priority:2" json:"environment_id,omitempty"`
	IndexPrefix      string  `gorm:"not null;uniqueIndex" json:"index_prefix"`
	IndexPattern     *string `json:"index_pattern,omitempty"`
	WriteAlias       *string `gorm:"uniqueIndex" json:"write_alias,omitempty"`
	RolloverType     string  `gorm:"not null;default:MONTHLY" json:"rollover_type"`
	NumberOfShards   int     `gorm:"not null;default:1" json:"number_of_shards"`
	NumberOfReplicas int     `gorm:"not null;default:1" json:"number_of_replicas"`
	RetentionDays    *int    `json:"retention_days,omitempty"`
	SchemaVersion    string  `gorm:"not null;default:v1" json:"schema_version"`
	IsActive         bool    `gorm:"not null;default:true;index" json:"is_active"`
	Timestamps
}

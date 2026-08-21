package models

type LogSchemaVersion struct {
	SchemaVersionID int     `gorm:"primaryKey;autoIncrement" json:"schema_version_id"`
	ProductID       *int    `gorm:"index:idx_schema_product,priority:1" json:"product_id,omitempty"`
	VersionNumber   int     `gorm:"not null;index:idx_schema_product,priority:2" json:"version_number"`
	Description     *string `json:"description,omitempty"`
	SnapshotJSON    []byte  `gorm:"type:jsonb" json:"snapshot_json"` // Store the entire schema dump
	IsActive        bool    `gorm:"not null;default:true" json:"is_active"`
	Timestamps
}

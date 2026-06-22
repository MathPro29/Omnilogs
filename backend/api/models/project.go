package models

type Project struct {
	ProjectID   int    `gorm:"primaryKey;autoIncrement" json:"project_id"`
	ProductID   int    `gorm:"not null;uniqueIndex:uq_project_code,priority:1" json:"product_id"`
	ProjectCode string `gorm:"not null;uniqueIndex:uq_project_code,priority:2" json:"project_code"`
	ProjectName string `gorm:"not null" json:"project_name"`
	IsActive    bool   `gorm:"not null;default:true" json:"is_active"`
	Timestamps
}

package models

type Project struct {
	ProjectID    int    `gorm:"primaryKey;autoIncrement" json:"project_id"`
	ProductID    int    `gorm:"not null;index" json:"product_id"`
	ProjectCode  string `gorm:"not null" json:"project_code"`
	ProjectName  string `gorm:"not null" json:"project_name"`
	Description  string `gorm:"type:text" json:"description,omitempty"`
	DisplayOrder int    `gorm:"not null;default:0" json:"display_order"`
	IsActive     bool   `gorm:"not null;default:true" json:"is_active"`
	Timestamps
}

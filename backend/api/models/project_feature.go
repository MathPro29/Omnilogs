package models

type ProjectFeature struct {
	CategoryID   int     `gorm:"primaryKey;autoIncrement" json:"category_id"`
	ProductID    int     `gorm:"not null;uniqueIndex:uq_project_feature,priority:1;index:idx_feature_path,priority:1" json:"product_id"`
	ProjectID    int     `gorm:"not null;uniqueIndex:uq_project_feature,priority:2;index:idx_feature_path,priority:2;index:idx_project_parent,priority:1" json:"project_id"`
	ParentID     *int    `gorm:"index:idx_project_parent,priority:2" json:"parent_id,omitempty"`
	CategoryType *string `json:"category_type,omitempty"`
	CategoryCode string  `gorm:"not null;uniqueIndex:uq_project_feature,priority:3" json:"category_code"`
	CategoryName string  `gorm:"not null" json:"category_name"`
	FullPath     *string `gorm:"index:idx_feature_path,priority:3" json:"full_path,omitempty"`
	PathIDs      *string `json:"path_ids,omitempty"`
	Level        int     `gorm:"not null;default:1" json:"level"`
	IsActive     bool    `gorm:"not null;default:true" json:"is_active"`
	Timestamps
}

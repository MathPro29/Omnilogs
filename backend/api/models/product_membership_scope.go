package models

type ProductMembershipScope struct {
	ScopeID      int    `gorm:"primaryKey;autoIncrement" json:"scope_id"`
	MembershipID int    `gorm:"not null;uniqueIndex:uq_membership_scope,priority:1" json:"membership_id"`
	ProductID    int    `gorm:"not null;uniqueIndex:uq_membership_scope,priority:2;index:idx_scope_location,priority:1" json:"product_id"`
	ProjectID    *int   `gorm:"uniqueIndex:uq_membership_scope,priority:3;index:idx_scope_location,priority:2" json:"project_id,omitempty"`
	CategoryID   *int   `gorm:"uniqueIndex:uq_membership_scope,priority:4;index:idx_scope_location,priority:3" json:"category_id,omitempty"`
	ScopeLevel   string `gorm:"not null" json:"scope_level"`
	IsActive     bool   `gorm:"not null;default:true" json:"is_active"`
	Timestamps
}

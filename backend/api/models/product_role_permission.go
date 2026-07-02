package models

type ProductRolePermission struct {
	ProductRolePermissionID int    `gorm:"primaryKey;autoIncrement" json:"product_role_permission_id"`
	RoleID                 int    `gorm:"not null;uniqueIndex:uq_role_permission,priority:1;index:idx_role_permission_role,priority:1" json:"role_id"`
	ResourceType           string `gorm:"not null;uniqueIndex:uq_role_permission,priority:2;index:idx_role_permission_resource_action,priority:1" json:"resource_type"`
	Action                 string `gorm:"not null;uniqueIndex:uq_role_permission,priority:3;index:idx_role_permission_resource_action,priority:2" json:"action"`
	Timestamps
}

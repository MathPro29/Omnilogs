package models

type UserPermission struct {
	UserPermissionID int                    `gorm:"primaryKey;autoIncrement" json:"user_permission_id"`
	UserID           int                    `gorm:"not null;uniqueIndex:uq_user_permission,priority:1" json:"user_id"`
	ProductID        int                    `gorm:"not null;uniqueIndex:uq_user_permission,priority:2" json:"product_id"`
	EnvironmentID    int                    `gorm:"not null;uniqueIndex:uq_user_permission,priority:3" json:"environment_id"`
	Permissions      map[string]interface{} `gorm:"serializer:json;type:jsonb" json:"permissions"`
	Timestamps
}

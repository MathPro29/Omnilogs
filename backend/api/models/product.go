package models

type Product struct {
	ProductID   int    `gorm:"primaryKey;autoIncrement" json:"product_id"`
	ProductName string `gorm:"not null" json:"product_name"`
	ProductCode string `gorm:"not null;uniqueIndex" json:"product_code"`
	IsActive    bool   `gorm:"not null;default:true" json:"is_active"`
	Timestamps
	ProductEnvironments []ProductEnvironment `gorm:"foreignKey:ProductID;references:ProductID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"environments"`
}

package main

import (
	"fmt"
	"log"

	"omnilogs-api/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=admin dbname=omnilogs port=5499 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	var roles []models.ProductRole
	db.Find(&roles)

	fmt.Printf("Found %d roles\n", len(roles))
	for _, r := range roles {
		if !r.IsActive {
			db.Model(&r).Update("is_active", true)
			fmt.Printf("Activated RoleID: %d, RoleCode: %s\n", r.RoleID, r.RoleCode)
		}
	}
}

package main

import (
	"flag"
	"fmt"
	"log"

	"omnilogs-api/configs"
	"omnilogs-api/models"
	"omnilogs-api/utils"

	"gorm.io/gorm"
)

func main() {
	emailFlag := flag.String("email", "godmode@godmail.com", "Email for the GOD user")
	passwordFlag := flag.String("password", "OhMyGod999@!", "Password for the GOD user")
	flag.Parse()

	env := configs.LoadEnv()
	if err := env.Validate(); err != nil {
		log.Fatalf("invalid environment configuration: %v", err)
	}

	db, err := configs.ConnectDB(env)
	if err != nil {
		log.Fatalf("connect database failed: %v", err)
	}

	if err := seedGodUser(db, *emailFlag, *passwordFlag); err != nil {
		log.Fatalf("failed to seed GOD user: %v", err)
	}
}

func seedGodUser(db *gorm.DB, email, password string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 1. Verify that the Platform Role "god" (ID: 1) exists.
		var godRole models.PlatformRole
		if err := tx.Where("platform_role_id = ? OR role_code = ?", 1, "god").First(&godRole).Error; err != nil {
			return fmt.Errorf("platform role 'god' not found (please run migrations first): %w", err)
		}

		// 2. Hash the password
		hash, err := utils.HashPassword(password)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		// 3. Find or Create User
		var user models.User
		if err := tx.Where("email = ?", email).Limit(1).Find(&user).Error; err != nil {
			return err
		}
		if user.UserID == 0 {
			// Create new user
			user = models.User{
				Email:        email,
				PasswordHash: hash,
				IsActive:     true,
			}
			// Set default name
			firstName := "God"
			lastName := "Administrator"
			user.FirstName = &firstName
			user.LastName = &lastName

			if err := tx.Create(&user).Error; err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
			log.Printf("created new GOD user: %s", email)
		} else {
			// Update password hash and make active
			user.PasswordHash = hash
			user.IsActive = true
			if err := tx.Save(&user).Error; err != nil {
				return fmt.Errorf("failed to update user: %w", err)
			}
			log.Printf("updated password for existing GOD user: %s", email)
		}

		// 4. Find or Create Platform Membership (Linking User to GOD Role)
		var membership models.PlatformMembership
		if err := tx.Where("user_id = ?", user.UserID).Limit(1).Find(&membership).Error; err != nil {
			return err
		}
		if membership.PlatformMembershipID == 0 {
			membership = models.PlatformMembership{
				UserID:         user.UserID,
				PlatformRoleID: godRole.PlatformRoleID,
				IsActive:       true,
			}
			if err := tx.Create(&membership).Error; err != nil {
				return fmt.Errorf("failed to create platform membership: %w", err)
			}
			log.Printf("assigned GOD platform role to user: %s", email)
		} else {
			// Update role to GOD if not already set, and set active
			membership.PlatformRoleID = godRole.PlatformRoleID
			membership.IsActive = true
			if err := tx.Save(&membership).Error; err != nil {
				return fmt.Errorf("failed to update platform membership: %w", err)
			}
			log.Printf("confirmed GOD platform role assignment for user: %s", email)
		}

		return nil
	})
}

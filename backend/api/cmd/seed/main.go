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

		// Helper function to seed a user
		seedUser := func(usrEmail, usrUsername, usrPassword, usrFirstName, usrLastName string) error {
			hash, err := utils.HashPassword(usrPassword)
			if err != nil {
				return fmt.Errorf("failed to hash password: %w", err)
			}

			var user models.User
			if err := tx.Where("email = ? OR username = ?", usrEmail, usrUsername).Limit(1).Find(&user).Error; err != nil {
				return err
			}

			if user.UserID == 0 {
				user = models.User{
					Email:        usrEmail,
					Username:     &usrUsername,
					PasswordHash: hash,
					IsActive:     true,
					FirstName:    &usrFirstName,
					LastName:     &usrLastName,
				}
				if err := tx.Create(&user).Error; err != nil {
					return fmt.Errorf("failed to create user %s: %w", usrEmail, err)
				}
				log.Printf("created new user: %s", usrEmail)
			} else {
				user.Email = usrEmail
				user.Username = &usrUsername
				user.PasswordHash = hash
				user.IsActive = true
				user.FirstName = &usrFirstName
				user.LastName = &usrLastName
				if err := tx.Save(&user).Error; err != nil {
					return fmt.Errorf("failed to update user %s: %w", usrEmail, err)
				}
				log.Printf("updated password for existing user: %s", usrEmail)
			}

			// Find or Create Platform Membership
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
				log.Printf("assigned GOD platform role to user: %s", usrEmail)
			} else {
				membership.PlatformRoleID = godRole.PlatformRoleID
				membership.IsActive = true
				if err := tx.Save(&membership).Error; err != nil {
					return fmt.Errorf("failed to update platform membership: %w", err)
				}
				log.Printf("confirmed GOD platform role assignment for user: %s", usrEmail)
			}
			return nil
		}

		// Seed godmode user
		if err := seedUser(email, "godmode", password, "God", "Administrator"); err != nil {
			return err
		}

		// Seed admin user
		if err := seedUser("admin@company.com", "admin", "admin123", "System", "Administrator"); err != nil {
			return err
		}

		return nil
	})
}

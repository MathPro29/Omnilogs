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

	if err := seedLogRoutes(db); err != nil {
		log.Printf("warning: seed log routes: %v", err)
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

func seedLogRoutes(db *gorm.DB) error {
	var products []models.Product
	if err := db.Find(&products).Error; err != nil {
		return err
	}

	for _, p := range products {
		var envs []models.ProductEnvironment
		db.Where("product_id = ?", p.ProductID).Find(&envs)
		if len(envs) == 0 {
			continue
		}

		var projects []models.Project
		db.Where("product_id = ?", p.ProductID).Find(&projects)
		if len(projects) == 0 {
			continue
		}

		firstProject := projects[0]
		var categories []models.ProjectFeature
		db.Where("product_id = ? AND project_id = ?", p.ProductID, firstProject.ProjectID).Find(&categories)
		var firstCategoryID *int
		if len(categories) > 0 {
			firstCategoryID = &categories[0].CategoryID
		}

		for _, env := range envs {
			routeKeys := []string{
				"testproduct.event",
				"testproduct.event.created",
				fmt.Sprintf("%s.event", p.ProductCode),
				fmt.Sprintf("%s.%s", p.ProductCode, firstProject.ProjectCode),
			}

			for idx, key := range routeKeys {
				var existing models.LogRoute
				err := db.Where("product_id = ? AND environment_id = ? AND LOWER(BTRIM(route_key)) = LOWER(BTRIM(?))", p.ProductID, env.EnvironmentID, key).First(&existing).Error
				if err != nil {
					newRoute := models.LogRoute{
						ProductID:     p.ProductID,
						EnvironmentID: env.EnvironmentID,
						RouteKey:      key,
						ProjectID:     firstProject.ProjectID,
						CategoryID:    firstCategoryID,
						Priority:      10 - idx,
						IsActive:      true,
					}
					if err := db.Create(&newRoute).Error; err != nil {
						log.Printf("failed to create log route for product %d, route_key %s: %v", p.ProductID, key, err)
						continue
					}
					log.Printf("[SEED ROUTE] Product=%s(%d) | Env=%s(%d) | RouteKey=%s -> ProjectID=%d, CatID=%v",
						p.ProductName, p.ProductID, env.EnvironmentName, env.EnvironmentID, key, firstProject.ProjectID, firstCategoryID)
				}
			}
		}
	}
	return nil
}

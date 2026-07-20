package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"omnilogs-api/configs"
	"omnilogs-api/migrate"
	"omnilogs-api/models"
	"omnilogs-api/utils"

	"gorm.io/gorm"
)

func main() {
	// Set dotenv path to root of backend api
	os.Setenv("ENV_PATH", "../.env")

	env := configs.LoadEnv()
	if err := env.Validate(); err != nil {
		log.Fatalf("invalid environment configuration: %v", err)
	}

	// 1. Connect to Postgres
	db, err := configs.ConnectDB(env)
	if err != nil {
		log.Fatalf("connect database failed: %v", err)
	}

	// 2. Connect to Elasticsearch
	esClient, err := configs.ConnectElasticsearch(env)
	if err != nil {
		log.Fatalf("connect elasticsearch failed: %v", err)
	}

	ctx := context.Background()

	// 3. Clear Postgres Tables
	fmt.Println("--- Dropping and Recreating Database Tables ---")

	// We can drop all tables by running DROP SCHEMA public CASCADE; and CREATE SCHEMA public;
	// This is the cleanest way to wipe the DB in Postgres.
	tx := db.Begin()
	if err := tx.Exec("DROP SCHEMA public CASCADE").Error; err != nil {
		tx.Rollback()
		log.Fatalf("failed to drop public schema: %v", err)
	}
	if err := tx.Exec("CREATE SCHEMA public").Error; err != nil {
		tx.Rollback()
		log.Fatalf("failed to create public schema: %v", err)
	}
	if err := tx.Exec("GRANT ALL ON SCHEMA public TO postgres").Error; err != nil {
		// Just in case postgres user doesn't own it
		tx.Exec("GRANT ALL ON SCHEMA public TO public")
	}
	tx.Commit()
	fmt.Println("Public schema dropped and recreated successfully.")

	// 4. Run Migrations
	fmt.Println("\n--- Running Migrations ---")
	migrate.Migrate(db, env)
	fmt.Println("Migrations completed.")

	// 5. Seed GOD and Admin users
	fmt.Println("\n--- Seeding Initial Users ---")
	if err := seedGodUser(db, "godmode@godmail.com", "OhMyGod999@!"); err != nil {
		log.Fatalf("failed to seed GOD user: %v", err)
	}
	fmt.Println("Seeding completed.")

	// 6. Clear Elasticsearch indices
	fmt.Println("\n--- Clearing Elasticsearch Indices ---")
	targets := []string{"omnilogs-*"}
	fmt.Printf("Deleting Elasticsearch indices matching: %s\n", strings.Join(targets, ", "))

	// Enable wildcard deletes temporarily
	disableWildcardCheckUrl := fmt.Sprintf("%s/_cluster/settings", env.ElasticURL)
	disableReq, err := http.NewRequest("PUT", disableWildcardCheckUrl, strings.NewReader(`{"transient":{"action.destructive_requires_name":false}}`))
	if err == nil {
		disableReq.Header.Set("Content-Type", "application/json")
		if disableResp, err := http.DefaultClient.Do(disableReq); err == nil {
			disableResp.Body.Close()
		}
	}

	res, err := esClient.Indices.Delete(
		targets,
		esClient.Indices.Delete.WithContext(ctx),
	)

	// Re-enable wildcard protection
	enableReq, err := http.NewRequest("PUT", disableWildcardCheckUrl, strings.NewReader(`{"transient":{"action.destructive_requires_name":true}}`))
	if err == nil {
		enableReq.Header.Set("Content-Type", "application/json")
		if enableResp, err := http.DefaultClient.Do(enableReq); err == nil {
			enableResp.Body.Close()
		}
	}

	if err != nil {
		fmt.Printf("Failed to delete indices: %v\n", err)
	} else {
		defer res.Body.Close()
		if res.IsError() {
			if res.StatusCode == 404 {
				fmt.Println("No matching indices found in Elasticsearch to delete.")
			} else {
				fmt.Printf("Elasticsearch delete returned error (Status %d): %s\n", res.StatusCode, res.String())
			}
		} else {
			fmt.Println("Successfully deleted indices from Elasticsearch.")
		}
	}

	fmt.Println("\nDatabase Reset Successfully Completed!")
	fmt.Println("Seed users available:")
	fmt.Println("1. GOD: godmode@godmail.com / OhMyGod999@!")
	fmt.Println("2. ADMIN: admin@company.com / admin123")
}

func seedGodUser(db *gorm.DB, email, password string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var godRole models.PlatformRole
		if err := tx.Where("platform_role_id = ? OR role_code = ?", 1, "god").First(&godRole).Error; err != nil {
			return fmt.Errorf("platform role 'god' not found (please run migrations first): %w", err)
		}

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

		if err := seedUser(email, "godmode", password, "God", "Administrator"); err != nil {
			return err
		}

		if err := seedUser("admin@company.com", "admin", "admin123", "System", "Administrator"); err != nil {
			return err
		}

		return nil
	})
}

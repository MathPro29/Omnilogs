package migrate

import (
	"log"
	"omnilogs-api/configs"
	"omnilogs-api/models"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB, env *configs.Env) {
	if err := repairOrphanProductEnvironments(db); err != nil {
		log.Fatalf("Product environment data repair failed: %v", err)
	}

	// Older schemas allowed more than one platform role per user through a
	// composite (user_id, platform_role_id) unique index. Normalize those rows
	// before AutoMigrate creates the user-only unique index.
	if err := migratePlatformMembershipUniqueness(db); err != nil {
		log.Fatalf("Platform membership uniqueness migration failed: %v", err)
	}

	err := db.AutoMigrate(
		&models.PlatformRole{},
		&models.PlatformMembership{},
		&models.User{},
		&models.Product{},
		&models.ProductEnvironment{},
		&models.ProductAPIKey{},
		&models.ProductRole{},
		&models.ProductMembership{},
		&models.ProductMembershipScope{},
		&models.Project{},
		&models.ProjectFeature{},
		&models.RoleTemplate{},
		&models.UserRolePermissionRule{},
		&models.ElasticIndexPolicy{},
		&models.LogArchive{},
		&models.LogFailure{},
		&models.LogFieldDefinition{},
		&models.LogFieldEnumOption{},
		&models.LogFieldValueSource{},
		&models.LogIndexRef{},
		&models.LogIngestionPolicy{},
		&models.LogMaskingRule{},
		&models.LogObjectStorageRef{},
		&models.LogQueueBatch{},
		&models.LogQueueItem{},
		&models.LogSensitiveFieldSecret{},
		&models.LogSource{},
		&models.SensitiveLogAccessHistory{},
		&models.SensitiveLogAccessRequest{},
		&models.SystemAuditLog{},
	)
	if err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	seedPlatformRoles(db)
}

func repairOrphanProductEnvironments(db *gorm.DB) error {
	if !db.Migrator().HasTable(&models.Product{}) || !db.Migrator().HasTable(&models.ProductEnvironment{}) {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// Preserve invalid legacy rows before removing them from the live table.
		// They can be inspected and restored after assigning a valid product_id.
		if err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS migration_orphan_product_environments AS
			SELECT pe.*, NOW()::timestamptz AS archived_at
			FROM product_environments pe
			WHERE FALSE
		`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
			INSERT INTO migration_orphan_product_environments
			SELECT pe.*, NOW()::timestamptz AS archived_at
			FROM product_environments pe
			LEFT JOIN products p ON p.product_id = pe.product_id
			WHERE p.product_id IS NULL
			  AND NOT EXISTS (
				SELECT 1
				FROM migration_orphan_product_environments archived
				WHERE archived.environment_id = pe.environment_id
			  )
		`).Error; err != nil {
			return err
		}

		return tx.Exec(`
			DELETE FROM product_environments pe
			WHERE NOT EXISTS (
				SELECT 1 FROM products p WHERE p.product_id = pe.product_id
			)
		`).Error
	})
}

func migratePlatformMembershipUniqueness(db *gorm.DB) error {
	if !db.Migrator().HasTable(&models.PlatformMembership{}) {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// Keep one row per user, preferring an active membership and then the
		// most recently changed row.
		if err := tx.Exec(`
			DELETE FROM platform_memberships
			WHERE platform_membership_id IN (
				SELECT platform_membership_id
				FROM (
					SELECT platform_membership_id,
						ROW_NUMBER() OVER (
							PARTITION BY user_id
							ORDER BY is_active DESC, updated_at DESC NULLS LAST,
								created_at DESC NULLS LAST, platform_membership_id DESC
						) AS row_number
					FROM platform_memberships
				) ranked_memberships
				WHERE row_number > 1
			)
		`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`DROP INDEX IF EXISTS uq_platform_membership`).Error; err != nil {
			return err
		}

		return tx.Exec(`
			CREATE UNIQUE INDEX IF NOT EXISTS uq_platform_membership_user
			ON platform_memberships (user_id)
		`).Error
	})
}

func seedPlatformRoles(db *gorm.DB) {
	roles := []models.PlatformRole{
		{
			PlatformRoleID: 1,
			RoleCode:       "god",
			RoleName:       "GOD",
			Permissions:    []byte(`{"all": true}`),
			IsSystemRole:   true,
			IsActive:       true,
		},
		{
			PlatformRoleID: 2,
			RoleCode:       "owner",
			RoleName:       "Owner",
			Permissions:    []byte(`{"own_product_access": true}`),
			IsSystemRole:   true,
			IsActive:       true,
		},
		{
			PlatformRoleID: 3,
			RoleCode:       "superadmin",
			RoleName:       "Superadmin",
			Permissions:    []byte(`{"assign_role_only": true}`),
			IsSystemRole:   true,
			IsActive:       true,
		},
		{
			PlatformRoleID: 4,
			RoleCode:       "user",
			RoleName:       "User",
			Permissions:    []byte(`{}`),
			IsSystemRole:   true,
			IsActive:       true,
		},
	}

	for _, r := range roles {
		var existing models.PlatformRole
		if err := db.Where("platform_role_id = ?", r.PlatformRoleID).First(&existing).Error; err != nil {
			// If not found, insert
			db.Create(&r)
		}
	}

	// Reset sequence in PostgreSQL
	db.Exec("SELECT setval(pg_get_serial_sequence('platform_roles', 'platform_role_id'), COALESCE((SELECT MAX(platform_role_id) FROM platform_roles), 1), true)")
}

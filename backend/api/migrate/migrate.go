package migrate

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"omnilogs-api/configs"
	"omnilogs-api/models"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB, env *configs.Env) {
	if err := migrateCustomFieldSchema(db); err != nil {
		log.Fatalf("Custom field schema migration failed: %v", err)
	}

	if err := repairOrphanProductEnvironments(db); err != nil {
		log.Fatalf("Product environment data repair failed: %v", err)
	}

	// Older schemas allowed more than one platform role per user through a
	// composite (user_id, platform_role_id) unique index. Normalize those rows
	// before AutoMigrate creates the user-only unique index.
	if err := migratePlatformMembershipUniqueness(db); err != nil {
		log.Fatalf("Platform membership uniqueness migration failed: %v", err)
	}

	if err := migrateProductMembershipUniqueness(db); err != nil {
		log.Fatalf("Product membership uniqueness migration failed: %v", err)
	}

	if err := migrateProductUniqueness(db); err != nil {
		log.Fatalf("Product uniqueness migration failed: %v", err)
	}

	if err := prepareLogArchiveBackupColumns(db); err != nil {
		log.Fatalf("Log archive backup columns migration failed: %v", err)
	}

	err := db.AutoMigrate(
		&models.PlatformRole{},
		&models.PlatformMembership{},
		&models.AuthSession{},
		&models.PasswordResetToken{},
		&models.User{},
		&models.Product{},
		&models.ProductEnvironment{},
		&models.ProductAPIKey{},
		&models.ProductRole{},
		&models.ProductRolePermission{},
		&models.ProductMembership{},
		&models.ProductMembershipScope{},
		&models.Project{},
		&models.ProjectFeature{},
		&models.RoleTemplate{},
		&models.UserRolePermissionRule{},
		&models.UserPermission{},
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
		&models.LogRoute{},
		&models.LogRoutingRule{},
		&models.AuditSecret{},
		&models.AuditSecretAccessRequest{},
		&models.AuditSecretAccessHistory{},
		&models.SensitiveLogAccessHistory{},
		&models.SensitiveLogAccessRequest{},
		&models.SystemAuditLog{},
		&models.LogSchemaVersion{},
		&models.LogFieldTemplate{},
		&models.LogRetentionPolicy{},
		&models.LogArchiveJob{},
		&models.LogArchiveDownload{},
	)
	if err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	if err := backfillRetentionCalendarFields(db); err != nil {
		log.Fatalf("Retention calendar field migration failed: %v", err)
	}
	if err := ensureRetentionPolicyUniqueness(db); err != nil {
		log.Fatalf("Retention policy uniqueness migration failed: %v", err)
	}

	if err := ensureProductAccessUniqueness(db); err != nil {
		log.Fatalf("Product access uniqueness migration failed: %v", err)
	}

	if err := backfillUnscopedProductMemberships(db); err != nil {
		log.Fatalf("Product membership scope backfill failed: %v", err)
	}

	if err := ensureProductAPIKeyColumns(db); err != nil {
		log.Fatalf("Product API key columns migration failed: %v", err)
	}

	if err := migrateEntityCodesToIDs(db); err != nil {
		log.Fatalf("Entity code-to-ID migration failed: %v", err)
	}

	if err := ensureHierarchyUniqueness(db); err != nil {
		log.Fatalf("Hierarchy uniqueness migration failed: %v", err)
	}

	if err := ensureQueryPerformanceIndexes(db); err != nil {
		log.Fatalf("Query performance index migration failed: %v", err)
	}

	if err := backfillLegacyProductRolePermissions(db); err != nil {
		log.Fatalf("Product role permission backfill failed: %v", err)
	}

	seedPlatformRoles(db)
}

// prepareLogArchiveBackupColumns makes the additive archive metadata migration
// safe for databases that already contain archives. PostgreSQL cannot add a
// NOT NULL column to a populated table without a value for existing rows.
func prepareLogArchiveBackupColumns(db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" || !db.Migrator().HasTable(&models.LogArchive{}) {
		return nil
	}
	statements := []string{
		`ALTER TABLE log_archives ADD COLUMN IF NOT EXISTS backup_type VARCHAR(20)`,
		`ALTER TABLE log_archives ADD COLUMN IF NOT EXISTS backup_tag VARCHAR(180)`,
		`ALTER TABLE log_archives ADD COLUMN IF NOT EXISTS coverage_key VARCHAR(255)`,
		`ALTER TABLE log_archives ADD COLUMN IF NOT EXISTS object_key TEXT`,
		`UPDATE log_archives SET backup_type = 'POLICY' WHERE backup_type IS NULL`,
		`UPDATE log_archives SET backup_tag = '' WHERE backup_tag IS NULL`,
		`ALTER TABLE log_archives ALTER COLUMN backup_type SET DEFAULT 'POLICY'`,
		`ALTER TABLE log_archives ALTER COLUMN backup_tag SET DEFAULT ''`,
		`ALTER TABLE log_archives ALTER COLUMN backup_type SET NOT NULL`,
		`ALTER TABLE log_archives ALTER COLUMN backup_tag SET NOT NULL`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func ensureRetentionPolicyUniqueness(db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" || !db.Migrator().HasTable(&models.LogRetentionPolicy{}) {
		return nil
	}
	return db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_retention_policy_product_environment_active
		ON log_retention_policies (product_id, environment_id)
		WHERE environment_id IS NOT NULL AND is_active = TRUE
	`).Error
}

func backfillRetentionCalendarFields(db *gorm.DB) error {
	if !db.Migrator().HasTable(&models.LogRetentionPolicy{}) {
		return nil
	}
	if err := db.Exec(`
		UPDATE log_retention_policies
		SET active_retention_value = CASE WHEN retention_value > 0 THEN retention_value ELSE 1 END,
		    active_retention_unit = CASE UPPER(retention_unit)
		        WHEN 'DAYS' THEN 'DAY' WHEN 'DAY' THEN 'DAY'
		        WHEN 'WEEKS' THEN 'WEEK' WHEN 'WEEK' THEN 'WEEK'
		        WHEN 'MONTHS' THEN 'MONTH' WHEN 'MONTH' THEN 'MONTH'
		        WHEN 'YEARS' THEN 'YEAR' WHEN 'YEAR' THEN 'YEAR'
		        ELSE 'DAY' END
		WHERE active_retention_value = 1 AND active_retention_unit = 'DAY'
	`).Error; err != nil {
		return err
	}
	if !db.Migrator().HasColumn(&models.LogRetentionPolicy{}, "effective_from") {
		return nil
	}
	return db.Exec(`
		UPDATE log_retention_policies
		SET effective_from = COALESCE(created_at, NOW())
		WHERE apply_to_existing_logs = FALSE AND effective_from IS NULL
	`).Error
}

type legacyProductRolePermissionRow struct {
	RoleID      int
	Permissions []byte
}

func backfillLegacyProductRolePermissions(db *gorm.DB) error {
	if !db.Migrator().HasTable(&models.ProductRole{}) || !db.Migrator().HasTable(&models.ProductRolePermission{}) {
		return nil
	}
	if !db.Migrator().HasColumn(&models.ProductRole{}, "permissions") {
		return nil
	}

	var rows []legacyProductRolePermissionRow
	if err := db.Raw("SELECT role_id, permissions FROM product_roles WHERE permissions IS NOT NULL").Scan(&rows).Error; err != nil {
		return err
	}

	for _, row := range rows {
		var existing int64
		if err := db.Model(&models.ProductRolePermission{}).Where("role_id = ?", row.RoleID).Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			continue
		}

		permissions := parseLegacyPermissions(row.RoleID, row.Permissions)
		if len(permissions) == 0 {
			continue
		}
		if err := db.Create(&permissions).Error; err != nil {
			return err
		}
	}

	return nil
}

func parseLegacyPermissions(roleID int, raw []byte) []models.ProductRolePermission {
	var value map[string]any
	if json.Unmarshal(raw, &value) != nil {
		return nil
	}

	var permissions []models.ProductRolePermission
	add := func(resource, action string) {
		permissions = append(permissions, models.ProductRolePermission{
			RoleID:       roleID,
			ResourceType: strings.ToUpper(strings.TrimSpace(resource)),
			Action:       strings.ToUpper(strings.TrimSpace(action)),
		})
	}

	for key, item := range value {
		if strings.EqualFold(key, "all") {
			continue
		}

		if strings.Contains(key, ".") {
			parts := strings.SplitN(key, ".", 2)
			if len(parts) == 2 {
				if allowed, ok := item.(bool); ok && allowed {
					add(parts[0], parts[1])
				}
			}
			continue
		}

		resource := key
		switch typed := item.(type) {
		case []any:
			for _, action := range typed {
				add(resource, fmt.Sprint(action))
			}
		case map[string]any:
			for action, allowed := range typed {
				if yes, ok := allowed.(bool); ok && yes {
					add(resource, action)
				}
			}
		case bool:
			if typed {
				for _, action := range []string{"CREATE", "READ", "UPDATE", "DELETE", "GRANT", "REVOKE", "EXPORT", "VIEW_SENSITIVE"} {
					add(resource, action)
				}
			}
		}
	}

	return dedupePermissions(permissions)
}

func dedupePermissions(values []models.ProductRolePermission) []models.ProductRolePermission {
	seen := map[string]struct{}{}
	result := make([]models.ProductRolePermission, 0, len(values))
	for _, value := range values {
		key := fmt.Sprintf("%d:%s:%s", value.RoleID, value.ResourceType, value.Action)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
	}
	return result
}

func repairOrphanProductEnvironments(db *gorm.DB) error {
	if !db.Migrator().HasTable(&models.Product{}) || !db.Migrator().HasTable(&models.ProductEnvironment{}) {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS migration_orphan_product_environments (
				environment_id INT,
				product_id INT,
				environment_code VARCHAR(255),
				environment_name VARCHAR(255),
				archived_at TIMESTAMPTZ
			)
		`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
			INSERT INTO migration_orphan_product_environments (environment_id, product_id, environment_code, environment_name, archived_at)
			SELECT pe.environment_id, pe.product_id, pe.environment_code, pe.environment_name, NOW()::timestamptz AS archived_at
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

func migrateProductMembershipUniqueness(db *gorm.DB) error {
	if !db.Migrator().HasTable(&models.ProductMembership{}) {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if tx.Dialector.Name() != "postgres" {
			return nil
		}
		if tx.Migrator().HasTable(&models.ProductMembershipScope{}) {
			if err := tx.Exec(`
			WITH ranked AS (
				SELECT membership_id, FIRST_VALUE(membership_id) OVER (PARTITION BY user_id, product_id ORDER BY is_active DESC, updated_at DESC NULLS LAST, membership_id DESC) keep_id,
				ROW_NUMBER() OVER (PARTITION BY user_id, product_id ORDER BY is_active DESC, updated_at DESC NULLS LAST, membership_id DESC) rn
				FROM product_memberships
			), duplicates AS (SELECT membership_id, keep_id FROM ranked WHERE rn > 1)
			DELETE FROM product_membership_scopes s USING duplicates d
			WHERE s.membership_id = d.membership_id AND EXISTS (
				SELECT 1 FROM product_membership_scopes k WHERE k.membership_id=d.keep_id AND k.product_id=s.product_id
				AND COALESCE(k.project_id,0)=COALESCE(s.project_id,0) AND COALESCE(k.category_id,0)=COALESCE(s.category_id,0) AND k.scope_level=s.scope_level)
		`).Error; err != nil {
				return err
			}
			if err := tx.Exec(`WITH ranked AS (SELECT membership_id, FIRST_VALUE(membership_id) OVER (PARTITION BY user_id, product_id ORDER BY is_active DESC, updated_at DESC NULLS LAST, membership_id DESC) keep_id, ROW_NUMBER() OVER (PARTITION BY user_id, product_id ORDER BY is_active DESC, updated_at DESC NULLS LAST, membership_id DESC) rn FROM product_memberships), duplicates AS (SELECT membership_id, keep_id FROM ranked WHERE rn > 1) UPDATE product_membership_scopes s SET membership_id=d.keep_id FROM duplicates d WHERE s.membership_id=d.membership_id`).Error; err != nil {
				return err
			}
		}
		if err := tx.Exec(`DELETE FROM product_memberships WHERE membership_id IN (SELECT membership_id FROM (SELECT membership_id, ROW_NUMBER() OVER (PARTITION BY user_id, product_id ORDER BY is_active DESC, updated_at DESC NULLS LAST, membership_id DESC) rn FROM product_memberships) x WHERE rn > 1)`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`DROP INDEX IF EXISTS uq_product_membership`).Error; err != nil {
			return err
		}
		return tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uq_product_membership ON product_memberships (user_id, product_id)`).Error
	})
}

func migrateProductUniqueness(db *gorm.DB) error {
	if !db.Migrator().HasTable(&models.Product{}) {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if tx.Dialector.Name() != "postgres" {
			return nil
		}
		_ = tx.Exec(`ALTER TABLE products DROP CONSTRAINT IF EXISTS uni_products_product_code`).Error
		_ = tx.Exec(`ALTER TABLE products DROP CONSTRAINT IF EXISTS products_product_code_key`).Error
		_ = tx.Exec(`DROP INDEX IF EXISTS uni_products_product_code`).Error
		_ = tx.Exec(`DROP INDEX IF EXISTS idx_products_product_code`).Error
		_ = tx.Exec(`DROP INDEX IF EXISTS idx_products_product_code_active`).Error

		return tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_products_product_code_active ON products (product_code) WHERE deleted_at IS NULL`).Error
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

func ensureProductAccessUniqueness(db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if tx.Migrator().HasTable(&models.ProductMembershipScope{}) {
			if err := tx.Exec(`DELETE FROM product_membership_scopes WHERE scope_id IN (SELECT scope_id FROM (SELECT scope_id, ROW_NUMBER() OVER (PARTITION BY membership_id, product_id, COALESCE(project_id,0), COALESCE(category_id,0), scope_level ORDER BY is_active DESC, updated_at DESC NULLS LAST, scope_id DESC) rn FROM product_membership_scopes) ranked WHERE rn > 1)`).Error; err != nil {
				return err
			}
			if err := tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uq_membership_scope_effective ON product_membership_scopes (membership_id, product_id, COALESCE(project_id,0), COALESCE(category_id,0), scope_level)`).Error; err != nil {
				return err
			}
		}
		if tx.Migrator().HasTable(&models.UserRolePermissionRule{}) {
			if err := tx.Exec(`DELETE FROM user_role_permission_rules WHERE permission_rule_id IN (SELECT permission_rule_id FROM (SELECT permission_rule_id, ROW_NUMBER() OVER (PARTITION BY user_id, COALESCE(product_id,0), COALESCE(role_id,0), COALESCE(project_id,0), COALESCE(category_id,0), resource_type, action ORDER BY is_active DESC, updated_at DESC NULLS LAST, permission_rule_id DESC) rn FROM user_role_permission_rules) ranked WHERE rn > 1)`).Error; err != nil {
				return err
			}
			if err := tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uq_permission_rule_effective ON user_role_permission_rules (user_id, COALESCE(product_id,0), COALESCE(role_id,0), COALESCE(project_id,0), COALESCE(category_id,0), resource_type, action)`).Error; err != nil {
				return err
			}
		}
		return nil
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
		{
			PlatformRoleID: 5,
			RoleCode:       "admin",
			RoleName:       "Admin",
			Permissions:    []byte(`{"default_menu_access": true}`),
			IsSystemRole:   true,
			IsActive:       true,
		},
	}

	for _, r := range roles {
		var count int64
		if err := db.Model(&models.PlatformRole{}).Where("platform_role_id = ?", r.PlatformRoleID).Count(&count).Error; err != nil {
			log.Printf("failed to check platform role: %v", err)
			continue
		}
		if count == 0 {
			if err := db.Create(&r).Error; err != nil {
				log.Printf("failed to seed platform role: %v", err)
			}
		}
	}

	// Reset sequence in PostgreSQL
	db.Exec("SELECT setval(pg_get_serial_sequence('platform_roles', 'platform_role_id'), COALESCE((SELECT MAX(platform_role_id) FROM platform_roles), 1), true)")
}

func migrateEntityCodesToIDs(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Move through a collision-free namespace first. Existing human codes may
		// already equal another row's numeric ID while unique indexes are active.
		updates := []string{
			`UPDATE products SET product_code = '__ID_MIGRATION__' || CAST(product_id AS TEXT)`,
			`UPDATE projects SET project_code = '__ID_MIGRATION__' || CAST(project_id AS TEXT)`,
			`UPDATE project_features SET category_code = '__ID_MIGRATION__' || CAST(category_id AS TEXT)`,
			`UPDATE products SET product_code = CAST(product_id AS TEXT)`,
			`UPDATE projects SET project_code = CAST(project_id AS TEXT)`,
			`UPDATE project_features SET category_code = CAST(category_id AS TEXT)`,
		}
		for _, statement := range updates {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func migrateCustomFieldSchema(db *gorm.DB) error {
	if !db.Migrator().HasTable(&models.LogFieldDefinition{}) {
		return nil
	}
	if !db.Migrator().HasColumn(&models.LogFieldDefinition{}, "config_json") {
		return nil // AutoMigrate hasn't run yet, we'll migrate next time or let it be
	}
	// Note: Proper migration logic for extracting flat columns to JSONB can be added here
	return nil
}

func ensureProductAPIKeyColumns(db *gorm.DB) error {
	if !db.Migrator().HasTable(&models.ProductAPIKey{}) {
		return nil
	}
	if db.Dialector.Name() == "postgres" {
		return db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec(`ALTER TABLE product_api_keys ADD COLUMN IF NOT EXISTS source_id INT`).Error; err != nil {
				return err
			}
			if err := tx.Exec(`ALTER TABLE product_api_keys ADD COLUMN IF NOT EXISTS default_project_id INT`).Error; err != nil {
				return err
			}
			if err := tx.Exec(`ALTER TABLE product_api_keys ADD COLUMN IF NOT EXISTS default_category_id INT`).Error; err != nil {
				return err
			}
			return nil
		})
	}
	return nil
}

// backfillUnscopedProductMemberships repairs memberships created by older
// single-member flows, which omitted the scope required for product access.
// Memberships that already have a narrower project/category scope are left
// unchanged.
func backfillUnscopedProductMemberships(db *gorm.DB) error {
	if !db.Migrator().HasTable(&models.ProductMembership{}) || !db.Migrator().HasTable(&models.ProductMembershipScope{}) {
		return nil
	}
	return db.Exec(`
		INSERT INTO product_membership_scopes (membership_id, product_id, scope_level, is_active, created_at, updated_at)
		SELECT pm.membership_id, pm.product_id, 'PRODUCT', TRUE, NOW(), NOW()
		FROM product_memberships pm
		WHERE NOT EXISTS (
			SELECT 1
			FROM product_membership_scopes pms
			WHERE pms.membership_id = pm.membership_id
				AND pms.product_id = pm.product_id
		)
		ON CONFLICT DO NOTHING
	`).Error
}

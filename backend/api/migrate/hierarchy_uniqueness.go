package migrate

import (
	"fmt"

	"gorm.io/gorm"
)

// ensureHierarchyUniqueness replaces legacy case-sensitive/global constraints
// with indexes matching the actual ownership hierarchy. Existing conflicts are
// reported instead of being silently renamed or deleted.
func ensureHierarchyUniqueness(db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" {
		return nil
	}
	if err := createUniquenessConflictReport(db); err != nil {
		return err
	}
	if err := recordHierarchyConflicts(db); err != nil {
		return err
	}

	var conflicts int64
	if err := db.Table("migration_uniqueness_conflicts").
		Where("migration_name = ?", hierarchyUniquenessMigration).
		Count(&conflicts).Error; err != nil {
		return err
	}
	if conflicts > 0 {
		return fmt.Errorf("%d hierarchy uniqueness conflicts recorded in migration_uniqueness_conflicts", conflicts)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		statements := []string{
			`ALTER TABLE products DROP CONSTRAINT IF EXISTS uni_products_product_code`,
			`ALTER TABLE products DROP CONSTRAINT IF EXISTS products_product_code_key`,
			`DROP INDEX IF EXISTS idx_products_product_code_active`,
			`DROP INDEX IF EXISTS uq_project_code`,
			`DROP INDEX IF EXISTS uq_project_feature`,
			`DROP INDEX IF EXISTS uq_log_source`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_products_active_name_normalized ON products (LOWER(BTRIM(product_name))) WHERE deleted_at IS NULL`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_products_active_code_normalized ON products (LOWER(BTRIM(product_code))) WHERE deleted_at IS NULL`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_projects_product_name_normalized ON projects (product_id, LOWER(BTRIM(project_name)))`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_projects_product_code_normalized ON projects (product_id, LOWER(BTRIM(project_code)))`,
			`DROP INDEX IF EXISTS uq_features_project_code_normalized`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_features_parent_code_normalized ON project_features (project_id, COALESCE(parent_id, 0), LOWER(BTRIM(category_code)))`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_features_parent_name_normalized ON project_features (project_id, COALESCE(parent_id, 0), LOWER(BTRIM(category_name)))`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_log_sources_scope_code_normalized ON log_sources (COALESCE(product_id, 0), COALESCE(environment_id, 0), LOWER(BTRIM(source_code)))`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_log_routes_scope_key_normalized ON log_routes (product_id, environment_id, COALESCE(source_id, 0), LOWER(BTRIM(route_key)))`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_log_routing_rules_scope_name_normalized ON log_routing_rules (product_id, COALESCE(environment_id, 0), COALESCE(source_id, 0), LOWER(BTRIM(rule_name)))`,
		}
		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

const hierarchyUniquenessMigration = "20260727_hierarchy_uniqueness"

func createUniquenessConflictReport(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS migration_uniqueness_conflicts (
			conflict_id BIGSERIAL PRIMARY KEY,
			migration_name TEXT NOT NULL,
			entity_type TEXT NOT NULL,
			scope_key TEXT NOT NULL,
			normalized_value TEXT NOT NULL,
			record_ids JSONB NOT NULL,
			reported_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE (migration_name, entity_type, scope_key, normalized_value)
		)
	`).Error
}

func recordHierarchyConflicts(db *gorm.DB) error {
	if err := db.Exec(`DELETE FROM migration_uniqueness_conflicts WHERE migration_name = ?`, hierarchyUniquenessMigration).Error; err != nil {
		return err
	}
	queries := []string{
		`INSERT INTO migration_uniqueness_conflicts (migration_name, entity_type, scope_key, normalized_value, record_ids)
		 SELECT ?, 'PRODUCT_NAME', 'ACTIVE', LOWER(BTRIM(product_name)), JSONB_AGG(product_id ORDER BY product_id)
		 FROM products WHERE deleted_at IS NULL GROUP BY LOWER(BTRIM(product_name)) HAVING COUNT(*) > 1`,
		`INSERT INTO migration_uniqueness_conflicts (migration_name, entity_type, scope_key, normalized_value, record_ids)
		 SELECT ?, 'PRODUCT_CODE', 'ACTIVE', LOWER(BTRIM(product_code)), JSONB_AGG(product_id ORDER BY product_id)
		 FROM products WHERE deleted_at IS NULL GROUP BY LOWER(BTRIM(product_code)) HAVING COUNT(*) > 1`,
		`INSERT INTO migration_uniqueness_conflicts (migration_name, entity_type, scope_key, normalized_value, record_ids)
		 SELECT ?, 'PROJECT_NAME', product_id::TEXT, LOWER(BTRIM(project_name)), JSONB_AGG(project_id ORDER BY project_id)
		 FROM projects GROUP BY product_id, LOWER(BTRIM(project_name)) HAVING COUNT(*) > 1`,
		`INSERT INTO migration_uniqueness_conflicts (migration_name, entity_type, scope_key, normalized_value, record_ids)
		 SELECT ?, 'PROJECT_CODE', product_id::TEXT, LOWER(BTRIM(project_code)), JSONB_AGG(project_id ORDER BY project_id)
		 FROM projects GROUP BY product_id, LOWER(BTRIM(project_code)) HAVING COUNT(*) > 1`,
		`INSERT INTO migration_uniqueness_conflicts (migration_name, entity_type, scope_key, normalized_value, record_ids)
		 SELECT ?, 'FEATURE_CODE', project_id::TEXT || ':' || COALESCE(parent_id, 0)::TEXT, LOWER(BTRIM(category_code)), JSONB_AGG(category_id ORDER BY category_id)
		 FROM project_features GROUP BY project_id, COALESCE(parent_id, 0), LOWER(BTRIM(category_code)) HAVING COUNT(*) > 1`,
		`INSERT INTO migration_uniqueness_conflicts (migration_name, entity_type, scope_key, normalized_value, record_ids)
		 SELECT ?, 'FEATURE_NAME', project_id::TEXT || ':' || COALESCE(parent_id, 0)::TEXT, LOWER(BTRIM(category_name)), JSONB_AGG(category_id ORDER BY category_id)
		 FROM project_features GROUP BY project_id, COALESCE(parent_id, 0), LOWER(BTRIM(category_name)) HAVING COUNT(*) > 1`,
	}
	for _, query := range queries {
		if err := db.Exec(query, hierarchyUniquenessMigration).Error; err != nil {
			return err
		}
	}
	return nil
}

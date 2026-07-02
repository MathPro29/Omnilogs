package usecase

import (
	"context"
	"errors"
	"fmt"

	"omnilogs-api/models"
)

func (u *usecase) validateLogHierarchy(
	ctx context.Context,
	productID int,
	environmentID int,
	meta indexedLogMeta,
) error {
	// ส่วนที่ 1: Environment ต้องยังไม่ถูกลบ และต้องอยู่ใน Product ของ Batch
	var environmentCount int64
	if err := u.repo.DB().
		Model(&models.ProductEnvironment{}).
		Where(
			"environment_id = ? AND product_id = ? AND deleted_at IS NULL",
			environmentID,
			productID,
		).
		Count(&environmentCount).Error; err != nil {
		return fmt.Errorf("validate environment: %w", err)
	}
	if environmentCount == 0 {
		return errors.New("environment does not belong to product")
	}

	// ส่วนที่ 2: Category ต้องอ้างอิงผ่าน Project เสมอ จึงห้ามมี category_id เพียงค่าเดียว
	if meta.ProjectID == nil && meta.CategoryID != nil {
		return errors.New("project_id is required when category_id is provided")
	}

	// ส่วนที่ 3: Project ต้อง Active และต้องอยู่ใน Product เดียวกัน
	if meta.ProjectID != nil {
		var projectCount int64
		if err := u.repo.DB().
			Model(&models.Project{}).
			Where(
				"project_id = ? AND product_id = ? AND is_active = TRUE",
				*meta.ProjectID,
				productID,
			).
			Count(&projectCount).Error; err != nil {
			return fmt.Errorf("validate project: %w", err)
		}
		if projectCount == 0 {
			return errors.New("project does not belong to product")
		}
	}

	// ส่วนที่ 4: Category ต้อง Active และตรงกันครบทั้ง Product และ Project
	// กันการนำ Category จาก Hierarchy อื่นมาปะปนกัน
	if meta.CategoryID != nil {
		var categoryCount int64
		if err := u.repo.DB().
			Model(&models.ProjectFeature{}).
			Where(
				"category_id = ? AND project_id = ? AND product_id = ? AND is_active = TRUE",
				*meta.CategoryID,
				*meta.ProjectID,
				productID,
			).
			Count(&categoryCount).Error; err != nil {
			return fmt.Errorf("validate category: %w", err)
		}
		if categoryCount == 0 {
			return errors.New(
				"category does not belong to the specified product and project",
			)
		}
	}

	return nil
}

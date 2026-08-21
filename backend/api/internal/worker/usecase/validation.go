package usecase

import (
	"context"
	"errors"
	"fmt"

	"omnilogs-api/models"
)

func applicableFieldDefinitions(fields []models.LogFieldDefinition, meta indexedLogMeta) []models.LogFieldDefinition {
	result := make([]models.LogFieldDefinition, 0, len(fields))
	for _, field := range fields {
		if field.ProjectID != nil && (meta.ProjectID == nil || *field.ProjectID != *meta.ProjectID) {
			continue
		}
		if field.CategoryID != nil && (meta.CategoryID == nil || *field.CategoryID != *meta.CategoryID) {
			continue
		}
		result = append(result, field)
	}
	return result
}
func (u *usecase) validateLogHierarchy(
	ctx context.Context,
	productID int,
	environmentID int,
	meta indexedLogMeta,
	cache *batchCache,
) (err error) {
	var projID, catID int
	if meta.ProjectID != nil {
		projID = *meta.ProjectID
	}
	if meta.CategoryID != nil {
		catID = *meta.CategoryID
	}
	cacheKey := fmt.Sprintf("%d-%d-%d-%d", productID, environmentID, projID, catID)

	if val, ok := cache.hierarchy.Load(cacheKey); ok {
		if val == nil {
			return nil
		}
		return val.(error)
	}

	defer func() {
		cache.hierarchy.Store(cacheKey, err)
	}()
	// ส่วนที่ 1: Environment ต้องยังไม่ถูกลบ และต้องอยู่ใน Product ของ Batch
	var environmentCount int64
	err = u.repo.DB().
		Model(&models.ProductEnvironment{}).
		Where(
			"environment_id = ? AND product_id = ? AND deleted_at IS NULL",
			environmentID,
			productID,
		).
		Count(&environmentCount).Error
	if err != nil {
		err = fmt.Errorf("validate environment: %w", err)
		return
	}
	if environmentCount == 0 {
		err = errors.New("environment does not belong to product")
		return
	}

	if routingStatus(meta.RoutingStatus) == "CLASSIFIED" {
		if meta.ProjectID == nil && meta.CategoryID != nil {
			err = errors.New("category requires project scope")
			return
		}
		if meta.ProjectID != nil {
			var projectCount int64
			err = u.repo.DB().WithContext(ctx).Model(&models.Project{}).
				Where("project_id = ? AND product_id = ? AND is_active = TRUE", *meta.ProjectID, productID).
				Count(&projectCount).Error
			if err != nil {
				err = fmt.Errorf("validate project: %w", err)
				return
			}
			if projectCount != 1 {
				err = errors.New("project does not belong to product")
				return
			}
		}
		if meta.CategoryID != nil {
			var categoryCount int64
			err = u.repo.DB().WithContext(ctx).Model(&models.ProjectFeature{}).
				Where("category_id = ? AND project_id = ? AND product_id = ? AND is_active = TRUE", *meta.CategoryID, *meta.ProjectID, productID).
				Count(&categoryCount).Error
			if err != nil {
				err = fmt.Errorf("validate category: %w", err)
				return
			}
			if categoryCount != 1 {
				err = errors.New("category does not belong to project and product")
				return
			}
		}
	}
	return nil
}

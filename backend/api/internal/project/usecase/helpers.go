package usecase

import (
	"errors"
	"fmt"
	"strings"

	"omnilogs-api/models"
	"omnilogs-api/responses"
	"omnilogs-api/utils"

	"gorm.io/gorm"
)

func (u *usecase) exists(model any, query string, args ...any) bool {
	return existsDB(u.repository.DB(), model, query, args...)
}

func existsDB(db *gorm.DB, model any, query string, args ...any) bool {
	var count int64
	return db.Model(model).Where(query, args...).Count(&count).Error == nil && count > 0
}

func normalizeCode(value string) string { return utils.NormalizeCode(value) }

func classifyDBError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return responses.ErrNotFound
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "duplicate key") || strings.Contains(message, "unique constraint") {
		return responses.ErrConflict
	}
	return err
}

func pathContains(path *string, id int) bool {
	if path == nil {
		return false
	}
	needle := "," + fmt.Sprint(id) + ","
	return strings.Contains(","+*path+",", needle)
}

func permissionListAllows(values []models.ProductRolePermission, resource, action string) bool {
	resource, action = strings.ToUpper(resource), strings.ToUpper(action)
	for _, value := range values {
		if strings.EqualFold(value.ResourceType, resource) && strings.EqualFold(value.Action, action) {
			return true
		}
	}
	return false
}

func membershipReadAllowed(resource string) bool {
	switch strings.ToUpper(resource) {
	case "PRODUCT", "PROJECT", "FEATURE", "CATEGORY", "ENVIRONMENT", "LOG":
		return true
	default:
		return false
	}
}

func validScope(db *gorm.DB, productID int, level string, projectID, categoryID *int) bool {
	switch level {
	case "PRODUCT":
		return projectID == nil && categoryID == nil
	case "PROJECT":
		return projectID != nil && categoryID == nil && existsDB(db, &models.Project{}, "project_id = ? AND product_id = ?", *projectID, productID)
	case "CATEGORY":
		if projectID == nil || categoryID == nil {
			return false
		}
		return existsDB(db, &models.ProjectFeature{}, "category_id = ? AND project_id = ? AND product_id = ?", *categoryID, *projectID, productID)
	default:
		return false
	}
}

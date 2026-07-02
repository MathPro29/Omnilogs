package usecase

import (
	"errors"
	"fmt"
	"strings"

	"omnilogs-api/models"

	"gorm.io/gorm"
)

func (u *usecase) exists(model any, query string, args ...any) bool {
	return existsDB(u.repository.DB(), model, query, args...)
}

func existsDB(db *gorm.DB, model any, query string, args ...any) bool {
	var count int64
	return db.Model(model).Where(query, args...).Count(&count).Error == nil && count > 0
}

func normalizeCode(value string) string { return strings.ToLower(strings.TrimSpace(value)) }

func classifyDBError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "duplicate key") || strings.Contains(message, "unique constraint") {
		return ErrConflict
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

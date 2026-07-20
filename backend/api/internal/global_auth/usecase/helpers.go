package usecase

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"omnilogs-api/dto"
	"omnilogs-api/models"
	"omnilogs-api/responses"

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
		return responses.ErrNotFound
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "duplicate key") || strings.Contains(message, "unique constraint") {
		return responses.ErrConflict
	}
	return err
}

func toAPIKeyResponse(value *models.ProductAPIKey) dto.APIKeyResponse {
	return dto.APIKeyResponse{KeyID: value.KeyID, ProductID: value.ProductID, EnvironmentID: value.EnvironmentID, KeyName: value.KeyName, KeyPrefix: value.KeyPrefix, Permissions: value.Permissions, IsActive: value.IsActive, ExpiresAt: value.ExpiresAt, TimestampResponse: dto.TimestampResponse{CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt}}
}

func pathContains(path *string, id int) bool {
	if path == nil {
		return false
	}
	needle := "," + fmt.Sprint(id) + ","
	return strings.Contains(","+*path+",", needle)
}

func validRolePermissions(values []dto.RolePermissionAssignment) bool {
	if len(values) == 0 {
		return false
	}
	seen := map[string]struct{}{}
	for _, value := range values {
		resource := strings.ToUpper(strings.TrimSpace(value.ResourceType))
		action := strings.ToUpper(strings.TrimSpace(value.Action))
		if resource == "" || action == "" || !validPermissionResource(resource) || !validPermissionAction(resource, action) {
			return false
		}
		key := resource + ":" + action
		if _, ok := seen[key]; ok {
			return false
		}
		seen[key] = struct{}{}
	}
	return true
}

func validPermissionResource(resource string) bool {
	allowed := map[string]bool{
		"PRODUCT": true, "PROJECT": true, "FEATURE": true, "CATEGORY": true,
		"ROLE": true, "ACCESS": true, "USER": true, "API_KEY": true,
		"ENVIRONMENT": true, "LOG": true, "ELASTIC_INDEX_POLICY": true,
		"CUSTOM_FIELD": true,
	}
	if allowed[resource] {
		return true
	}
	fieldID, ok := strings.CutPrefix(resource, "CUSTOM_FIELD:")
	if !ok {
		return false
	}
	id, err := strconv.Atoi(fieldID)
	return err == nil && id > 0
}

func validPermissionAction(resource, action string) bool {
	standard := map[string]bool{
		"CREATE": true, "READ": true, "UPDATE": true, "DELETE": true,
		"GRANT": true, "REVOKE": true, "EXPORT": true, "VIEW_SENSITIVE": true,
	}
	if standard[action] {
		return true
	}
	if resource == "CUSTOM_FIELD" || strings.HasPrefix(resource, "CUSTOM_FIELD:") {
		return map[string]bool{"VISIBLE": true, "SEARCH": true, "FILTER": true, "SORT": true, "AGGREGATE": true}[action]
	}
	return false
}

func toRolePermissions(roleID int, values []dto.RolePermissionAssignment) []models.ProductRolePermission {
	permissions := make([]models.ProductRolePermission, 0, len(values))
	for _, value := range values {
		permissions = append(permissions, models.ProductRolePermission{
			RoleID:       roleID,
			ResourceType: strings.ToUpper(strings.TrimSpace(value.ResourceType)),
			Action:       strings.ToUpper(strings.TrimSpace(value.Action)),
		})
	}
	return permissions
}

func permissionListAllows(values []models.ProductRolePermission, resource, action string) bool {
	resource, action = strings.ToUpper(resource), strings.ToUpper(action)
	for _, value := range values {
		roleResource := strings.ToUpper(value.ResourceType)
		matchesResource := roleResource == resource || (strings.HasPrefix(resource, "CUSTOM_FIELD:") && roleResource == "CUSTOM_FIELD")
		if matchesResource && strings.EqualFold(value.Action, action) {
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

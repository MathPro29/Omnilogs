package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"omnilogs-api/dto"
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

func validJSON(value json.RawMessage) bool { return len(value) > 0 && json.Valid(value) }

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

func permissionJSONAllows(raw json.RawMessage, resource, action string) bool {
	var value map[string]any
	if json.Unmarshal(raw, &value) != nil {
		return false
	}
	if all, ok := value["all"].(bool); ok && all {
		return true
	}
	resource, action = strings.ToUpper(resource), strings.ToUpper(action)
	for _, key := range []string{resource, strings.ToLower(resource)} {
		entry, ok := value[key]
		if !ok {
			continue
		}
		switch typed := entry.(type) {
		case bool:
			if typed {
				return true
			}
		case []any:
			for _, item := range typed {
				if strings.EqualFold(fmt.Sprint(item), action) {
					return true
				}
			}
		case map[string]any:
			for k, item := range typed {
				if strings.EqualFold(k, action) {
					allowed, _ := item.(bool)
					return allowed
				}
			}
		}
	}
	for key, item := range value {
		if strings.EqualFold(key, resource+"."+action) {
			allowed, _ := item.(bool)
			return allowed
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

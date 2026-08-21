package usecase

import (
	"errors"
	"strings"

	"omnilogs-api/dto"
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

func trimStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

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

func permissionListAllows(values []models.ProductRolePermission, resource, action string) bool {
	resource, action = strings.ToUpper(resource), strings.ToUpper(action)
	for _, value := range values {
		if strings.EqualFold(value.ResourceType, resource) && strings.EqualFold(value.Action, action) {
			return true
		}
	}
	return false
}

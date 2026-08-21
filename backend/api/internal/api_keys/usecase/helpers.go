package usecase

import (
	"encoding/json"
	"errors"
	"strings"

	"omnilogs-api/dto"
	"omnilogs-api/models"

	"gorm.io/gorm"
)

func existsDB(db *gorm.DB, model any, query string, args ...any) bool {
	var count int64
	return db.Model(model).Where(query, args...).Count(&count).Error == nil && count > 0
}

func validJSON(value json.RawMessage) bool {
	return len(value) > 0 && json.Valid(value)
}

const logIngestPermission = "LOG_INGEST_CREATE"

// normalizeIngestPermissions accepts the only permission format currently
// enforced by the public ingestion endpoint. Keeping this strict prevents the
// management API from issuing keys that will always be rejected at ingest.
func normalizeIngestPermissions(value json.RawMessage) (json.RawMessage, error) {
	if len(value) == 0 || string(value) == "null" {
		return json.RawMessage(`["LOG_INGEST_CREATE"]`), nil
	}

	var permissions []string
	if err := json.Unmarshal(value, &permissions); err != nil || len(permissions) == 0 {
		return nil, ErrInvalid
	}

	for _, permission := range permissions {
		if strings.EqualFold(strings.TrimSpace(permission), logIngestPermission) {
			return value, nil
		}
	}

	return nil, ErrInvalid
}

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
		return ErrNotFound
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "duplicate key") || strings.Contains(message, "unique constraint") {
		return ErrConflict
	}
	return err
}

func toAPIKeyResponse(value *models.ProductAPIKey) dto.APIKeyResponse {
	return dto.APIKeyResponse{
		KeyID:             value.KeyID,
		ProductID:         value.ProductID,
		EnvironmentID:     value.EnvironmentID,
		SourceID:          value.SourceID,
		DefaultProjectID:  value.DefaultProjectID,
		DefaultCategoryID: value.DefaultCategoryID,
		KeyName:           value.KeyName,
		KeyPrefix:         value.KeyPrefix,
		Permissions:       value.Permissions,
		IsActive:          value.IsActive,
		ExpiresAt:         value.ExpiresAt,
		TimestampResponse: dto.TimestampResponse{
			CreatedAt: value.CreatedAt,
			UpdatedAt: value.UpdatedAt,
		},
	}
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

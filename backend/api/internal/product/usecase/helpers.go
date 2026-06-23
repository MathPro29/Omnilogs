package usecase

import (
	"encoding/json"
	"errors"
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

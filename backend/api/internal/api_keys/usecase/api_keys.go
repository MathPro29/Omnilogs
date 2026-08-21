package usecase

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"omnilogs-api/dto"
	"omnilogs-api/models"
)

func (u *usecase) CreateAPIKey(actor Actor, productID int, req dto.CreateAPIKeyRequest) (*dto.CreateAPIKeyResponse, error) {
	if err := u.authorize(actor, productID, "CREATE"); err != nil {
		return nil, err
	}
	if req.ProductID != 0 && req.ProductID != productID {
		return nil, ErrInvalid
	}
	if req.EnvironmentID == nil && req.EnvironmentCode != "" {
		var env models.ProductEnvironment
		if err := u.repository.DB().Where("product_id = ? AND (LOWER(environment_code) = LOWER(?) OR LOWER(environment_name) = LOWER(?))", productID, req.EnvironmentCode, req.EnvironmentCode).First(&env).Error; err == nil {
			req.EnvironmentID = &env.EnvironmentID
		}
	}
	// Every key is scoped to exactly one environment. Product-wide keys are not
	// allowed, and an environment can have only one active key at a time.
	if req.EnvironmentID == nil || !existsDB(u.repository.DB(), &models.ProductEnvironment{}, "environment_id = ? AND product_id = ?", *req.EnvironmentID, productID) {
		return nil, ErrInvalid
	}
	var activeKeyCount int64
	if err := u.repository.DB().Model(&models.ProductAPIKey{}).
		Where("product_id = ? AND environment_id = ? AND is_active = TRUE", productID, *req.EnvironmentID).
		Count(&activeKeyCount).Error; err != nil {
		return nil, err
	}
	if activeKeyCount > 0 {
		return nil, ErrConflict
	}
	if req.SourceID != nil {
		query := u.repository.DB().Model(&models.LogSource{}).Where("source_id = ? AND product_id = ? AND is_active = TRUE", *req.SourceID, productID)
		if req.EnvironmentID != nil {
			query = query.Where("environment_id IS NULL OR environment_id = ?", *req.EnvironmentID)
		}
		var count int64
		if query.Count(&count).Error != nil || count == 0 {
			return nil, ErrInvalid
		}
	}
	if req.DefaultProjectID != nil && !existsDB(u.repository.DB(), &models.Project{}, "project_id = ? AND product_id = ? AND is_active = TRUE", *req.DefaultProjectID, productID) {
		return nil, ErrInvalid
	}
	if req.DefaultCategoryID != nil {
		if req.DefaultProjectID == nil || !existsDB(u.repository.DB(), &models.ProjectFeature{}, "category_id = ? AND project_id = ? AND product_id = ? AND is_active = TRUE", *req.DefaultCategoryID, *req.DefaultProjectID, productID) {
			return nil, ErrInvalid
		}
	}
	permissions, err := normalizeIngestPermissions(req.Permissions)
	if err != nil {
		return nil, err
	}
	req.Permissions = permissions
	if req.ExpiresAt != nil && !req.ExpiresAt.After(time.Now()) {
		return nil, ErrInvalid
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}

	secret := "omni_keys_" + hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(secret))
	value := &models.ProductAPIKey{
		ProductID:         productID,
		EnvironmentID:     req.EnvironmentID,
		SourceID:          req.SourceID,
		DefaultProjectID:  req.DefaultProjectID,
		DefaultCategoryID: req.DefaultCategoryID,
		KeyName:           trimStringPtr(req.KeyName),
		KeyPrefix:         secret[:13],
		KeyHash:           hex.EncodeToString(sum[:]),
		Permissions:       req.Permissions,
		IsActive:          true,
		ExpiresAt:         req.ExpiresAt,
	}
	if err := u.repository.DB().Create(value).Error; err != nil {
		return nil, classifyDBError(err)
	}

	response := toAPIKeyResponse(value)
	return &dto.CreateAPIKeyResponse{APIKeyResponse: response, APIKey: secret}, nil
}

func (u *usecase) ListAPIKeys(actor Actor, productID int) ([]dto.APIKeyResponse, error) {
	if err := u.authorize(actor, productID, "READ"); err != nil {
		return nil, err
	}

	var values []models.ProductAPIKey
	if err := u.repository.DB().Where("product_id = ?", productID).Order("key_id").Find(&values).Error; err != nil {
		return nil, err
	}

	result := make([]dto.APIKeyResponse, 0, len(values))
	for i := range values {
		result = append(result, toAPIKeyResponse(&values[i]))
	}
	return result, nil
}

func (u *usecase) UpdateAPIKey(actor Actor, productID, keyID int, req dto.UpdateAPIKeyRequest) (*dto.APIKeyResponse, error) {
	if err := u.authorize(actor, productID, "UPDATE"); err != nil {
		return nil, err
	}

	updates := map[string]any{}
	if req.KeyName != nil {
		updates["key_name"] = trimStringPtr(req.KeyName)
	}
	if len(req.Permissions) > 0 {
		permissions, err := normalizeIngestPermissions(req.Permissions)
		if err != nil {
			return nil, err
		}
		updates["permissions"] = permissions
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.ExpiresAt != nil {
		if !req.ExpiresAt.After(time.Now()) {
			return nil, ErrInvalid
		}
		updates["expires_at"] = req.ExpiresAt
	}

	res := u.repository.DB().Model(&models.ProductAPIKey{}).
		Where("key_id = ? AND product_id = ?", keyID, productID).
		Updates(updates)
	if res.Error != nil {
		return nil, classifyDBError(res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}

	var value models.ProductAPIKey
	if err := u.repository.DB().Where("key_id = ? AND product_id = ?", keyID, productID).First(&value).Error; err != nil {
		return nil, classifyDBError(err)
	}

	response := toAPIKeyResponse(&value)
	return &response, nil
}

func (u *usecase) RevokeAPIKey(actor Actor, productID, keyID int) error {
	if err := u.authorize(actor, productID, "DELETE"); err != nil {
		return err
	}

	now := time.Now()
	res := u.repository.DB().Model(&models.ProductAPIKey{}).
		Where("key_id = ? AND product_id = ?", keyID, productID).
		Updates(map[string]any{"is_active": false, "revoked_at": &now})
	if res.Error != nil {
		return classifyDBError(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

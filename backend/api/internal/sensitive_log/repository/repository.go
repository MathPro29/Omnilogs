package repository

import (
	"context"
	"gorm.io/gorm"
	"omnilogs-api/models"
)

type Repository interface {
	CreateRequest(ctx context.Context, req *models.SensitiveLogAccessRequest) error
	GetRequest(ctx context.Context, id string) (*models.SensitiveLogAccessRequest, error)
	ListRequests(ctx context.Context, productID int) ([]models.SensitiveLogAccessRequest, error)
	UpdateRequest(ctx context.Context, req *models.SensitiveLogAccessRequest) error

	ListSecrets(ctx context.Context, productID int) ([]models.LogSensitiveFieldSecret, error)
	GetSecret(ctx context.Context, secretID string) (*models.LogSensitiveFieldSecret, error)
	GetSecretByLogAndPath(ctx context.Context, logID, path string) (*models.LogSensitiveFieldSecret, error)
	GetPostgresPayload(ctx context.Context, logID string) (*models.LogObjectStorageRef, error)
	HasRawMainLogAccess(ctx context.Context, userID int, productID int) (bool, error)

	CreateHistory(ctx context.Context, hist *models.SensitiveLogAccessHistory) error
	ListHistory(ctx context.Context, productID int) ([]models.SensitiveLogAccessHistory, error)
	GetUser(ctx context.Context, userID int) (*models.User, error)
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) CreateRequest(ctx context.Context, req *models.SensitiveLogAccessRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *repository) GetRequest(ctx context.Context, id string) (*models.SensitiveLogAccessRequest, error) {
	var req models.SensitiveLogAccessRequest
	err := r.db.WithContext(ctx).First(&req, "request_id = ?", id).Error
	return &req, err
}

func (r *repository) ListRequests(ctx context.Context, productID int) ([]models.SensitiveLogAccessRequest, error) {
	var list []models.SensitiveLogAccessRequest
	err := r.db.WithContext(ctx).Where("product_id = ?", productID).Order("created_at desc").Find(&list).Error
	return list, err
}

func (r *repository) UpdateRequest(ctx context.Context, req *models.SensitiveLogAccessRequest) error {
	return r.db.WithContext(ctx).Save(req).Error
}

func (r *repository) ListSecrets(ctx context.Context, productID int) ([]models.LogSensitiveFieldSecret, error) {
	var list []models.LogSensitiveFieldSecret
	err := r.db.WithContext(ctx).
		Where("product_id = ? AND purged_at IS NULL", productID).
		Order("created_at desc").
		Limit(200).
		Find(&list).Error
	return list, err
}

func (r *repository) GetSecret(ctx context.Context, secretID string) (*models.LogSensitiveFieldSecret, error) {
	var s models.LogSensitiveFieldSecret
	err := r.db.WithContext(ctx).First(&s, "secret_id = ?", secretID).Error
	return &s, err
}

func (r *repository) GetSecretByLogAndPath(ctx context.Context, logID, path string) (*models.LogSensitiveFieldSecret, error) {
	var s models.LogSensitiveFieldSecret
	err := r.db.WithContext(ctx).First(&s, "log_id = ? AND field_path = ?", logID, path).Error
	return &s, err
}

func (r *repository) GetPostgresPayload(ctx context.Context, logID string) (*models.LogObjectStorageRef, error) {
	var objectRef models.LogObjectStorageRef
	err := r.db.WithContext(ctx).
		Where("log_id = ? AND object_type = ? AND purged_at IS NULL", logID, "INPUT_PAYLOAD").
		Order("created_at DESC NULLS LAST").
		First(&objectRef).Error
	return &objectRef, err
}

func (r *repository) HasRawMainLogAccess(ctx context.Context, userID int, productID int) (bool, error) {
	var denied int64
	if err := r.db.WithContext(ctx).Model(&models.UserRolePermissionRule{}).
		Where("user_id = ? AND product_id = ? AND resource_type = 'LOG' AND action = 'VIEW_SENSITIVE' AND effect = 'DENY' AND is_active = TRUE AND (expires_at IS NULL OR expires_at > NOW())", userID, productID).
		Count(&denied).Error; err != nil {
		return false, err
	}
	if denied > 0 {
		return false, nil
	}

	var allowed int64
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*)
		FROM product_memberships pm
		JOIN product_roles pr ON pr.role_id = pm.role_id AND pr.product_id = pm.product_id AND pr.is_active = TRUE
		JOIN product_role_permissions pp ON pp.role_id = pm.role_id
		WHERE pm.user_id = ? AND pm.product_id = ? AND pm.is_active = TRUE
		  AND (pm.expires_at IS NULL OR pm.expires_at > NOW())
		  AND pp.resource_type = 'LOG' AND pp.action = 'VIEW_SENSITIVE'
	`, userID, productID).Scan(&allowed).Error; err != nil {
		return false, err
	}
	if allowed > 0 {
		return true, nil
	}

	var explicitAllowed int64
	if err := r.db.WithContext(ctx).Model(&models.UserRolePermissionRule{}).
		Where("user_id = ? AND product_id = ? AND resource_type = 'LOG' AND action = 'VIEW_SENSITIVE' AND effect = 'ALLOW' AND is_active = TRUE AND (expires_at IS NULL OR expires_at > NOW())", userID, productID).
		Count(&explicitAllowed).Error; err != nil {
		return false, err
	}
	if explicitAllowed > 0 {
		return true, nil
	}

	var temporaryAccess int64
	if err := r.db.WithContext(ctx).Model(&models.SensitiveLogAccessRequest{}).
		Where("user_id = ? AND product_id = ? AND field_path = ? AND approval_status = 'APPROVED' AND (expires_at IS NULL OR expires_at > NOW())", userID, productID, "$main_logs").
		Count(&temporaryAccess).Error; err != nil {
		return false, err
	}
	return temporaryAccess > 0, nil
}

func (r *repository) CreateHistory(ctx context.Context, hist *models.SensitiveLogAccessHistory) error {
	return r.db.WithContext(ctx).Create(hist).Error
}

func (r *repository) ListHistory(ctx context.Context, productID int) ([]models.SensitiveLogAccessHistory, error) {
	var list []models.SensitiveLogAccessHistory
	err := r.db.WithContext(ctx).Where("product_id = ?", productID).Order("created_at desc").Find(&list).Error
	return list, err
}

func (r *repository) GetUser(ctx context.Context, userID int) (*models.User, error) {
	var u models.User
	err := r.db.WithContext(ctx).First(&u, "user_id = ?", userID).Error
	return &u, err
}

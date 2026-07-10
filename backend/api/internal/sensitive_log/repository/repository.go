package repository

import (
	"context"
	"omnilogs-api/models"
	"gorm.io/gorm"
)

type Repository interface {
	CreateRequest(ctx context.Context, req *models.SensitiveLogAccessRequest) error
	GetRequest(ctx context.Context, id string) (*models.SensitiveLogAccessRequest, error)
	ListRequests(ctx context.Context, productID int) ([]models.SensitiveLogAccessRequest, error)
	UpdateRequest(ctx context.Context, req *models.SensitiveLogAccessRequest) error
	
	ListSecrets(ctx context.Context, productID int) ([]models.LogSensitiveFieldSecret, error)
	GetSecret(ctx context.Context, secretID string) (*models.LogSensitiveFieldSecret, error)
	GetSecretByLogAndPath(ctx context.Context, logID, path string) (*models.LogSensitiveFieldSecret, error)
	
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

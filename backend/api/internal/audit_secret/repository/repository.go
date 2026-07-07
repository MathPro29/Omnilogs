package repository

import (
	"context"

	"omnilogs-api/models"

	"gorm.io/gorm"
)

type Repository interface {
	GetAudit(ctx context.Context, auditID string) (*models.SystemAuditLog, error)
	GetSecret(ctx context.Context, auditID string, secretID string) (*models.AuditSecret, error)
	CreateRequest(ctx context.Context, req *models.AuditSecretAccessRequest) error
	GetRequest(ctx context.Context, requestID string) (*models.AuditSecretAccessRequest, error)
	UpdateRequest(ctx context.Context, req *models.AuditSecretAccessRequest) error
	ListRequests(ctx context.Context, auditID string) ([]models.AuditSecretAccessRequest, error)
	CreateHistory(ctx context.Context, hist *models.AuditSecretAccessHistory) error
	ListHistory(ctx context.Context, auditID string) ([]models.AuditSecretAccessHistory, error)
	GetUser(ctx context.Context, userID int) (*models.User, error)
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) GetAudit(ctx context.Context, auditID string) (*models.SystemAuditLog, error) {
	var audit models.SystemAuditLog
	err := r.db.WithContext(ctx).First(&audit, "audit_id = ?", auditID).Error
	return &audit, err
}

func (r *repository) GetSecret(ctx context.Context, auditID string, secretID string) (*models.AuditSecret, error) {
	var secret models.AuditSecret
	err := r.db.WithContext(ctx).First(&secret, "audit_id = ? AND secret_id = ?", auditID, secretID).Error
	return &secret, err
}

func (r *repository) CreateRequest(ctx context.Context, req *models.AuditSecretAccessRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *repository) GetRequest(ctx context.Context, requestID string) (*models.AuditSecretAccessRequest, error) {
	var req models.AuditSecretAccessRequest
	err := r.db.WithContext(ctx).First(&req, "request_id = ?", requestID).Error
	return &req, err
}

func (r *repository) UpdateRequest(ctx context.Context, req *models.AuditSecretAccessRequest) error {
	return r.db.WithContext(ctx).Save(req).Error
}

func (r *repository) ListRequests(ctx context.Context, auditID string) ([]models.AuditSecretAccessRequest, error) {
	var list []models.AuditSecretAccessRequest
	err := r.db.WithContext(ctx).Where("audit_id = ?", auditID).Order("created_at desc").Find(&list).Error
	return list, err
}

func (r *repository) CreateHistory(ctx context.Context, hist *models.AuditSecretAccessHistory) error {
	return r.db.WithContext(ctx).Create(hist).Error
}

func (r *repository) ListHistory(ctx context.Context, auditID string) ([]models.AuditSecretAccessHistory, error) {
	var list []models.AuditSecretAccessHistory
	err := r.db.WithContext(ctx).Where("audit_id = ?", auditID).Order("created_at desc").Find(&list).Error
	return list, err
}

func (r *repository) GetUser(ctx context.Context, userID int) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, "user_id = ?", userID).Error
	return &user, err
}

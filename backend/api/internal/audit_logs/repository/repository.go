package repository

import (
	"context"
	"strings"

	"omnilogs-api/dto"
	"omnilogs-api/models"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, audit *models.SystemAuditLog) error
	FindByID(ctx context.Context, auditID string) (*models.SystemAuditLog, error)
	List(ctx context.Context, filter dto.AuditLogFilterRequest) ([]models.SystemAuditLog, int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, audit *models.SystemAuditLog) error {
	return r.db.WithContext(ctx).Create(audit).Error
}

func (r *repository) FindByID(ctx context.Context, auditID string) (*models.SystemAuditLog, error) {
	var value models.SystemAuditLog
	if err := r.db.WithContext(ctx).Where("audit_id = ?", auditID).First(&value).Error; err != nil {
		return nil, err
	}
	return &value, nil
}

func (r *repository) List(ctx context.Context, filter dto.AuditLogFilterRequest) ([]models.SystemAuditLog, int64, error) {
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}

	query := r.db.WithContext(ctx).Model(&models.SystemAuditLog{})
	if filter.AllowedProductIDs != nil {
		if len(filter.AllowedProductIDs) == 0 {
			query = query.Where("1 = 0")
		} else {
			query = query.Where("product_id IN ?", filter.AllowedProductIDs)
		}
	}
	if filter.ActorUserID != nil {
		query = query.Where("actor_user_id = ?", *filter.ActorUserID)
	}
	if filter.ProductID != nil {
		query = query.Where("product_id = ?", *filter.ProductID)
	}
	if filter.Action != nil && strings.TrimSpace(*filter.Action) != "" {
		query = query.Where("action = ?", strings.ToUpper(strings.TrimSpace(*filter.Action)))
	}
	if filter.ResourceType != nil && strings.TrimSpace(*filter.ResourceType) != "" {
		query = query.Where("resource_type = ?", strings.ToUpper(strings.TrimSpace(*filter.ResourceType)))
	}
	if filter.ResourceID != nil && strings.TrimSpace(*filter.ResourceID) != "" {
		query = query.Where("resource_id = ?", strings.TrimSpace(*filter.ResourceID))
	}
	if filter.Result != nil && strings.TrimSpace(*filter.Result) != "" {
		query = query.Where("result = ?", strings.ToUpper(strings.TrimSpace(*filter.Result)))
	}
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		query = query.Where("created_at <= ?", *filter.DateTo)
	}
	if filter.Keyword != nil && strings.TrimSpace(*filter.Keyword) != "" {
		keyword := "%" + strings.TrimSpace(*filter.Keyword) + "%"
		query = query.Where(
			"action ILIKE ? OR resource_type ILIKE ? OR resource_id ILIKE ? OR request_id ILIKE ? OR trace_id ILIKE ? OR path ILIKE ?",
			keyword, keyword, keyword, keyword, keyword, keyword,
		)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var values []models.SystemAuditLog
	if err := query.Order("created_at DESC").Limit(perPage).Offset((page - 1) * perPage).Find(&values).Error; err != nil {
		return nil, 0, err
	}

	return values, total, nil
}

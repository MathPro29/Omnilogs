package repository

import (
	"context"
	"fmt"

	"omnilogs-api/dto"
	"omnilogs-api/models"
)

func (r *repository) GetAuditLogs(ctx context.Context, q dto.AuditLogQuery) ([]models.SystemAuditLog, int64, error) {
	var auditLogs []models.SystemAuditLog
	query := r.db.WithContext(ctx).Model(&models.SystemAuditLog{})
	if q.AllowedProductIDs != nil {
		if len(q.AllowedProductIDs) == 0 {
			query = query.Where("1 = 0")
		} else {
			query = query.Where("product_id IN ?", q.AllowedProductIDs)
		}
	}

	if q.ProductID != nil {
		query = query.Where("product_id = ?", *q.ProductID)
	}
	if q.ProjectID != nil {
		projIDStr := fmt.Sprintf("%d", *q.ProjectID)
		query = query.Where(
			"metadata->'query'->>'project_id' = ? OR metadata->'request'->>'project_id' = ? OR path LIKE ?",
			projIDStr, projIDStr, "%/projects/"+projIDStr+"%",
		)
	}
	if q.FeatureID != nil {
		featIDStr := fmt.Sprintf("%d", *q.FeatureID)
		query = query.Where(
			"metadata->'query'->>'category_id' = ? OR metadata->'request'->>'category_id' = ? OR "+
				"metadata->'query'->>'feature_id' = ? OR metadata->'request'->>'feature_id' = ? OR "+
				"path LIKE ? OR path LIKE ?",
			featIDStr, featIDStr, featIDStr, featIDStr, "%/features/"+featIDStr, "%/features/"+featIDStr+"/%",
		)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Limit(q.Limit).Offset(q.Offset).Find(&auditLogs).Error; err != nil {
		return nil, 0, err
	}

	return auditLogs, total, nil
}

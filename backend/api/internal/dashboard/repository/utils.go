package repository

import (
	"fmt"
	"strconv"
	"strings"

	"omnilogs-api/models"
)

// parseCSVInt64 parses a comma-separated string of int64 values
func parseCSVInt64(csv string) []int64 {
	if strings.TrimSpace(csv) == "" {
		return nil
	}
	parts := strings.Split(csv, ",")
	result := make([]int64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if v, err := strconv.ParseInt(part, 10, 64); err == nil && v > 0 {
			result = append(result, v)
		}
	}
	return result
}

func (r *repository) resolveSearchIndices(productID int) []string {
	indices := []string{fmt.Sprintf("omnilogs-product-%d-*", productID)}
	var policies []models.ElasticIndexPolicy
	if err := r.db.Where("product_id = ? AND is_active = TRUE", productID).Find(&policies).Error; err == nil {
		for _, policy := range policies {
			if strings.TrimSpace(policy.IndexPrefix) != "" {
				prefix := strings.ToLower(strings.TrimSpace(policy.IndexPrefix))
				prefix = strings.ReplaceAll(prefix, "_", "-")
				prefix = strings.ReplaceAll(prefix, " ", "-")
				indices = append(indices, fmt.Sprintf("%s-*", prefix))
			}
		}
	}
	return indices
}

package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"omnilogs-api/models"
)

const maxFilterIDs = 100

// parseCSVInt64 parses a comma-separated string of int64 values
func parseCSVInt64(csv string) []int64 {
	if strings.TrimSpace(csv) == "" {
		return nil
	}
	parts := strings.Split(csv, ",")
	result := make([]int64, 0, min(len(parts), maxFilterIDs))
	seen := make(map[int64]struct{}, min(len(parts), maxFilterIDs))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if v, err := strconv.ParseInt(part, 10, 64); err == nil && v > 0 {
			if _, exists := seen[v]; exists {
				continue
			}
			seen[v] = struct{}{}
			result = append(result, v)
			if len(result) == maxFilterIDs {
				break
			}
		}
	}
	return result
}

func (r *repository) resolveSearchIndices(ctx context.Context, productID int) []string {
	indices := []string{fmt.Sprintf("omnilogs-product-%d-*", productID)}
	seen := map[string]struct{}{indices[0]: {}}
	var policies []models.ElasticIndexPolicy
	if err := r.db.WithContext(ctx).
		Select("index_prefix").
		Where("product_id = ? AND is_active = TRUE", productID).
		Find(&policies).Error; err == nil {
		for _, policy := range policies {
			if strings.TrimSpace(policy.IndexPrefix) != "" {
				prefix := strings.ToLower(strings.TrimSpace(policy.IndexPrefix))
				prefix = strings.ReplaceAll(prefix, "_", "-")
				prefix = strings.ReplaceAll(prefix, " ", "-")
				pattern := fmt.Sprintf("%s-*", prefix)
				if _, exists := seen[pattern]; !exists {
					seen[pattern] = struct{}{}
					indices = append(indices, pattern)
				}
			}
		}
	}
	return indices
}

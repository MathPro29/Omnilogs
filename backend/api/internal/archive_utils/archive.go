package archive_utils

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	retentionprovider "omnilogs-api/internal/retention_policy/provider"
	"omnilogs-api/models"

	"github.com/elastic/go-elasticsearch/v8"
	"gorm.io/gorm"
)

// ArchiveIndex is the compatibility entry point used by older callers. It now
// delegates to the streaming NDJSON writer and refuses unscoped product-wide
// archives because they cannot satisfy the product+environment delete guard.
func ArchiveIndex(ctx context.Context, db *gorm.DB, esClient *elasticsearch.Client, indexName string, productID int, environmentID *int, archiveFormat string, storagePath string) (int, int64, string, error) {
	if environmentID == nil || *environmentID <= 0 {
		return 0, 0, "", fmt.Errorf("legacy archive requires environment_id")
	}
	date, err := parseDateFromIndexName(indexName)
	if err != nil {
		return 0, 0, "", err
	}
	date = date.UTC()
	dateEnd := date.AddDate(0, 0, 1)
	root := strings.TrimPrefix(strings.TrimSpace(storagePath), "local://")
	if root == "" {
		root = "data"
	}
	filePath := filepath.Join(root, fmt.Sprintf("%s-env-%d.ndjson.gz", indexName, *environmentID))
	result, err := StreamArchive(StreamArchiveRequest{Client: esClient, Context: ctx, Indices: []string{indexName}, Query: retentionprovider.BuildQuery(productID, *environmentID, &date, &dateEnd, nil, nil, nil), ProductID: productID, EnvironmentID: *environmentID, DateFrom: date, DateTo: dateEnd, FilePath: filePath})
	if err != nil {
		return 0, 0, "", err
	}
	if result.Empty {
		return 0, 0, "", nil
	}
	return int(result.DocumentCount), result.CompressedBytes, result.FilePath, nil
}

// DeleteArchivedAuditLogs is deliberately separate from ArchiveIndex so the
// caller can persist the LogArchive record before deleting the source rows.
func DeleteArchivedAuditLogs(ctx context.Context, db *gorm.DB, indexName string, productID int, environmentID *int, archiveFormat string) error {
	if environmentID != nil || strings.ToUpper(strings.TrimSpace(archiveFormat)) == "CSV" {
		return nil
	}
	date, err := parseDateFromIndexName(indexName)
	if err != nil {
		return nil
	}
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24*time.Hour - time.Nanosecond)
	return db.WithContext(ctx).
		Where("product_id = ? AND created_at >= ? AND created_at <= ?", int64(productID), startOfDay, endOfDay).
		Delete(&models.SystemAuditLog{}).Error
}

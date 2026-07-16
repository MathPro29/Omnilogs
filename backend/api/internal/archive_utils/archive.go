package archive_utils

import (
	"compress/gzip"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"omnilogs-api/models"

	"github.com/elastic/go-elasticsearch/v8"
	"gorm.io/gorm"
)

// ArchiveIndex pulls logs from Elasticsearch and, for product-wide JSON
// archives, system audit logs from PostgreSQL, then writes a GZIP archive.
func ArchiveIndex(ctx context.Context, db *gorm.DB, esClient *elasticsearch.Client, indexName string, productID int, environmentID *int, archiveFormat string, storagePath string) (int, int64, string, error) {
	// The index is write-locked by the caller, so scrolling gives a stable and
	// complete snapshot instead of silently truncating archives at 10,000 logs.
	logs, err := fetchArchivedLogs(ctx, esClient, indexName, environmentID)
	if err != nil {
		return 0, 0, "", err
	}

	// 2. Fetch PostgreSQL system_audit_logs for this product on the index date
	var auditLogs []models.SystemAuditLog
	date, parseErr := parseDateFromIndexName(indexName)
	if environmentID == nil && !strings.EqualFold(strings.TrimSpace(archiveFormat), "CSV") && parseErr == nil {
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
		endOfDay := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, time.UTC)

		prodID64 := int64(productID)
		err = db.WithContext(ctx).
			Where("product_id = ? AND created_at >= ? AND created_at <= ?", prodID64, startOfDay, endOfDay).
			Find(&auditLogs).Error
		if err != nil {
			return 0, 0, "", fmt.Errorf("failed to query audit logs from postgres: %w", err)
		}
	}

	// 3. Compress data into .json.gz file locally
	archiveData := ArchiveData{
		Logs:      logs,
		AuditLogs: auditLogs,
	}

	dir := strings.TrimSpace(storagePath)
	if dir == "" {
		dir = fmt.Sprintf("data/archives/product-%d", productID)
	}
	dir = strings.TrimPrefix(dir, "local://")

	if err := os.MkdirAll(dir, 0755); err != nil {
		return 0, 0, "", err
	}

	ext := ".json.gz"
	formatUpper := strings.ToUpper(strings.TrimSpace(archiveFormat))
	if formatUpper == "CSV" {
		ext = ".csv.gz"
	}

	archiveName := indexName
	if environmentID != nil {
		archiveName = fmt.Sprintf("%s-env-%d", indexName, *environmentID)
	}
	filePath := filepath.Join(dir, archiveName+ext)
	file, err := os.Create(filePath)
	if err != nil {
		return 0, 0, "", err
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	var errEncode error
	if formatUpper == "CSV" {
		csvWriter := csv.NewWriter(gzipWriter)
		if err := csvWriter.Write([]string{"log_id", "timestamp", "message", "raw_payload"}); err != nil {
			errEncode = err
		} else {
			for _, logDoc := range logs {
				logID, _ := logDoc["_id"].(string)
				source, _ := logDoc["_source"].(map[string]any)
				timestamp, _ := source["@timestamp"].(string)
				payload, _ := source["payload"].(map[string]any)
				message := ""
				if payload != nil {
					message, _ = payload["message"].(string)
				}
				rawPayload, _ := json.Marshal(source)
				record := []string{logID, timestamp, message, string(rawPayload)}
				if err := csvWriter.Write(record); err != nil {
					errEncode = err
					break
				}
			}
			if errEncode == nil {
				csvWriter.Flush()
				errEncode = csvWriter.Error()
			}
		}
	} else {
		errEncode = json.NewEncoder(gzipWriter).Encode(archiveData)
	}

	if errEncode != nil {
		gzipWriter.Close()
		file.Close()
		os.Remove(filePath)
		return 0, 0, "", errEncode
	}

	if err := gzipWriter.Close(); err != nil {
		file.Close()
		os.Remove(filePath)
		return 0, 0, "", err
	}

	if err := file.Close(); err != nil {
		os.Remove(filePath)
		return 0, 0, "", err
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, 0, "", err
	}

	return len(logs), fileInfo.Size(), filePath, nil
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

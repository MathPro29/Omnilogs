package archive_utils

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"omnilogs-api/models"

	"github.com/elastic/go-elasticsearch/v8"
	"gorm.io/gorm"
)

// ArchiveIndex pulls logs from Elasticsearch and system audit logs from PostgreSQL,
// compresses them in GZIP, saves the archive locally, and deletes the postgres audit logs.
func ArchiveIndex(ctx context.Context, db *gorm.DB, esClient *elasticsearch.Client, indexName string, productID int, archiveFormat string, storagePath string) (int, int64, string, error) {
	// 1. Fetch logs from Elasticsearch index
	var buf bytes.Buffer
	query := map[string]any{
		"query": map[string]any{
			"match_all": map[string]any{},
		},
		"size": 10000,
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return 0, 0, "", err
	}

	res, err := esClient.Search(
		esClient.Search.WithContext(ctx),
		esClient.Search.WithIndex(indexName),
		esClient.Search.WithBody(&buf),
	)
	if err != nil {
		return 0, 0, "", err
	}
	defer res.Body.Close()

	var searchRes map[string]any
	if res.IsError() {
		// If index does not exist in Elasticsearch (empty logs), we can still proceed with audit logs or default to empty
		if res.StatusCode != 404 {
			return 0, 0, "", fmt.Errorf("es search returned error status: %s", res.Status())
		}
	} else {
		if err := json.NewDecoder(res.Body).Decode(&searchRes); err != nil {
			return 0, 0, "", err
		}
	}

	var logs []map[string]any
	if searchRes != nil {
		hitsObj, _ := searchRes["hits"].(map[string]any)
		hitList, _ := hitsObj["hits"].([]any)
		logs = make([]map[string]any, 0, len(hitList))
		for _, hit := range hitList {
			hitMap, ok := hit.(map[string]any)
			if !ok {
				continue
			}
			doc := map[string]any{
				"_id":     hitMap["_id"],
				"_source": hitMap["_source"],
			}
			logs = append(logs, doc)
		}
	}

	// 2. Fetch PostgreSQL system_audit_logs for this product on the index date
	var auditLogs []models.SystemAuditLog
	date, parseErr := parseDateFromIndexName(indexName)
	if parseErr == nil {
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

	filePath := fmt.Sprintf("%s/%s%s", dir, indexName, ext)
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

	// 4. Delete system_audit_logs from PostgreSQL since they are archived
	if len(auditLogs) > 0 && parseErr == nil {
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
		endOfDay := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, time.UTC)
		prodID64 := int64(productID)
		err = db.WithContext(ctx).
			Where("product_id = ? AND created_at >= ? AND created_at <= ?", prodID64, startOfDay, endOfDay).
			Delete(&models.SystemAuditLog{}).Error
		if err != nil {
			return 0, 0, "", fmt.Errorf("failed to delete postgres audit logs: %w", err)
		}
	}

	return len(logs), fileInfo.Size(), filePath, nil
}

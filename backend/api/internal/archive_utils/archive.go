package archive_utils

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"omnilogs-api/models"

	"github.com/elastic/go-elasticsearch/v8"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ArchiveData struct {
	Logs      []map[string]any        `json:"logs"`
	AuditLogs []models.SystemAuditLog `json:"audit_logs"`
}

func parseDateFromIndexName(indexName string) (time.Time, error) {
	parts := strings.Split(indexName, "-")
	if len(parts) < 2 {
		return time.Time{}, fmt.Errorf("invalid index name structure: %s", indexName)
	}
	dateStr := parts[len(parts)-1]
	return time.Parse("2006.01.02", dateStr)
}

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
		var csvBuf bytes.Buffer
		csvBuf.WriteString("log_id,timestamp,message,raw_payload\n")
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
			csvBuf.WriteString(fmt.Sprintf("%q,%q,%q,%q\n", logID, timestamp, message, string(rawPayload)))
		}
		_, errEncode = gzipWriter.Write(csvBuf.Bytes())
	} else {
		errEncode = json.NewEncoder(gzipWriter).Encode(archiveData)
	}

	if errEncode != nil {
		gzipWriter.Close()
		return 0, 0, "", errEncode
	}
	if err := gzipWriter.Close(); err != nil {
		return 0, 0, "", err
	}

	fileInfo, err := file.Stat()
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

// RestoreIndex reads the compressed archive, indexes logs back to Elasticsearch, and saves audits back to PostgreSQL.
func RestoreIndex(ctx context.Context, db *gorm.DB, esClient *elasticsearch.Client, filePath string, targetIndexName string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return 0, err
	}
	defer gzipReader.Close()

	var archiveData ArchiveData
	if err := json.NewDecoder(gzipReader).Decode(&archiveData); err != nil {
		return 0, err
	}

	// 1. Re-index Elasticsearch documents
	for _, doc := range archiveData.Logs {
		docID, _ := doc["_id"].(string)
		source, _ := doc["_source"].(map[string]any)

		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(source); err != nil {
			return 0, err
		}

		res, err := esClient.Index(
			targetIndexName,
			&buf,
			esClient.Index.WithContext(ctx),
			esClient.Index.WithDocumentID(docID),
		)
		if err != nil {
			return 0, fmt.Errorf("failed to index document %s during restore: %w", docID, err)
		}
		res.Body.Close()
		if res.IsError() {
			return 0, fmt.Errorf("failed to index document %s: status %s", docID, res.Status())
		}
	}

	// 2. Restore PostgreSQL SystemAuditLog logs
	if len(archiveData.AuditLogs) > 0 {
		err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			for _, audit := range archiveData.AuditLogs {
				err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&audit).Error
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return 0, fmt.Errorf("failed to restore postgres audit logs: %w", err)
		}
	}

	return len(archiveData.Logs), nil
}

// Helper function to decode compressed payload to inspect/read stream (if needed)
func GetDecompressedReader(filePath string) (io.ReadCloser, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		file.Close()
		return nil, err
	}
	return struct {
		io.Reader
		io.Closer
	}{
		Reader: gzipReader,
		Closer: gzipCloser{gzipReader, file},
	}, nil
}

type gzipCloser struct {
	gr *gzip.Reader
	f  *os.File
}

func (gc gzipCloser) Close() error {
	err1 := gc.gr.Close()
	err2 := gc.f.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

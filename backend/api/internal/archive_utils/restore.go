package archive_utils

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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

	archiveData, err := decodeArchiveData(gzipReader, filePath)
	if err != nil {
		return 0, err
	}

	// 1. Re-index Elasticsearch documents in bulk batches
	const bulkBatchSize = 1000
	for i := 0; i < len(archiveData.Logs); i += bulkBatchSize {
		end := i + bulkBatchSize
		if end > len(archiveData.Logs) {
			end = len(archiveData.Logs)
		}

		var body bytes.Buffer
		for _, doc := range archiveData.Logs[i:end] {
			docID, _ := doc["_id"].(string)
			source, _ := doc["_source"].(map[string]any)

			meta := map[string]any{
				"index": map[string]any{
					"_index": targetIndexName,
					"_id":    docID,
				},
			}
			metaBytes, err := json.Marshal(meta)
			if err != nil {
				return 0, err
			}
			docBytes, err := json.Marshal(source)
			if err != nil {
				return 0, err
			}
			body.Write(metaBytes)
			body.WriteByte('\n')
			body.Write(docBytes)
			body.WriteByte('\n')
		}

		res, err := esClient.Bulk(
			bytes.NewReader(body.Bytes()),
			esClient.Bulk.WithContext(ctx),
		)
		if err != nil {
			return 0, fmt.Errorf("failed to bulk index documents during restore: %w", err)
		}
		if res.IsError() {
			res.Body.Close()
			return 0, fmt.Errorf("bulk index returned error status: %s", res.Status())
		}

		var bulkResponse struct {
			Errors bool `json:"errors"`
			Items  []map[string]struct {
				Status int            `json:"status"`
				Error  map[string]any `json:"error"`
			} `json:"items"`
		}
		decodeErr := json.NewDecoder(res.Body).Decode(&bulkResponse)
		res.Body.Close()
		if decodeErr != nil {
			return 0, decodeErr
		}

		if bulkResponse.Errors {
			for _, itemMap := range bulkResponse.Items {
				for _, op := range itemMap {
					if op.Status >= 300 {
						reason := "bulk index item failed"
						if msg, ok := op.Error["reason"].(string); ok && msg != "" {
							reason = msg
						}
						return 0, fmt.Errorf("failed to bulk index document during restore: %s", reason)
					}
				}
			}
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

func decodeArchiveData(reader io.Reader, filePath string) (ArchiveData, error) {
	if !strings.HasSuffix(strings.ToLower(filePath), ".csv.gz") {
		var data ArchiveData
		err := json.NewDecoder(reader).Decode(&data)
		return data, err
	}

	records, err := csv.NewReader(reader).ReadAll()
	if err != nil {
		return ArchiveData{}, err
	}
	if len(records) == 0 {
		return ArchiveData{Logs: []map[string]any{}}, nil
	}
	if len(records[0]) != 4 || records[0][0] != "log_id" || records[0][3] != "raw_payload" {
		return ArchiveData{}, fmt.Errorf("invalid CSV archive header")
	}
	logs := make([]map[string]any, 0, len(records)-1)
	for rowNumber, record := range records[1:] {
		if len(record) != 4 {
			return ArchiveData{}, fmt.Errorf("invalid CSV archive row %d", rowNumber+2)
		}
		var source map[string]any
		if err := json.Unmarshal([]byte(record[3]), &source); err != nil {
			return ArchiveData{}, fmt.Errorf("invalid raw payload at CSV archive row %d: %w", rowNumber+2, err)
		}
		logs = append(logs, map[string]any{"_id": record[0], "_source": source})
	}
	return ArchiveData{Logs: logs}, nil
}

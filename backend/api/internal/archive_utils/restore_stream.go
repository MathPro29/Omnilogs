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

	"github.com/elastic/go-elasticsearch/v8"

	"omnilogs-api/internal/searchindex"
)

const (
	RestoreSearchOnly     = "SEARCH_ONLY"
	RestoreSnapshotSearch = "SNAPSHOT_SEARCH_ONLY"
	RestoreToActive       = "RESTORE_TO_ACTIVE"
	ConflictSkipExisting  = "SKIP_EXISTING"
	ConflictOverwrite     = "OVERWRITE"
)

type RestoreResult struct {
	Restored int64
	Skipped  int64
	Failed   int64
}

type archiveLogLine struct {
	Type   string         `json:"_type"`
	ID     string         `json:"_id"`
	Index  string         `json:"_index"`
	Source map[string]any `json:"_source"`
}

func RestoreNDJSON(ctx context.Context, client *elasticsearch.Client, filePath, mode, conflict, targetIndex string, expectedProductID, expectedEnvironmentID int) (RestoreResult, error) {
	if client == nil {
		return RestoreResult{}, fmt.Errorf("elasticsearch client is not configured")
	}
	if mode != RestoreSearchOnly && mode != RestoreToActive {
		return RestoreResult{}, fmt.Errorf("invalid restore mode")
	}
	if conflict == "" {
		conflict = ConflictSkipExisting
	}
	if conflict != ConflictSkipExisting && conflict != ConflictOverwrite {
		return RestoreResult{}, fmt.Errorf("invalid conflict strategy")
	}
	if targetIndex == "" {
		return RestoreResult{}, fmt.Errorf("target index is required")
	}
	manifest, _, err := VerifyArchive(filePath, expectedProductID, expectedEnvironmentID)
	if err != nil {
		return RestoreResult{}, err
	}
	_ = manifest
	if err := ensureRestoreIndex(ctx, client, targetIndex); err != nil {
		return RestoreResult{}, err
	}
	file, err := os.Open(filePath)
	if err != nil {
		return RestoreResult{}, err
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		return RestoreResult{}, err
	}
	defer gz.Close()
	decoder := json.NewDecoder(gz)
	var ignored ArchiveManifest
	if err := decoder.Decode(&ignored); err != nil {
		return RestoreResult{}, err
	}

	batch := make([]archiveLogLine, 0, 500)
	result := RestoreResult{}
	failureReasons := make([]string, 0, 3)
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		var body bytes.Buffer
		for _, line := range batch {
			indexName := targetIndex
			action := "index"
			if conflict == ConflictSkipExisting {
				action = "create"
			}
			meta := map[string]any{action: map[string]any{"_index": indexName, "_id": line.ID}}
			metaBytes, marshalErr := json.Marshal(meta)
			if marshalErr != nil {
				return marshalErr
			}
			sourceBytes, marshalErr := json.Marshal(line.Source)
			if marshalErr != nil {
				return marshalErr
			}
			body.Write(metaBytes)
			body.WriteByte('\n')
			body.Write(sourceBytes)
			body.WriteByte('\n')
		}
		// A successful bulk response only means Elasticsearch accepted the
		// writes. Wait until they are searchable before reporting the restore as
		// complete; re-archiving can otherwise run immediately and observe zero
		// documents in the active index.
		response, requestErr := client.Bulk(
			bytes.NewReader(body.Bytes()),
			client.Bulk.WithContext(ctx),
			client.Bulk.WithRefresh("wait_for"),
		)
		if requestErr != nil {
			return requestErr
		}
		defer response.Body.Close()
		if response.IsError() {
			return fmt.Errorf("restore bulk request returned status %s", response.Status())
		}
		var payload struct {
			Errors bool `json:"errors"`
			Items  []map[string]struct {
				Status int            `json:"status"`
				Error  map[string]any `json:"error"`
			} `json:"items"`
		}
		if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
			return err
		}
		for _, item := range payload.Items {
			for _, operation := range item {
				if operation.Status == 409 && conflict == ConflictSkipExisting {
					result.Skipped++
					continue
				}
				if operation.Status >= 300 {
					reason := "restore item failed"
					if value, ok := operation.Error["reason"].(string); ok && value != "" {
						reason = value
					}
					result.Failed++
					if len(failureReasons) < cap(failureReasons) {
						failureReasons = append(failureReasons, reason)
					}
					continue
				}
				result.Restored++
			}
		}
		batch = batch[:0]
		return nil
	}

	for {
		var line archiveLogLine
		err := decoder.Decode(&line)
		if err == io.EOF {
			break
		}
		if err != nil {
			return RestoreResult{}, err
		}
		if line.Type == "manifest" {
			continue
		}
		if line.Type != "log" || line.ID == "" || line.Source == nil {
			return RestoreResult{}, fmt.Errorf("invalid archive log line")
		}
		batch = append(batch, line)
		if len(batch) >= cap(batch) {
			if err := flush(); err != nil {
				return result, err
			}
		}
	}
	if err := flush(); err != nil {
		return result, err
	}
	if result.Failed > 0 {
		return result, fmt.Errorf("%d archived logs failed to restore: %s", result.Failed, strings.Join(failureReasons, "; "))
	}
	return result, nil
}

// ensureRestoreIndex prevents flexible JSON objects from exhausting the
// default Elasticsearch field limit. Existing rollback indexes may have been
// partially restored by older versions, so they are preserved and given a
// bounded compatibility limit; SKIP_EXISTING then fills only missing logs.
func ensureRestoreIndex(ctx context.Context, client *elasticsearch.Client, targetIndex string) error {
	existsResponse, err := client.Indices.Exists(
		[]string{targetIndex},
		client.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("check restore index %s: %w", targetIndex, err)
	}
	exists := existsResponse.StatusCode == 200
	status := existsResponse.StatusCode
	_ = existsResponse.Body.Close()
	if !exists && status != 404 {
		return fmt.Errorf("check restore index %s: unexpected status %d", targetIndex, status)
	}
	if !exists {
		body := append([]byte(`{"mappings":`), []byte(searchindex.ControlledLogMapping)...)
		body = append(body, '}')
		response, createErr := client.Indices.Create(
			targetIndex,
			client.Indices.Create.WithBody(bytes.NewReader(body)),
			client.Indices.Create.WithContext(ctx),
		)
		if createErr != nil {
			return fmt.Errorf("create restore index %s: %w", targetIndex, createErr)
		}
		defer response.Body.Close()
		if response.IsError() {
			message, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
			return fmt.Errorf("create restore index %s returned status %s: %s", targetIndex, response.Status(), string(message))
		}
		return nil
	}

	// Older releases allowed Elasticsearch to infer thousands of fields. Do
	// not recreate or delete that partially restored index; raise only its
	// bounded field limit so an idempotent retry can finish safely.
	response, err := client.Indices.PutSettings(
		bytes.NewReader([]byte(`{"index.mapping.total_fields.limit":10000}`)),
		client.Indices.PutSettings.WithIndex(targetIndex),
		client.Indices.PutSettings.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("prepare existing restore index %s: %w", targetIndex, err)
	}
	defer response.Body.Close()
	if response.IsError() {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("prepare existing restore index %s returned status %s: %s", targetIndex, response.Status(), string(message))
	}
	return nil
}

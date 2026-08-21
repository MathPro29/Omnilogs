package archive_utils

import (
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

func TestRestoreNDJSONCreatesControlledIndexAndSkipsExisting(t *testing.T) {
	archivePath := writeRestoreArchive(t, []archiveLogLine{
		{Type: "log", ID: "existing", Index: "source", Source: map[string]any{"product_id": 7, "environment_id": 9, "data": map[string]any{"arbitrary": "value"}}},
		{Type: "log", ID: "new", Index: "source", Source: map[string]any{"product_id": 7, "environment_id": 9, "message": "new"}},
	})
	createdWithControlledMapping := false
	bulkWaitedForRefresh := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		switch {
		case r.Method == http.MethodHead && r.URL.Path == "/restore-target":
			w.WriteHeader(http.StatusNotFound)
		case r.Method == http.MethodPut && r.URL.Path == "/restore-target":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode create index body: %v", err)
			}
			mappings, _ := body["mappings"].(map[string]any)
			properties, _ := mappings["properties"].(map[string]any)
			data, _ := properties["data"].(map[string]any)
			createdWithControlledMapping = mappings["dynamic"] == false && data["type"] == "flattened"
			_, _ = w.Write([]byte(`{"acknowledged":true}`))
		case r.Method == http.MethodPost && r.URL.Path == "/_bulk":
			bulkWaitedForRefresh = r.URL.Query().Get("refresh") == "wait_for"
			_, _ = w.Write([]byte(`{"errors":true,"items":[{"create":{"status":409,"error":{"reason":"already exists"}}},{"create":{"status":201}}]}`))
		default:
			http.Error(w, "unexpected request", http.StatusBadRequest)
		}
	}))
	defer server.Close()

	client, err := elasticsearch.NewClient(elasticsearch.Config{Addresses: []string{server.URL}})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	result, err := RestoreNDJSON(t.Context(), client, archivePath, RestoreToActive, ConflictSkipExisting, "restore-target", 7, 9)
	if err != nil {
		t.Fatalf("restore archive: %v", err)
	}
	if !createdWithControlledMapping {
		t.Fatal("restore target was not created with the controlled flattened mapping")
	}
	if !bulkWaitedForRefresh {
		t.Fatal("restore bulk request must wait until restored logs are searchable")
	}
	if result.Restored != 1 || result.Skipped != 1 {
		t.Fatalf("unexpected restore result: %#v", result)
	}
}

func TestRestoreNDJSONReportsPartialFailuresWithoutDiscardingCounts(t *testing.T) {
	archivePath := writeRestoreArchive(t, []archiveLogLine{
		{Type: "log", ID: "ok", Index: "source", Source: map[string]any{"product_id": 7, "environment_id": 9}},
		{Type: "log", ID: "bad", Index: "source", Source: map[string]any{"product_id": 7, "environment_id": 9}},
	})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		switch {
		case r.Method == http.MethodHead:
			w.WriteHeader(http.StatusNotFound)
		case r.Method == http.MethodPut:
			_, _ = w.Write([]byte(`{"acknowledged":true}`))
		case r.Method == http.MethodPost && r.URL.Path == "/_bulk":
			_, _ = w.Write([]byte(`{"errors":true,"items":[{"create":{"status":201}},{"create":{"status":400,"error":{"reason":"mapping rejected"}}}]}`))
		default:
			http.Error(w, "unexpected request", http.StatusBadRequest)
		}
	}))
	defer server.Close()
	client, err := elasticsearch.NewClient(elasticsearch.Config{Addresses: []string{server.URL}})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	result, err := RestoreNDJSON(t.Context(), client, archivePath, RestoreToActive, ConflictSkipExisting, "target", 7, 9)
	if err == nil || !strings.Contains(err.Error(), "1 archived logs failed") {
		t.Fatalf("expected partial restore error, got %v", err)
	}
	if result.Restored != 1 || result.Failed != 1 {
		t.Fatalf("partial result was lost: %#v", result)
	}
}

func TestEnsureRestoreIndexPreservesExistingPartialIndex(t *testing.T) {
	settingsUpdated := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		switch {
		case r.Method == http.MethodHead && r.URL.Path == "/partial-target":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPut && r.URL.Path == "/partial-target/_settings":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode settings body: %v", err)
			}
			settingsUpdated = body["index.mapping.total_fields.limit"] == float64(10000)
			_, _ = w.Write([]byte(`{"acknowledged":true}`))
		default:
			http.Error(w, "unexpected request", http.StatusBadRequest)
		}
	}))
	defer server.Close()

	client, err := elasticsearch.NewClient(elasticsearch.Config{Addresses: []string{server.URL}})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	if err := ensureRestoreIndex(t.Context(), client, "partial-target"); err != nil {
		t.Fatalf("prepare existing index: %v", err)
	}
	if !settingsUpdated {
		t.Fatal("existing partial index was not preserved with the compatibility field limit")
	}
}

func writeRestoreArchive(t *testing.T, logs []archiveLogLine) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "restore.ndjson.gz")
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}
	gz := gzip.NewWriter(file)
	encoder := json.NewEncoder(gz)
	day := time.Date(2026, time.August, 14, 0, 0, 0, 0, time.UTC)
	manifest := ArchiveManifest{Type: "manifest", Version: 1, ProductID: 7, EnvironmentID: 9, DateFrom: day, DateTo: day.AddDate(0, 0, 1)}
	if err := encoder.Encode(manifest); err != nil {
		t.Fatalf("write opening manifest: %v", err)
	}
	for _, log := range logs {
		if err := encoder.Encode(log); err != nil {
			t.Fatalf("write log: %v", err)
		}
	}
	manifest.DocumentCount = int64(len(logs))
	if err := encoder.Encode(manifest); err != nil {
		t.Fatalf("write closing manifest: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("close gzip: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close archive: %v", err)
	}
	if !strings.HasSuffix(path, ".gz") {
		t.Fatal("test archive must be gzip")
	}
	return path
}

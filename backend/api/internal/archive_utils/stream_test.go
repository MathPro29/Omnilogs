package archive_utils

import (
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

func TestStreamArchiveWritesVerifiableNDJSONGZIP(t *testing.T) {
	searchCalls := 0
	scrollCalls := 0
	clearCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/_search"):
			searchCalls++
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode search request: %v", err)
			}
			if body["size"] != float64(archivePageSize) {
				t.Fatalf("unexpected page size: %#v", body["size"])
			}
			_, _ = w.Write([]byte(`{"_scroll_id":"scroll-1","hits":{"hits":[{"_id":"log-1","_index":"logs-2026.01.01","_source":{"product_id":7,"environment_id":9,"message":"hello"}}]}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/_search/scroll":
			scrollCalls++
			_, _ = w.Write([]byte(`{"_scroll_id":"scroll-2","hits":{"hits":[]}}`))
		case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "scroll"):
			clearCalls++
			_, _ = w.Write([]byte(`{"succeeded":true}`))
		default:
			http.Error(w, "unexpected request", http.StatusBadRequest)
		}
	}))
	defer server.Close()

	client, err := elasticsearch.NewClient(elasticsearch.Config{Addresses: []string{server.URL}})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	day := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	path := filepath.Join(t.TempDir(), "daily.ndjson.gz")
	result, err := StreamArchive(StreamArchiveRequest{
		Client: client, Context: t.Context(), Indices: []string{"logs-2026.01.01"},
		Query: map[string]any{"query": map[string]any{"bool": map[string]any{"filter": []any{
			map[string]any{"term": map[string]any{"product_id": 7}},
			map[string]any{"term": map[string]any{"environment_id": 9}},
		}}}},
		ProductID: 7, EnvironmentID: 9, DateFrom: day, DateTo: day.AddDate(0, 0, 1), FilePath: path,
	})
	if err != nil {
		t.Fatalf("stream archive: %v", err)
	}
	if result.DocumentCount != 1 || result.OriginalBytes == 0 || result.CompressedBytes == 0 || result.Checksum == "" {
		t.Fatalf("unexpected archive result: %#v", result)
	}
	manifest, count, err := VerifyArchive(path, 7, 9)
	if err != nil {
		t.Fatalf("verify archive: %v", err)
	}
	if count != 1 || manifest.DocumentCount != 1 || manifest.OriginalBytes != result.OriginalBytes {
		t.Fatalf("unexpected verified manifest: %#v count=%d", manifest, count)
	}
	if searchCalls != 1 || scrollCalls != 1 || clearCalls != 1 {
		t.Fatalf("unexpected scroll lifecycle: search=%d scroll=%d clear=%d", searchCalls, scrollCalls, clearCalls)
	}
}

func TestStreamArchiveDoesNotCreateManifestOnlyFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/_search") {
			_, _ = w.Write([]byte(`{"_scroll_id":"scroll-empty","hits":{"hits":[]}}`))
			return
		}
		if r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "scroll") {
			_, _ = w.Write([]byte(`{"succeeded":true}`))
			return
		}
		http.Error(w, "unexpected request", http.StatusBadRequest)
	}))
	defer server.Close()
	client, err := elasticsearch.NewClient(elasticsearch.Config{Addresses: []string{server.URL}})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	path := filepath.Join(t.TempDir(), "empty.ndjson.gz")
	result, err := StreamArchive(StreamArchiveRequest{
		Client: client, Context: t.Context(), Indices: []string{"logs-2026.01.01"},
		Query: map[string]any{"query": map[string]any{"bool": map[string]any{"filter": []any{
			map[string]any{"term": map[string]any{"product_id": 7}},
			map[string]any{"term": map[string]any{"environment_id": 9}},
		}}}}, ProductID: 7, EnvironmentID: 9, FilePath: path,
	})
	if err != nil {
		t.Fatalf("stream empty archive: %v", err)
	}
	if !result.Empty || result.DocumentCount != 0 {
		t.Fatalf("empty archive result was not reported: %#v", result)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("manifest-only archive file still exists: %v", err)
	}
}

func TestStreamArchiveRequiresProductAndEnvironmentQueryGuards(t *testing.T) {
	_, err := StreamArchive(StreamArchiveRequest{
		Client:        &elasticsearch.Client{},
		Context:       t.Context(),
		Indices:       []string{"logs-2026.01.01"},
		Query:         map[string]any{"query": map[string]any{"term": map[string]any{"product_id": 7}}},
		ProductID:     7,
		EnvironmentID: 9,
		FilePath:      filepath.Join(t.TempDir(), "archive.ndjson.gz"),
	})
	if err == nil || !strings.Contains(err.Error(), "product_id and environment_id") {
		t.Fatalf("expected scope guard error, got %v", err)
	}
}

func TestBuildScopedDailyArchiveVersionPathDoesNotOverwritePreviousGeneration(t *testing.T) {
	day := time.Date(2026, time.August, 19, 0, 0, 0, 0, time.UTC)
	first := BuildScopedDailyArchiveVersionPath(t.TempDir(), 7, 9, nil, nil, nil, day, "archive-one")
	second := BuildScopedDailyArchiveVersionPath(t.TempDir(), 7, 9, nil, nil, nil, day, "archive-two")
	if first == second {
		t.Fatal("archive generations must have different file paths")
	}
	if !strings.HasSuffix(first, "19-archive-one.ndjson.gz") || !strings.HasSuffix(second, "19-archive-two.ndjson.gz") {
		t.Fatalf("unexpected versioned paths: %q %q", first, second)
	}
}

package archive_utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
)

func TestFetchArchivedLogsScrollsAndScopesEnvironment(t *testing.T) {
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
				t.Fatalf("decode search body: %v", err)
			}
			query := body["query"].(map[string]any)
			term := query["term"].(map[string]any)
			if term["environment_id"] != float64(9) {
				t.Fatalf("search is not environment-scoped: %#v", body)
			}
			_, _ = w.Write([]byte(`{"_scroll_id":"scroll-1","hits":{"hits":[{"_id":"log-1","_source":{"product_id":7}}]}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/_search/scroll":
			scrollCalls++
			_, _ = w.Write([]byte(`{"_scroll_id":"scroll-2","hits":{"hits":[]}}`))
		case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "scroll"):
			clearCalls++
			_, _ = w.Write([]byte(`{"succeeded":true,"num_freed":1}`))
		default:
			http.Error(w, "unexpected request", http.StatusBadRequest)
		}
	}))
	defer server.Close()

	client, err := elasticsearch.NewClient(elasticsearch.Config{Addresses: []string{server.URL}})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	environmentID := 9
	logs, err := fetchArchivedLogs(t.Context(), client, "logs-index", &environmentID)
	if err != nil {
		t.Fatalf("fetch archive logs: %v", err)
	}
	if len(logs) != 1 || logs[0]["_id"] != "log-1" {
		t.Fatalf("unexpected logs: %#v", logs)
	}
	if searchCalls != 1 || scrollCalls != 1 || clearCalls != 1 {
		t.Fatalf("unexpected lifecycle: search=%d scroll=%d clear=%d", searchCalls, scrollCalls, clearCalls)
	}
}

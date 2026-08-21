package archive_utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

const archiveScrollTTL = 2 * time.Minute

type scrollPage struct {
	ScrollID string `json:"_scroll_id"`
	Hits     struct {
		Hits []struct {
			ID     string         `json:"_id"`
			Source map[string]any `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func fetchArchivedLogs(ctx context.Context, client *elasticsearch.Client, indexName string, environmentID *int) ([]map[string]any, error) {
	if environmentID == nil || *environmentID <= 0 {
		return nil, fmt.Errorf("archive search requires environment_id guard")
	}
	query := map[string]any{
		"query": map[string]any{"term": map[string]any{"environment_id": *environmentID}},
		"sort":  []string{"_doc"},
		"size":  1000,
	}
	body, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	response, err := client.Search(
		client.Search.WithContext(ctx),
		client.Search.WithIndex(indexName),
		client.Search.WithBody(bytes.NewReader(body)),
		client.Search.WithScroll(archiveScrollTTL),
	)
	if err != nil {
		return nil, err
	}
	if response.StatusCode == http.StatusNotFound {
		response.Body.Close()
		return []map[string]any{}, nil
	}
	if response.IsError() {
		defer response.Body.Close()
		return nil, fmt.Errorf("elasticsearch archive search returned status %s", response.Status())
	}

	logs := make([]map[string]any, 0, 1000)
	var scrollID string
	defer func() {
		if scrollID == "" {
			return
		}
		clearContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		clearResponse, clearErr := client.ClearScroll(
			client.ClearScroll.WithContext(clearContext),
			client.ClearScroll.WithScrollID(scrollID),
		)
		if clearErr == nil {
			clearResponse.Body.Close()
		}
	}()

	for {
		var page scrollPage
		decodeErr := json.NewDecoder(response.Body).Decode(&page)
		response.Body.Close()
		if decodeErr != nil {
			return nil, decodeErr
		}
		scrollID = page.ScrollID
		if len(page.Hits.Hits) == 0 {
			return logs, nil
		}
		for _, hit := range page.Hits.Hits {
			logs = append(logs, map[string]any{"_id": hit.ID, "_source": hit.Source})
		}

		response, err = client.Scroll(
			client.Scroll.WithContext(ctx),
			client.Scroll.WithScrollID(scrollID),
			client.Scroll.WithScroll(archiveScrollTTL),
		)
		if err != nil {
			return nil, err
		}
		if response.IsError() {
			defer response.Body.Close()
			return nil, fmt.Errorf("elasticsearch archive scroll returned status %s", response.Status())
		}
	}
}

// DeleteArchivedLogs is kept only as a compatibility symbol. The old
// signature cannot prove product, environment, and timestamp guards together,
// so it is deliberately disabled. Use retention_policy/provider.DeleteActive.
func DeleteArchivedLogs(ctx context.Context, client *elasticsearch.Client, indexName string, environmentID *int) error {
	_ = ctx
	_ = client
	_ = indexName
	_ = environmentID
	return fmt.Errorf("unsafe legacy archive delete is disabled; use the verified retention delete flow")
}

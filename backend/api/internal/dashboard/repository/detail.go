package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (r *repository) GetLogDetail(ctx context.Context, indexName, logID string) (map[string]any, error) {
	res, err := r.esClient.Get(
		indexName,
		logID,
		r.esClient.Get.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch error: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("log not found") // You can handle custom errors if needed
		}
		return nil, fmt.Errorf("elasticsearch failed: %s", res.String())
	}

	var doc map[string]any
	if err := json.NewDecoder(res.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("failed to parse document: %w", err)
	}

	return doc, nil
}

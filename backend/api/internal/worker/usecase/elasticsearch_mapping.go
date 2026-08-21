package usecase

import (
	"bytes"
	"context"
	"fmt"

	"omnilogs-api/internal/searchindex"
)

func (u *usecase) ensureControlledMappings(ctx context.Context, entries []processedLog) error {
	indexes := make(map[string]struct{}, len(entries))
	for i := range entries {
		indexes[entries[i].indexName] = struct{}{}
	}
	for indexName := range indexes {
		existsResponse, err := u.esClient.Indices.Exists(
			[]string{indexName},
			u.esClient.Indices.Exists.WithContext(ctx),
		)
		if err != nil {
			return fmt.Errorf("check controlled mapping for %s: %w", indexName, err)
		}
		exists := existsResponse.StatusCode == 200
		existsResponse.Body.Close()
		if !exists {
			body := append([]byte(`{"mappings":`), []byte(searchindex.ControlledLogMapping)...)
			body = append(body, '}')
			createResponse, createErr := u.esClient.Indices.Create(
				indexName,
				u.esClient.Indices.Create.WithBody(bytes.NewReader(body)),
				u.esClient.Indices.Create.WithContext(ctx),
			)
			if createErr != nil {
				return fmt.Errorf("create controlled index %s: %w", indexName, createErr)
			}
			created := !createResponse.IsError()
			createResponse.Body.Close()
			if created {
				continue
			}
		}
		response, err := u.esClient.Indices.PutMapping(
			[]string{indexName},
			bytes.NewReader([]byte(searchindex.ControlledLogMapping)),
			u.esClient.Indices.PutMapping.WithContext(ctx),
		)
		if err != nil {
			return fmt.Errorf("apply controlled mapping to %s: %w", indexName, err)
		}
		if response != nil {
			response.Body.Close()
			if response.IsError() && response.StatusCode != 404 {
				return fmt.Errorf("apply controlled mapping to %s: status %d", indexName, response.StatusCode)
			}
		}
	}
	return nil
}

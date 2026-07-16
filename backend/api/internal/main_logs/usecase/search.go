package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	auditusecase "omnilogs-api/internal/audit_logs/usecase"
	"omnilogs-api/internal/main_logs/document"
	"omnilogs-api/models"
)

func (u *usecase) Search(ctx context.Context, input SearchInput, requestID, traceID, ipAddress, userAgent *string) (*SearchResult, error) {
	if err := u.ensureProductAccess(ctx, input.ActorUserID, input.PlatformAdmin, input.ProductID); err != nil {
		_ = u.auditUsecase.Record(ctx, auditusecase.RecordAuditInput{
			ActorUserID: uintToInt64Ptr(input.ActorUserID), ProductID: &input.ProductID,
			Action: models.AuditActionSearchLogs, ResourceType: models.AuditResourceTypeMainLog,
			RequestID: requestID, TraceID: traceID, Result: models.AuditResultDenied,
			IPAddress: ipAddress, UserAgent: userAgent,
		})
		return nil, err
	}

	queryBody, err := json.Marshal(buildSearchQuery(input))
	if err != nil {
		return nil, err
	}
	response, err := u.repo.SearchLogs(ctx, u.resolveSearchIndices(ctx, input.ProductID), queryBody)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.IsError() {
		return nil, fmt.Errorf("elasticsearch returned status %d", response.StatusCode)
	}

	var payload struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				ID     string         `json:"_id"`
				Source map[string]any `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}
	result := &SearchResult{Items: make([]MainLogDocument, 0, len(payload.Hits.Hits)), Total: payload.Hits.Total.Value}
	for _, hit := range payload.Hits.Hits {
		result.Items = append(result.Items, document.Map(hit.ID, hit.Source))
	}
	return result, nil
}

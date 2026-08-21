package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
	response, err := u.repo.SearchLogs(ctx, u.resolveSearchIndices(ctx, input.ProductID, input.EnvironmentID), queryBody)
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

func (u *usecase) SearchV2(ctx context.Context, input DynamicSearchInput, requestID, traceID, ipAddress, userAgent *string) (*SearchResult, error) {
	if err := u.ensureProductAccess(ctx, input.ActorUserID, input.PlatformAdmin, input.Scope.ProductID); err != nil {
		_ = u.auditUsecase.Record(ctx, auditusecase.RecordAuditInput{
			ActorUserID: uintToInt64Ptr(input.ActorUserID), ProductID: &input.Scope.ProductID,
			Action: models.AuditActionSearchLogs, ResourceType: models.AuditResourceTypeMainLog,
			RequestID: requestID, TraceID: traceID, Result: models.AuditResultDenied,
			IPAddress: ipAddress, UserAgent: userAgent,
		})
		return nil, err
	}
	queryBody, err := json.Marshal(buildDynamicSearchQuery(input))
	if err != nil {
		return nil, err
	}
	indices, err := u.resolveDynamicSearchIndices(ctx, input)
	if err != nil {
		return nil, err
	}
	response, err := u.repo.SearchLogs(ctx, indices, queryBody)
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

func (u *usecase) resolveDynamicSearchIndices(ctx context.Context, input DynamicSearchInput) ([]string, error) {
	archiveID := strings.TrimSpace(input.Scope.ArchiveID)
	if archiveID == "" {
		return u.resolveSearchIndices(ctx, input.Scope.ProductID, input.Scope.EnvironmentID), nil
	}
	pattern, err := u.repo.GetRestoredArchiveIndexPattern(ctx, input.Scope.ProductID, input.Scope.EnvironmentID, archiveID)
	if err != nil || strings.TrimSpace(pattern) == "" {
		return nil, ErrRestoredArchiveUnavailable
	}
	return []string{pattern}, nil
}

package usecase

import (
	"context"

	"omnilogs-api/dto"
	"omnilogs-api/internal/dashboard/repository"
	"omnilogs-api/models"
)

type Usecase interface {
	GetLogs(ctx context.Context, query dto.LogQuery) (map[string]any, error)
	GetLogStats(ctx context.Context, query dto.LogQuery) (map[string]any, error)
	GetLogDetail(ctx context.Context, indexName, logID string) (map[string]any, error)
	GetAuditLogs(ctx context.Context, query dto.AuditLogQuery) ([]models.SystemAuditLog, int64, error)
}

type usecase struct {
	repo repository.Repository
}

func NewUsecase(repo repository.Repository) Usecase {
	return &usecase{
		repo: repo,
	}
}

func (u *usecase) GetLogs(ctx context.Context, query dto.LogQuery) (map[string]any, error) {
	searchResult, err := u.repo.SearchLogs(ctx, query)
	if err != nil {
		return nil, err
	}

	hits, _ := searchResult["hits"].(map[string]any)
	totalHitsMap, _ := hits["total"].(map[string]any)
	totalLogs := 0
	if totalHitsMap != nil {
		if v, ok := totalHitsMap["value"].(float64); ok {
			totalLogs = int(v)
		}
	}

	hitsList, _ := hits["hits"].([]any)
	logs := make([]any, 0)
	for _, item := range hitsList {
		if hitMap, ok := item.(map[string]any); ok {
			source, _ := hitMap["_source"].(map[string]any)
			if source != nil {
				// แนบ document ID และ index กลับไปด้วยสำหรับการดูรายละเอียดเดี่ยวๆ
				source["_id"] = hitMap["_id"]
				source["_index"] = hitMap["_index"]
				logs = append(logs, source)
			}
		}
	}

	return map[string]any{
		"total":  totalLogs,
		"limit":  query.Limit,
		"offset": query.Offset,
		"data":   logs,
	}, nil
}

func (u *usecase) GetLogStats(ctx context.Context, query dto.LogQuery) (map[string]any, error) {
	searchResult, err := u.repo.GetLogStats(ctx, query)
	if err != nil {
		return nil, err
	}
	
	aggs, _ := searchResult["aggregations"].(map[string]any)
	return aggs, nil
}

func (u *usecase) GetLogDetail(ctx context.Context, indexName, logID string) (map[string]any, error) {
	doc, err := u.repo.GetLogDetail(ctx, indexName, logID)
	if err != nil {
		return nil, err
	}
	
	source, _ := doc["_source"].(map[string]any)
	return source, nil
}

func (u *usecase) GetAuditLogs(ctx context.Context, query dto.AuditLogQuery) ([]models.SystemAuditLog, int64, error) {
	return u.repo.GetAuditLogs(ctx, query)
}

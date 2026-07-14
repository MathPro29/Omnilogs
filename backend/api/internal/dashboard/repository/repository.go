package repository

import (
	"context"
	"time"

	"omnilogs-api/dto"
	"omnilogs-api/models"

	"github.com/elastic/go-elasticsearch/v8"
	"gorm.io/gorm"
)

type Repository interface {
	SearchLogs(ctx context.Context, query dto.LogQuery) (map[string]any, error)
	GetLogStats(ctx context.Context, query dto.LogQuery) (map[string]any, error)
	GetLogDetail(ctx context.Context, indexName, logID string) (map[string]any, error)
	GetAuditLogs(ctx context.Context, query dto.AuditLogQuery) ([]models.SystemAuditLog, int64, error)
}

type repository struct {
	db                 *gorm.DB
	esClient           *elasticsearch.Client
	queryTimeout       time.Duration
	slowQueryThreshold time.Duration
}

func NewRepository(db *gorm.DB, esClient *elasticsearch.Client, queryTimeout, slowQueryThreshold time.Duration) Repository {
	if queryTimeout <= 0 {
		queryTimeout = 10 * time.Second
	}
	if slowQueryThreshold <= 0 {
		slowQueryThreshold = time.Second
	}
	return &repository{
		db:                 db,
		esClient:           esClient,
		queryTimeout:       queryTimeout,
		slowQueryThreshold: slowQueryThreshold,
	}
}

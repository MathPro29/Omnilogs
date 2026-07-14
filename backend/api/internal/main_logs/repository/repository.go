package repository

import (
	"bytes"
	"context"
	"log/slog"
	"time"

	"omnilogs-api/models"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"gorm.io/gorm"
)

type Repository interface {
	SearchLogs(ctx context.Context, indices []string, queryBody []byte) (*esapi.Response, error)
	GetLogIndexRef(ctx context.Context, productID int64, logID string) (*models.LogIndexRef, error)
	GetMainLogFromES(ctx context.Context, index, docID string) (*esapi.Response, error)
	GetPostgresPayload(ctx context.Context, logID string) (*models.LogObjectStorageRef, error)
	HasProductMembership(ctx context.Context, userID uint, productID int64) (bool, error)
	GetActiveIndexPolicies(ctx context.Context, productID int64) ([]models.ElasticIndexPolicy, error)
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

func (r *repository) SearchLogs(ctx context.Context, indices []string, queryBody []byte) (*esapi.Response, error) {
	startedAt := time.Now()
	defer func() {
		duration := time.Since(startedAt)
		if duration >= r.slowQueryThreshold {
			slog.Warn("slow elasticsearch query",
				"operation", "main_log_search",
				"duration_ms", duration.Milliseconds(),
				"slow_threshold_ms", r.slowQueryThreshold.Milliseconds(),
				"index_count", len(indices),
			)
		}
	}()
	return r.esClient.Search(
		r.esClient.Search.WithContext(ctx),
		r.esClient.Search.WithIndex(indices...),
		r.esClient.Search.WithBody(bytes.NewReader(queryBody)),
		r.esClient.Search.WithTrackTotalHits(true),
		r.esClient.Search.WithTimeout(r.queryTimeout),
	)
}

func (r *repository) GetLogIndexRef(ctx context.Context, productID int64, logID string) (*models.LogIndexRef, error) {
	var ref models.LogIndexRef
	err := r.db.WithContext(ctx).Where("log_id = ? AND product_id = ?", logID, productID).First(&ref).Error
	return &ref, err
}

func (r *repository) GetMainLogFromES(ctx context.Context, index, docID string) (*esapi.Response, error) {
	return r.esClient.Get(index, docID, r.esClient.Get.WithContext(ctx))
}

func (r *repository) GetPostgresPayload(ctx context.Context, logID string) (*models.LogObjectStorageRef, error) {
	var objectRef models.LogObjectStorageRef
	err := r.db.WithContext(ctx).
		Where("log_id = ? AND object_type = ?", logID, "INPUT_PAYLOAD").
		Order("created_at DESC NULLS LAST").
		First(&objectRef).Error
	return &objectRef, err
}

func (r *repository) HasProductMembership(ctx context.Context, userID uint, productID int64) (bool, error) {
	var membershipID int
	err := r.db.WithContext(ctx).
		Model(&models.ProductMembership{}).
		Select("membership_id").
		Where("user_id = ? AND product_id = ? AND is_active = TRUE", userID, productID).
		Limit(1).
		Scan(&membershipID).Error
	return membershipID != 0, err
}

func (r *repository) GetActiveIndexPolicies(ctx context.Context, productID int64) ([]models.ElasticIndexPolicy, error) {
	var policies []models.ElasticIndexPolicy
	err := r.db.WithContext(ctx).
		Select("index_prefix").
		Where("product_id = ? AND is_active = TRUE", productID).
		Find(&policies).Error
	return policies, err
}

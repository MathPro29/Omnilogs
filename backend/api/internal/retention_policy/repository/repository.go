package repository

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"omnilogs-api/dto"
	retentionprovider "omnilogs-api/internal/retention_policy/provider"
	"omnilogs-api/models"

	"github.com/elastic/go-elasticsearch/v8"
	"gorm.io/gorm"
)

var ErrActivePolicyExists = errors.New("an active retention policy already exists for this product and environment; update or deactivate it before creating another active policy")

type Repository interface {
	Create(ctx context.Context, policy *models.LogRetentionPolicy) error
	Update(ctx context.Context, policy *models.LogRetentionPolicy) error
	Delete(ctx context.Context, policyID int, productID int) error
	GetByID(ctx context.Context, policyID int, productID int) (*models.LogRetentionPolicy, error)
	List(ctx context.Context, productID int, environmentID *int) ([]models.LogRetentionPolicy, error)
	ToggleActive(ctx context.Context, policyID int, productID int, isActive bool) error
	Stats(ctx context.Context, productID int, environmentID *int) (PolicyStats, error)
	EnvironmentBelongsToProduct(ctx context.Context, productID, environmentID int) (bool, error)
	Historical(ctx context.Context, req dto.RetentionHistoricalRequest) (*dto.RetentionHistoricalResponse, error)
	CountCandidates(ctx context.Context, req dto.RetentionHistoricalRequest) (int64, error)
}

type PolicyStats struct {
	TotalPolicies           int
	ActivePolicies          int
	TotalStorageBytes       int64
	ActiveStorageBytes      int64
	ArchiveGZIPStorageBytes int64
	ReadyZIPStorageBytes    int64
	ScheduledPurges24h      int
	TotalFoldersCount       int
}

type repository struct {
	db       *gorm.DB
	provider *retentionprovider.Provider
}

func NewRepository(db *gorm.DB, clients ...*elasticsearch.Client) Repository {
	var client *elasticsearch.Client
	if len(clients) > 0 {
		client = clients[0]
	}
	return &repository{db: db, provider: retentionprovider.New(db, client)}
}

func (r *repository) Create(ctx context.Context, policy *models.LogRetentionPolicy) error {
	err := r.db.WithContext(ctx).Create(policy).Error
	return translatePolicyWriteError(err)
}

func (r *repository) Update(ctx context.Context, policy *models.LogRetentionPolicy) error {
	// Save is intentional here: the archive flow must persist zero values too
	// (for example, turning Archive off clears the legacy retention fields and
	// permanently keeps DeleteActiveAfterArchive=false).
	err := r.db.WithContext(ctx).Save(policy).Error
	return translatePolicyWriteError(err)
}

func (r *repository) Delete(ctx context.Context, policyID int, productID int) error {
	return r.db.WithContext(ctx).
		Where("policy_id = ? AND product_id = ?", policyID, productID).
		Delete(&models.LogRetentionPolicy{}).Error
}

func (r *repository) GetByID(ctx context.Context, policyID int, productID int) (*models.LogRetentionPolicy, error) {
	var policy models.LogRetentionPolicy
	err := r.db.WithContext(ctx).
		Where("policy_id = ? AND product_id = ?", policyID, productID).
		First(&policy).Error
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

func (r *repository) List(ctx context.Context, productID int, environmentID *int) ([]models.LogRetentionPolicy, error) {
	var policies []models.LogRetentionPolicy
	query := r.db.WithContext(ctx).Where("product_id = ?", productID)
	if environmentID != nil && *environmentID > 0 {
		query = query.Where("environment_id = ?", *environmentID)
	}
	err := query.Order("created_at DESC").Find(&policies).Error
	return policies, err
}

func (r *repository) ToggleActive(ctx context.Context, policyID int, productID int, isActive bool) error {
	err := r.db.WithContext(ctx).
		Model(&models.LogRetentionPolicy{}).
		Where("policy_id = ? AND product_id = ?", policyID, productID).
		Update("is_active", isActive).Error
	return translatePolicyWriteError(err)
}

func translatePolicyWriteError(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(strings.ToLower(err.Error()), "uq_retention_policy_product_environment_active") {
		return ErrActivePolicyExists
	}
	return err
}

func (r *repository) Stats(ctx context.Context, productID int, environmentID *int) (PolicyStats, error) {
	policyQuery := r.db.WithContext(ctx).Model(&models.LogRetentionPolicy{}).Where("product_id = ?", productID)
	archiveQuery := r.db.WithContext(ctx).Model(&models.LogArchive{}).Where("product_id = ?", productID)
	downloadQuery := r.db.WithContext(ctx).Model(&models.LogArchiveDownload{}).Where("product_id = ?", productID)
	if environmentID != nil && *environmentID > 0 {
		policyQuery = policyQuery.Where("environment_id = ?", *environmentID)
		archiveQuery = archiveQuery.Where("environment_id = ?", *environmentID)
		downloadQuery = downloadQuery.Where("environment_id = ?", *environmentID)
	}

	var total, active int64
	if err := policyQuery.Count(&total).Error; err != nil {
		return PolicyStats{}, err
	}
	if err := policyQuery.Where("is_active = TRUE").Count(&active).Error; err != nil {
		return PolicyStats{}, err
	}

	var scheduled int64
	now := time.Now().UTC()
	if err := policyQuery.Where("is_active = TRUE AND next_purge_at > ? AND next_purge_at <= ?", now, now.Add(24*time.Hour)).Count(&scheduled).Error; err != nil {
		return PolicyStats{}, err
	}

	var archives []models.LogArchive
	if err := archiveQuery.Where("file_path IS NOT NULL").Find(&archives).Error; err != nil {
		return PolicyStats{}, err
	}
	archivePaths := make([]string, 0, len(archives))
	for _, archive := range archives {
		if archive.FilePath != nil {
			archivePaths = append(archivePaths, *archive.FilePath)
		}
	}
	archiveBytes, archiveFiles := actualFileUsage(archivePaths)

	var downloads []models.LogArchiveDownload
	if err := downloadQuery.Where("file_path <> ''").Find(&downloads).Error; err != nil {
		return PolicyStats{}, err
	}
	downloadPaths := make([]string, 0, len(downloads))
	for _, download := range downloads {
		downloadPaths = append(downloadPaths, download.FilePath)
	}
	readyZIPBytes, _ := actualFileUsage(downloadPaths)

	var activeBytes int64
	if environmentID != nil && *environmentID > 0 && r.provider != nil {
		value, storageErr := r.provider.StorageBytes(ctx, productID, *environmentID)
		if storageErr != nil {
			return PolicyStats{}, storageErr
		}
		activeBytes = value
	}

	return PolicyStats{
		TotalPolicies:           int(total),
		ActivePolicies:          int(active),
		TotalStorageBytes:       activeBytes + archiveBytes + readyZIPBytes,
		ActiveStorageBytes:      activeBytes,
		ArchiveGZIPStorageBytes: archiveBytes,
		ReadyZIPStorageBytes:    readyZIPBytes,
		ScheduledPurges24h:      int(scheduled),
		TotalFoldersCount:       archiveFiles,
	}, nil
}

func actualFileUsage(paths []string) (int64, int) {
	seen := make(map[string]struct{}, len(paths))
	var bytes int64
	var files int
	for _, path := range paths {
		cleaned := filepath.Clean(path)
		if cleaned == "." || cleaned == "" {
			continue
		}
		if _, exists := seen[cleaned]; exists {
			continue
		}
		seen[cleaned] = struct{}{}
		info, err := os.Stat(cleaned)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		bytes += info.Size()
		files++
	}
	return bytes, files
}

func (r *repository) EnvironmentBelongsToProduct(ctx context.Context, productID, environmentID int) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.ProductEnvironment{}).
		Where("product_id = ? AND environment_id = ? AND is_active = TRUE", productID, environmentID).Count(&count).Error; err != nil {
		return false, err
	}
	return count == 1, nil
}

func (r *repository) Historical(ctx context.Context, req dto.RetentionHistoricalRequest) (*dto.RetentionHistoricalResponse, error) {
	return r.provider.Historical(ctx, req)
}

func (r *repository) CountCandidates(ctx context.Context, req dto.RetentionHistoricalRequest) (int64, error) {
	from, to := req.DateFrom, req.DateTo
	if from == nil || to == nil {
		return 0, nil
	}
	return r.provider.Count(ctx, req.ProductID, req.EnvironmentID, from.UTC(), to.UTC(), req.CategoryID, req.FeatureID, req.SubFeatureID)
}

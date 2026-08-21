package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"omnilogs-api/dto"
	"omnilogs-api/internal/archive_utils"
	"omnilogs-api/models"

	"github.com/elastic/go-elasticsearch/v8"
	"gorm.io/gorm"
)

type Repository interface {
	Create(policy *models.ElasticIndexPolicy) error
	GetByID(id int) (*models.ElasticIndexPolicy, error)
	Update(policy *models.ElasticIndexPolicy) error
	Delete(id int) error
	List(productID int, envID *int) ([]models.ElasticIndexPolicy, error)
	PushToArchives(req dto.PushToArchivesRequest) (*dto.PushToArchivesResponse, error)
	ClearAllLogs(productID int) error
}

type repository struct {
	db       *gorm.DB
	esClient *elasticsearch.Client
}

func NewRepository(db *gorm.DB, clients ...*elasticsearch.Client) Repository {
	var client *elasticsearch.Client
	if len(clients) > 0 {
		client = clients[0]
	}
	return &repository{db: db, esClient: client}
}

func (r *repository) Create(policy *models.ElasticIndexPolicy) error {
	return r.db.Create(policy).Error
}

func (r *repository) GetByID(id int) (*models.ElasticIndexPolicy, error) {
	var policy models.ElasticIndexPolicy
	err := r.db.First(&policy, "elastic_policy_id = ?", id).Error
	return &policy, err
}

func (r *repository) Update(policy *models.ElasticIndexPolicy) error {
	return r.db.Save(policy).Error
}

func (r *repository) Delete(id int) error {
	return r.db.Delete(&models.ElasticIndexPolicy{}, "elastic_policy_id = ?", id).Error
}

func (r *repository) List(productID int, envID *int) ([]models.ElasticIndexPolicy, error) {
	var list []models.ElasticIndexPolicy
	db := r.db
	if productID > 0 {
		db = db.Where("product_id = ?", productID)
	}
	if envID != nil {
		db = db.Where("environment_id = ?", *envID)
	}
	err := db.Find(&list).Error
	return list, err
}

func (r *repository) PushToArchives(req dto.PushToArchivesRequest) (*dto.PushToArchivesResponse, error) {
	if r.esClient == nil {
		return nil, fmt.Errorf("elasticsearch client is not configured")
	}
	policy, err := r.GetByID(req.ElasticPolicyID)
	if err != nil {
		return nil, err
	}
	if policy.ProductID != req.ProductID {
		return nil, fmt.Errorf("policy does not belong to product")
	}

	from, err := time.Parse("2006-01-02", req.FromDate)
	if err != nil {
		return nil, fmt.Errorf("invalid from_date: use YYYY-MM-DD")
	}
	to, err := time.Parse("2006-01-02", req.ToDate)
	if err != nil {
		return nil, fmt.Errorf("invalid to_date: use YYYY-MM-DD")
	}
	if to.Before(from) {
		return nil, fmt.Errorf("to_date must not be before from_date")
	}

	prefix := strings.ToLower(strings.TrimSpace(policy.IndexPrefix))
	prefix = strings.ReplaceAll(prefix, "_", "-")
	prefix = strings.ReplaceAll(prefix, " ", "-")

	targets := []string{prefix + "-*"}
	defaultPrefix := fmt.Sprintf("omnilogs-product-%d", policy.ProductID)
	if policy.EnvironmentID != nil {
		defaultEnvPrefix := fmt.Sprintf("%s-env-%d", defaultPrefix, *policy.EnvironmentID)
		if prefix != defaultEnvPrefix {
			targets = append(targets, defaultEnvPrefix+"-*")
		}
		// Also include product-wide fallback index to scan for environment logs stored there
		targets = append(targets, defaultPrefix+"-*")
	} else {
		if prefix != defaultPrefix {
			targets = append(targets, defaultPrefix+"-*")
		}
	}

	res, err := r.esClient.Indices.Get(targets, r.esClient.Indices.Get.WithContext(context.Background()))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			return &dto.PushToArchivesResponse{}, nil
		}
		return nil, fmt.Errorf("elasticsearch list indices failed: %s", res.Status())
	}
	var indices map[string]json.RawMessage
	if err := json.NewDecoder(res.Body).Decode(&indices); err != nil {
		return nil, err
	}

	result := &dto.PushToArchivesResponse{}
	for indexName := range indices {
		parts := strings.Split(indexName, "-")
		indexDate, parseErr := time.Parse("2006.01.02", parts[len(parts)-1])
		if parseErr != nil || indexDate.Before(from) || indexDate.After(to) {
			continue
		}

		// B. Check if there are active queue batches for this product and environment
		var pendingCount int64
		q := r.db.Model(&models.LogQueueBatch{}).Where("product_id = ? AND status IN ?", policy.ProductID, []string{"QUEUED", "CLAIMED", "PROCESSING"})
		if policy.EnvironmentID != nil {
			q = q.Where("environment_id = ?", *policy.EnvironmentID)
		} else {
			q = q.Where("environment_id IS NULL")
		}
		err = q.Count(&pendingCount).Error
		if err == nil && pendingCount > 0 {
			continue
		}

		var latestArchive models.LogArchive
		var archiveFound bool
		archiveQuery := r.db.Where("product_id = ? AND date_from = ?", policy.ProductID, indexDate)
		if policy.EnvironmentID != nil {
			archiveQuery = archiveQuery.Where("environment_id = ?", *policy.EnvironmentID)
		} else {
			archiveQuery = archiveQuery.Where("environment_id IS NULL")
		}
		archiveResult := archiveQuery.Order("created_at DESC").Limit(1).Find(&latestArchive)
		if archiveResult.Error != nil {
			return nil, archiveResult.Error
		}
		if archiveResult.RowsAffected > 0 {
			archiveFound = true
			if latestArchive.Status != nil && (*latestArchive.Status == "COMPLETED" || *latestArchive.Status == "ARCHIVED" || *latestArchive.Status == "VERIFIED") {
				continue
			}
		}

		// C. Lock Index as Read-only before archiving
		lockRes, lockErr := r.esClient.Indices.PutSettings(
			strings.NewReader(`{"index":{"blocks.write":true}}`),
			r.esClient.Indices.PutSettings.WithIndex(indexName),
		)
		if lockErr == nil {
			lockRes.Body.Close()
		}

		// Load Ingestion Policy for custom formats and paths
		var ingestPolicy models.LogIngestionPolicy
		archiveFormat := "JSON"
		archiveStoragePath := fmt.Sprintf("data/archives/product-%d", policy.ProductID)
		errPolicy := r.db.
			Where("product_id = ? AND (environment_id = ? OR environment_id IS NULL)", policy.ProductID, policy.EnvironmentID).
			Order("environment_id DESC NULLS LAST").
			Limit(1).
			Find(&ingestPolicy).Error
		if errPolicy == nil {
			if ingestPolicy.ArchiveFormat != nil && *ingestPolicy.ArchiveFormat != "" {
				archiveFormat = *ingestPolicy.ArchiveFormat
			}
			if ingestPolicy.ArchiveStoragePath != nil && *ingestPolicy.ArchiveStoragePath != "" {
				archiveStoragePath = *ingestPolicy.ArchiveStoragePath
			}
		}

		// 1. Archive actual Elasticsearch logs and PostgreSQL system_audit_logs
		totalLogs, compressedSize, localPath, archiveErr := archive_utils.ArchiveIndex(context.Background(), r.db, r.esClient, indexName, policy.ProductID, policy.EnvironmentID, archiveFormat, archiveStoragePath)
		if archiveErr != nil {
			// Unlock index if archiving fails
			unlockRes, unlockErr := r.esClient.Indices.PutSettings(
				strings.NewReader(`{"index":{"blocks.write":false}}`),
				r.esClient.Indices.PutSettings.WithIndex(indexName),
			)
			if unlockErr == nil {
				unlockRes.Body.Close()
			}
			result.FailedCount++
			continue
		}

		// 2. Create or Update the LogArchive database record
		now := time.Now().UTC()
		dateEnd := indexDate.Add(24*time.Hour - time.Second)
		// Legacy push-to-archives is retained for compatibility, but it is now
		// archive-only. Active logs can be deleted only by the verified daily
		// retention flow.
		status, provider, format := "ARCHIVED", "LOCAL", "JSON"

		var archive *models.LogArchive
		var dbErr error

		if archiveFound {
			// Update the existing record (e.g. from RESTORED back to COMPLETED)
			archive = &latestArchive
			dbErr = r.db.Model(archive).Updates(map[string]any{
				"status":                &status,
				"exported_at":           &now,
				"restored_at":           nil,
				"total_logs":            &totalLogs,
				"compressed_size_bytes": &compressedSize,
			}).Error
		} else {
			// Create a new record
			archive = &models.LogArchive{
				ArchiveID:           newUUID(),
				ProductID:           policy.ProductID,
				EnvironmentID:       policy.EnvironmentID,
				ArchiveYear:         indexDate.Year(),
				ArchiveMonth:        int(indexDate.Month()),
				DateFrom:            &indexDate,
				DateTo:              &dateEnd,
				StorageProvider:     provider,
				FilePath:            &localPath,
				FileFormat:          &format,
				TotalLogs:           &totalLogs,
				CompressedSizeBytes: &compressedSize,
				Status:              &status,
				ExportedAt:          &now,
			}
			dbErr = r.db.Create(archive).Error
		}

		if dbErr != nil {
			unlockIndex(r.esClient, indexName)
			result.FailedCount++
			continue
		}
		// Do not delete active logs here. This endpoint predates checksum
		// verification and must not bypass the VERIFIED gate.
		if err := unlockIndex(r.esClient, indexName); err != nil {
			result.FailedCount++
			failed := "FAILED"
			_ = r.db.Model(archive).Updates(map[string]any{"status": &failed}).Error
			continue
		}
		result.SuccessCount++
	}
	return result, nil
}

func unlockIndex(esClient *elasticsearch.Client, indexName string) error {
	res, err := esClient.Indices.PutSettings(
		strings.NewReader(`{"index":{"blocks.write":false}}`),
		esClient.Indices.PutSettings.WithIndex(indexName),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("elasticsearch unlock index failed: %s", res.Status())
	}
	return nil
}

func newUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "00000000-0000-4000-8000-000000000000"
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	d := make([]byte, 36)
	hex.Encode(d[0:8], b[0:4])
	d[8] = '-'
	hex.Encode(d[9:13], b[4:6])
	d[13] = '-'
	hex.Encode(d[14:18], b[6:8])
	d[18] = '-'
	hex.Encode(d[19:23], b[8:10])
	d[23] = '-'
	hex.Encode(d[24:], b[10:])
	return string(d)
}

func (r *repository) ClearAllLogs(productID int) error {
	// 1. Delete Postgres log records
	if err := r.db.Where("product_id = ?", productID).Delete(&models.LogArchive{}).Error; err != nil {
		return err
	}
	if err := r.db.Where("product_id = ?", productID).Delete(&models.LogObjectStorageRef{}).Error; err != nil {
		return err
	}
	if err := r.db.Where("product_id = ?", productID).Delete(&models.SystemAuditLog{}).Error; err != nil {
		return err
	}
	if err := r.db.Where("product_id = ?", productID).Delete(&models.LogIndexRef{}).Error; err != nil {
		return err
	}
	if err := r.db.Where("product_id = ?", productID).Delete(&models.LogFailure{}).Error; err != nil {
		return err
	}
	if err := r.db.Where("batch_id IN (SELECT batch_id FROM log_queue_batches WHERE product_id = ?)", productID).Delete(&models.LogQueueItem{}).Error; err != nil {
		return err
	}
	if err := r.db.Where("product_id = ?", productID).Delete(&models.LogQueueBatch{}).Error; err != nil {
		return err
	}
	if err := r.db.Where("product_id = ?", productID).Delete(&models.LogSensitiveFieldSecret{}).Error; err != nil {
		return err
	}

	// 2. Delete Elasticsearch indices
	if r.esClient != nil {
		var policies []models.ElasticIndexPolicy
		if err := r.db.Where("product_id = ?", productID).Find(&policies).Error; err != nil {
			return err
		}

		targets := []string{fmt.Sprintf("omnilogs-product-%d-*", productID)}
		for _, policy := range policies {
			if strings.TrimSpace(policy.IndexPrefix) != "" {
				prefix := strings.ToLower(strings.TrimSpace(policy.IndexPrefix))
				prefix = strings.ReplaceAll(prefix, "_", "-")
				prefix = strings.ReplaceAll(prefix, " ", "-")
				pattern := fmt.Sprintf("%s-*", prefix)
				found := false
				for _, t := range targets {
					if t == pattern {
						found = true
						break
					}
				}
				if !found {
					targets = append(targets, pattern)
				}
			}
		}

		ctx := context.Background()
		// Disable wildcard protection temporarily
		disableRes, err := r.esClient.Cluster.PutSettings(
			strings.NewReader(`{"transient":{"action.destructive_requires_name":false}}`),
			r.esClient.Cluster.PutSettings.WithContext(ctx),
		)
		if err == nil {
			disableRes.Body.Close()
		}

		// Delete targets
		res, err := r.esClient.Indices.Delete(
			targets,
			r.esClient.Indices.Delete.WithContext(ctx),
		)
		if err == nil {
			res.Body.Close()
		}

		// Re-enable wildcard protection
		enableRes, err := r.esClient.Cluster.PutSettings(
			strings.NewReader(`{"transient":{"action.destructive_requires_name":true}}`),
			r.esClient.Cluster.PutSettings.WithContext(ctx),
		)
		if err == nil {
			enableRes.Body.Close()
		}
	}
	return nil
}

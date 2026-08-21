package service

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"omnilogs-api/dto"
	"omnilogs-api/internal/archive_utils"
	globalauth "omnilogs-api/internal/global_auth/usecase"
	"omnilogs-api/internal/log_archive/objectstore"
	archivesnapshot "omnilogs-api/internal/log_archive/snapshot"
	retentionhelper "omnilogs-api/internal/retention_policy/helper"
	retentionprovider "omnilogs-api/internal/retention_policy/provider"
	retentionschedule "omnilogs-api/internal/retention_policy/schedule"
	"omnilogs-api/models"
	"omnilogs-api/responses"

	"github.com/elastic/go-elasticsearch/v8"
	"gorm.io/gorm"
)

type Actor struct {
	UserID        int
	PlatformAdmin bool
}
type actorContextKey struct{}

func WithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, actorContextKey{}, actor)
}
func actorFromContext(ctx context.Context) (Actor, bool) {
	value, ok := ctx.Value(actorContextKey{}).(Actor)
	return value, ok
}

type Service interface {
	Create(ctx context.Context, req dto.CreateArchiveRequest) (*models.LogArchiveJob, []models.LogArchive, error)
	List(ctx context.Context, req dto.ListArchiveRequest) ([]models.LogArchive, error)
	Get(ctx context.Context, productID, environmentID int, archiveID string) (*models.LogArchive, error)
	Verify(ctx context.Context, productID, environmentID int, archiveID string, policyID *int) (*dto.VerifyArchiveResponse, error)
	Restore(ctx context.Context, productID, environmentID int, archiveID string, req dto.RestoreArchiveRequest) (*dto.RestoreArchiveResponse, error)
	Download(ctx context.Context, productID, environmentID int, policyID, categoryID, featureID, subFeatureID *int, from, to *time.Time) (*DownloadedArchive, error)
	ListReadyDownloads(ctx context.Context, productID, environmentID int) ([]models.LogArchiveDownload, error)
	DownloadReady(ctx context.Context, productID, environmentID int, downloadID string) (*DownloadedArchive, error)
	Import(ctx context.Context, productID, environmentID int, sourcePath string, categoryID, featureID, subFeatureID *int) (*dto.ImportArchiveResponse, error)
	ListJobs(ctx context.Context, productID, environmentID int) ([]models.LogArchiveJob, error)
	GetJob(ctx context.Context, productID, environmentID int, jobID string) (*models.LogArchiveJob, error)
	PurgeExpired(ctx context.Context, productID, environmentID int, value int, unit string, neverDelete bool) error
}

type DownloadedArchive struct {
	File     *os.File
	Filename string
	Size     int64
}

type service struct {
	db       *gorm.DB
	es       *elasticsearch.Client
	provider *retentionprovider.Provider
	authz    globalauth.Usecase
	snapshot *archivesnapshot.Client
	r2       *objectstore.R2
}

func New(db *gorm.DB, es *elasticsearch.Client, authz globalauth.Usecase) Service {
	return &service{
		db:       db,
		es:       es,
		provider: retentionprovider.New(db, es),
		authz:    authz,
		snapshot: archivesnapshot.New(es, archivesnapshot.ConfigFromEnv()),
		r2:       objectstore.NewR2FromEnv(),
	}
}

func (s *service) Create(ctx context.Context, req dto.CreateArchiveRequest) (*models.LogArchiveJob, []models.LogArchive, error) {
	if err := s.authorize(ctx, req.ProductID, req.EnvironmentID, "ARCHIVE"); err != nil {
		return nil, nil, err
	}
	s.recordAudit(ctx, req.ProductID, "ARCHIVE_STARTED", nil, map[string]any{"environment_id": req.EnvironmentID})
	policy, err := s.policy(ctx, req.ProductID, req.EnvironmentID, req.PolicyID)
	if err != nil {
		return nil, nil, err
	}
	if !policy.ArchiveEnabled {
		return nil, nil, errors.New("archive is disabled by retention policy")
	}
	backupType, err := normalizeBackupType(req.BackupType)
	if err != nil {
		return nil, nil, err
	}
	if backupType == "MANUAL" && (req.DateFrom == nil || req.DateTo == nil) {
		return nil, nil, errors.New("date_from and date_to are required for manual backup")
	}
	from, to, err := s.archiveRange(ctx, policy, req)
	if err != nil {
		return nil, nil, err
	}
	job := &models.LogArchiveJob{JobID: newUUID(), ProductID: req.ProductID, EnvironmentID: req.EnvironmentID, Status: "PENDING", ArchiveIDs: []byte("[]")}
	if err := s.db.WithContext(ctx).Create(job).Error; err != nil {
		return nil, nil, err
	}
	started := time.Now().UTC()
	job.StartedAt = &started
	job.Status = "EXPORTING"
	_ = s.db.WithContext(ctx).Model(job).Updates(map[string]any{"status": job.Status, "started_at": started}).Error

	location, err := retentionschedule.Location(policy.ScheduleTimezone)
	if err != nil {
		return s.failJob(ctx, job, err)
	}
	backupTag := strings.TrimSpace(req.BackupTag)
	if backupTag == "" {
		backupTag = buildBackupTag(backupType, policy.PolicyID, from.In(location), to.In(location))
	}
	if len([]rune(backupTag)) > 180 {
		return s.failJob(ctx, job, errors.New("backup_tag must not exceed 180 characters"))
	}
	archives := make([]models.LogArchive, 0)
	for localDay := from.In(location); localDay.Before(to.In(location)); localDay = localDay.AddDate(0, 0, 1) {
		day := localDay.UTC()
		dayEnd := localDay.AddDate(0, 0, 1).UTC()
		queryFrom := archiveDayQueryFrom(policy, day, dayEnd, req.DateFrom != nil)
		indices, indexErr := s.provider.IndexNames(ctx, req.ProductID, req.EnvironmentID, &queryFrom, &dayEnd)
		if indexErr != nil {
			return s.failJob(ctx, job, indexErr)
		}
		var existingArchives []models.LogArchive
		if err := s.db.WithContext(ctx).Where("product_id = ? AND environment_id = ? AND date_from < ? AND date_to > ? AND status IN ?", req.ProductID, req.EnvironmentID, dayEnd, day, []string{"ARCHIVED", "VERIFIED", "SNAPSHOT_ONLY"}).Order("created_at DESC").Find(&existingArchives).Error; err != nil {
			return s.failJob(ctx, job, err)
		}
		// Do not create a database record or a manifest-only .ndjson.gz file for
		// a day that has no logs in this policy scope. Existing valid archives
		// are checked first so a retry can still reuse them after active logs
		// have been deleted.
		existingCount, countErr := s.provider.Count(ctx, req.ProductID, req.EnvironmentID, queryFrom, dayEnd, req.CategoryID, req.FeatureID, req.SubFeatureID)
		if countErr != nil {
			return s.failJob(ctx, job, countErr)
		}
		for _, existing := range existingArchives {
			if archiveScopeMatches(existing, req) && archiveReusable(existing, existingCount) {
				archives = append(archives, existing)
				existingCount = 0
				break
			}
		}
		if existingCount == 0 {
			continue
		}
		if len(indices) == 0 {
			return s.failJob(ctx, job, errors.New("active logs were counted but no Elasticsearch index matched the archive range"))
		}
		archiveID := newUUID()
		path := archive_utils.BuildScopedDailyArchiveVersionPath(storageRoot(), req.ProductID, req.EnvironmentID, req.CategoryID, req.FeatureID, req.SubFeatureID, day, archiveID)
		coverageKey := archiveCoverageKey(req.ProductID, req.EnvironmentID, req.CategoryID, req.FeatureID, req.SubFeatureID, localDay)
		status := "EXPORTING"
		format, compression := "NDJSON", "GZIP"
		archive := &models.LogArchive{ArchiveID: archiveID, ProductID: req.ProductID, EnvironmentID: &req.EnvironmentID, PolicyID: intPtr(policy.PolicyID), BackupType: backupType, BackupTag: backupTag, CoverageKey: &coverageKey, ArchiveYear: localDay.Year(), ArchiveMonth: int(localDay.Month()), DateFrom: &queryFrom, DateTo: &dayEnd, StorageProvider: "LOCAL", FilePath: &path, FileFormat: &format, Format: format, Compression: compression, Status: &status, IndexName: firstIndex(indices), CategoryID: req.CategoryID, FeatureID: req.FeatureID, SubFeatureID: req.SubFeatureID}
		_ = s.db.WithContext(ctx).Model(job).Update("status", "COMPRESSING").Error
		query := retentionprovider.BuildQuery(req.ProductID, req.EnvironmentID, &queryFrom, &dayEnd, req.CategoryID, req.FeatureID, req.SubFeatureID)
		result, streamErr := archive_utils.StreamArchive(archive_utils.StreamArchiveRequest{Client: s.es, Context: ctx, Indices: indices, Query: query, ProductID: req.ProductID, EnvironmentID: req.EnvironmentID, CategoryID: req.CategoryID, FeatureID: req.FeatureID, SubFeatureID: req.SubFeatureID, DateFrom: day, DateTo: dayEnd, FilePath: path})
		if streamErr != nil {
			return s.failJob(ctx, job, streamErr)
		}
		// Logs can disappear between the preflight count and streaming. Treat
		// that race as an empty day and leave no export artifact behind.
		if result.Empty || result.DocumentCount == 0 {
			_ = os.Remove(path)
			continue
		}
		_ = s.db.WithContext(ctx).Model(job).Update("status", "SNAPSHOTTING").Error
		checksum := result.Checksum
		archive.DocumentCount = result.DocumentCount
		archive.OriginalSizeBytes = result.OriginalBytes
		archive.CompressedSizeBytes = &result.CompressedBytes
		archive.Checksum = &checksum
		archive.ManifestValid = true
		if strings.EqualFold(strings.TrimSpace(policy.StorageProvider), "R2") {
			if !s.r2.Available() {
				_ = os.Remove(path)
				return s.failJob(ctx, job, fmt.Errorf("R2 backup is selected but unavailable: %w", s.r2.ConfigurationError()))
			}
			relative, relErr := filepath.Rel(storageRoot(), path)
			if relErr != nil {
				_ = os.Remove(path)
				return s.failJob(ctx, job, relErr)
			}
			objectKey := s.r2.Key(relative)
			tagging := "backup-type=" + url.QueryEscape(strings.ToLower(backupType)) + "&backup-tag=" + url.QueryEscape(backupTag)
			if uploadErr := s.r2.PutFile(ctx, objectKey, path, tagging, map[string]string{"coverage-key": coverageKey}); uploadErr != nil {
				_ = os.Remove(path)
				return s.failJob(ctx, job, uploadErr)
			}
			bucket := s.r2.Bucket()
			archive.StorageProvider = "R2"
			archive.BucketName = &bucket
			archive.ObjectKey = &objectKey
			archive.FilePath = nil
		}
		if err := s.db.WithContext(ctx).Create(archive).Error; err != nil {
			if archive.ObjectKey != nil {
				_ = s.r2.Delete(context.WithoutCancel(ctx), *archive.ObjectKey)
			}
			_ = os.Remove(path)
			return s.failJob(ctx, job, err)
		}
		if archive.ObjectKey != nil {
			_ = os.Remove(path)
		}
		if s.snapshot.Enabled() {
			snapshotName := buildSnapshotName(req.ProductID, req.EnvironmentID, day, archive.ArchiveID)
			snapshotResult, snapshotErr := s.snapshot.Create(ctx, snapshotName, indices)
			if snapshotErr != nil {
				failed := "FAILED"
				_ = s.db.WithContext(ctx).Model(archive).Updates(map[string]any{"status": failed, "snapshot_status": "FAILED"}).Error
				return s.failJob(ctx, job, snapshotErr)
			}
			snapshotIndices, marshalErr := json.Marshal(indices)
			if marshalErr != nil {
				return s.failJob(ctx, job, marshalErr)
			}
			repository := s.snapshot.Repository()
			snapshotStatus := snapshotResult.State
			archive.SnapshotRepository = &repository
			archive.SnapshotName = &snapshotResult.Name
			archive.SnapshotUUID = stringPtr(snapshotResult.UUID)
			archive.SnapshotStatus = &snapshotStatus
			archive.SnapshotIndices = snapshotIndices
		}
		archived := "ARCHIVED"
		now := time.Now().UTC()
		archive.Status = &archived
		archive.ExportedAt = &now
		archive.TotalLogs = intPtr(int(result.DocumentCount))
		if err := s.db.WithContext(ctx).Save(archive).Error; err != nil {
			return s.failJob(ctx, job, err)
		}
		_ = s.db.WithContext(ctx).Model(job).Update("status", "ARCHIVED").Error
		archives = append(archives, *archive)
	}
	ids := make([]string, 0, len(archives))
	for _, archive := range archives {
		ids = append(ids, archive.ArchiveID)
	}
	job.ArchiveIDs, _ = json.Marshal(ids)
	done := time.Now().UTC()
	job.Status = "COMPLETED"
	job.CompletedAt = &done
	if err := s.db.WithContext(ctx).Model(job).Updates(map[string]any{"status": job.Status, "archive_ids": job.ArchiveIDs, "completed_at": done}).Error; err != nil {
		return nil, nil, err
	}
	s.recordAudit(ctx, req.ProductID, "ARCHIVE_COMPLETED", firstArchiveID(archives), map[string]any{"environment_id": req.EnvironmentID, "document_count": documentCount(archives), "backup_type": backupType, "backup_tag": backupTag, "date_from": from, "date_to": to})
	return job, archives, nil
}

func (s *service) Verify(ctx context.Context, productID, environmentID int, archiveID string, policyID *int) (*dto.VerifyArchiveResponse, error) {
	if err := s.authorize(ctx, productID, environmentID, "VERIFY"); err != nil {
		return nil, err
	}
	archive, err := s.Get(ctx, productID, environmentID, archiveID)
	if err != nil {
		return nil, err
	}
	if archive.Status != nil && *archive.Status == "VERIFIED" && archive.DeletedFromActiveAt != nil {
		return &dto.VerifyArchiveResponse{ArchiveID: archiveID, Status: "VERIFIED", ChecksumValid: true, ManifestValid: archive.ManifestValid, SnapshotValid: archive.SnapshotVerifiedAt != nil, DocumentCount: archive.DocumentCount, DeletedActive: true}, nil
	}
	s.setJobStatus(ctx, archiveID, "VERIFYING")
	checksumValid := false
	manifestValid := false
	count := archive.DocumentCount
	if hasPortableArchive(*archive) {
		if archive.Checksum == nil {
			return nil, errors.New("archive checksum is missing")
		}
		archivePath, cleanup, pathErr := s.archiveLocalPath(ctx, *archive)
		if pathErr != nil {
			return nil, pathErr
		}
		defer cleanup()
		checksum, checksumErr := archive_utils.ChecksumFile(archivePath)
		if checksumErr != nil {
			return nil, checksumErr
		}
		if !strings.EqualFold(checksum, *archive.Checksum) {
			return nil, errors.New("archive checksum is invalid")
		}
		_, fileCount, verifyErr := archive_utils.VerifyArchiveScope(archivePath, productID, environmentID, archive.CategoryID, archive.FeatureID, archive.SubFeatureID)
		if verifyErr != nil {
			return nil, verifyErr
		}
		if fileCount != archive.DocumentCount {
			return nil, fmt.Errorf("archive document count mismatch: metadata=%d file=%d", archive.DocumentCount, fileCount)
		}
		count = fileCount
		checksumValid = true
		manifestValid = true
	}
	snapshotValid := false
	if archive.SnapshotName != nil && strings.TrimSpace(*archive.SnapshotName) != "" {
		if _, snapshotErr := s.snapshot.Verify(ctx, *archive.SnapshotName); snapshotErr != nil {
			return nil, snapshotErr
		}
		snapshotValid = true
	}
	if !checksumValid && !snapshotValid {
		return nil, errors.New("archive has no verifiable NDJSON or Elasticsearch snapshot")
	}
	verified := "VERIFIED"
	now := time.Now().UTC()
	updates := map[string]any{"status": verified, "verified_at": now, "manifest_valid": manifestValid}
	if snapshotValid {
		updates["snapshot_verified_at"] = now
		updates["snapshot_status"] = "VERIFIED"
	}
	if err := s.db.WithContext(ctx).Model(archive).Updates(updates).Error; err != nil {
		return nil, err
	}
	deleted := false
	deleteActive, err := s.deleteActiveEnabled(ctx, productID, environmentID, archive, policyID)
	if err != nil {
		return nil, err
	}
	if count > 0 && deleteActive {
		if archive.DateFrom == nil || archive.DateTo == nil {
			return nil, errors.New("archive date range is missing")
		}
		if err := s.authorize(ctx, productID, environmentID, "DELETE"); err != nil {
			return nil, err
		}
		s.setJobStatus(ctx, archiveID, "DELETING_ACTIVE")
		deletedCount, deleteErr := s.provider.DeleteActive(ctx, productID, environmentID, archive.DateFrom.UTC(), archive.DateTo.UTC(), archive.CategoryID, archive.FeatureID, archive.SubFeatureID, count)
		if deleteErr != nil {
			s.setJobStatus(ctx, archiveID, "FAILED")
			s.recordAudit(ctx, productID, "ARCHIVE_FAILED", &archiveID, map[string]any{"environment_id": environmentID, "error": deleteErr.Error()})
			return nil, deleteErr
		}
		if deletedCount != count {
			return nil, errors.New("safe delete count verification failed")
		}
		deleted = true
		deletedAt := time.Now().UTC()
		deleteUpdates := map[string]any{"deleted_from_active_at": deletedAt, "deleted_at": deletedAt}
		if archive.RestoreStatus != nil && *archive.RestoreStatus == "RESTORED" {
			deleteUpdates["restore_status"] = "REARCHIVED"
		}
		if err := s.db.WithContext(ctx).Model(archive).Updates(deleteUpdates).Error; err != nil {
			return nil, err
		}
	}
	s.setJobStatus(ctx, archiveID, "COMPLETED")
	s.recordAudit(ctx, productID, "ARCHIVE_VERIFIED", &archiveID, map[string]any{"environment_id": environmentID, "document_count": count})
	if deleted {
		_ = s.markPreviousRestoresRearchived(ctx, archive)
		s.recordAudit(ctx, productID, "HISTORICAL_LOGS_DELETED", &archiveID, map[string]any{"environment_id": environmentID, "document_count": count})
	}
	return &dto.VerifyArchiveResponse{ArchiveID: archiveID, Status: verified, ChecksumValid: checksumValid, ManifestValid: manifestValid, SnapshotValid: snapshotValid, DocumentCount: count, DeletedActive: deleted}, nil
}

func (s *service) Restore(ctx context.Context, productID, environmentID int, archiveID string, req dto.RestoreArchiveRequest) (*dto.RestoreArchiveResponse, error) {
	if err := s.authorize(ctx, productID, environmentID, "RESTORE"); err != nil {
		return nil, err
	}
	archive, err := s.Get(ctx, productID, environmentID, archiveID)
	if err != nil {
		return nil, err
	}
	mode := strings.ToUpper(strings.TrimSpace(req.Mode))
	if mode == "" {
		mode = archive_utils.RestoreSearchOnly
	}
	snapshotReady := archive.SnapshotName != nil && strings.TrimSpace(*archive.SnapshotName) != "" &&
		archive.SnapshotStatus != nil && (*archive.SnapshotStatus == "SUCCESS" || *archive.SnapshotStatus == "VERIFIED")
	if (archive.Status == nil || *archive.Status != "VERIFIED") && !(mode == archive_utils.RestoreSnapshotSearch && snapshotReady) {
		return nil, errors.New("archive must be VERIFIED before restore")
	}
	s.recordAudit(ctx, productID, "RESTORE_STARTED", &archiveID, map[string]any{"environment_id": environmentID, "mode": req.Mode})
	if mode == archive_utils.RestoreSnapshotSearch {
		if archive.SnapshotName == nil || strings.TrimSpace(*archive.SnapshotName) == "" {
			return nil, errors.New("archive does not contain an Elasticsearch snapshot")
		}
		var indices []string
		if len(archive.SnapshotIndices) > 0 {
			if err := json.Unmarshal(archive.SnapshotIndices, &indices); err != nil {
				return nil, fmt.Errorf("decode snapshot indices: %w", err)
			}
		}
		prefix := fmt.Sprintf("omnilogs-restore-%s-%d", archiveID, time.Now().UTC().Unix())
		pattern, restoreErr := s.snapshot.Restore(ctx, *archive.SnapshotName, indices, prefix)
		if restoreErr != nil {
			s.recordAudit(ctx, productID, "RESTORE_FAILED", &archiveID, map[string]any{"environment_id": environmentID, "error": restoreErr.Error()})
			return nil, restoreErr
		}
		now := time.Now().UTC()
		restored := "RESTORED"
		_ = s.db.WithContext(ctx).Model(archive).Updates(map[string]any{"restored_at": now, "restore_index_pattern": pattern, "restore_status": restored}).Error
		s.recordAudit(ctx, productID, "RESTORE_COMPLETED", &archiveID, map[string]any{"environment_id": environmentID, "index_pattern": pattern, "snapshot": true})
		return &dto.RestoreArchiveResponse{ArchiveID: archiveID, Mode: mode, IndexName: pattern, Snapshot: true, Conflict: archive_utils.ConflictSkipExisting}, nil
	}
	if !hasPortableArchive(*archive) || archive.Checksum == nil {
		return nil, errors.New("archive file or checksum is missing")
	}
	archivePath, cleanup, err := s.archiveLocalPath(ctx, *archive)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	checksum, err := archive_utils.ChecksumFile(archivePath)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(checksum, *archive.Checksum) {
		return nil, errors.New("archive checksum is invalid")
	}
	if _, _, err := archive_utils.VerifyArchiveScope(archivePath, productID, environmentID, archive.CategoryID, archive.FeatureID, archive.SubFeatureID); err != nil {
		return nil, err
	}
	conflict := strings.ToUpper(strings.TrimSpace(req.Conflict))
	if conflict == "" {
		conflict = archive_utils.ConflictSkipExisting
	}
	target := fmt.Sprintf("omnilogs-archive-search-%s", archiveID)
	if mode == archive_utils.RestoreToActive {
		if archive.DateFrom == nil {
			return nil, errors.New("archive date range is missing")
		}
		target = fmt.Sprintf("omnilogs-product-%d-env-%d-search-v3-%s", productID, environmentID, archive.DateFrom.UTC().Format("2006.01.02"))
	}
	result, err := archive_utils.RestoreNDJSON(ctx, s.es, archivePath, mode, conflict, target, productID, environmentID)
	if err != nil {
		s.recordAudit(ctx, productID, "RESTORE_FAILED", &archiveID, map[string]any{
			"environment_id": environmentID,
			"restored_count": result.Restored,
			"skipped_count":  result.Skipped,
			"failed_count":   result.Failed,
			"error":          err.Error(),
		})
		return nil, fmt.Errorf("restore incomplete for archive %s (restored=%d skipped=%d failed=%d): %w", archiveID, result.Restored, result.Skipped, result.Failed, err)
	}
	now := time.Now().UTC()
	restoredStatus := "RESTORED"
	restoreUpdates := map[string]any{"restored_at": now, "restore_index_pattern": target, "restore_status": restoredStatus}
	if mode == archive_utils.RestoreToActive {
		restoreUpdates["deleted_from_active_at"] = nil
		restoreUpdates["deleted_at"] = nil
	}
	if err := s.db.WithContext(ctx).Model(archive).Updates(restoreUpdates).Error; err != nil {
		return nil, err
	}
	s.recordAudit(ctx, productID, "RESTORE_COMPLETED", &archiveID, map[string]any{"environment_id": environmentID, "restored_count": result.Restored, "skipped_count": result.Skipped})
	return &dto.RestoreArchiveResponse{ArchiveID: archiveID, Mode: mode, IndexName: target, Restored: result.Restored, Skipped: result.Skipped, Conflict: conflict}, nil
}

func (s *service) Download(ctx context.Context, productID, environmentID int, policyID, categoryID, featureID, subFeatureID *int, from, to *time.Time) (*DownloadedArchive, error) {
	if err := s.authorize(ctx, productID, environmentID, "READ"); err != nil {
		return nil, err
	}
	if policyID != nil {
		var policy models.LogRetentionPolicy
		if err := s.db.WithContext(ctx).Where("policy_id = ? AND product_id = ? AND environment_id = ?", *policyID, productID, environmentID).First(&policy).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, responses.ErrNotFound
		} else if err != nil {
			return nil, err
		}
	}
	query := s.db.WithContext(ctx).Where("product_id = ? AND environment_id = ? AND status IN ? AND document_count > 0", productID, environmentID, []string{"ARCHIVED", "VERIFIED"})
	if policyID != nil {
		// Archives created before policy_id was added remain downloadable as
		// legacy archives, provided they belong to the same product/environment.
		query = query.Where("policy_id = ? OR policy_id IS NULL", *policyID)
	}
	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}
	if featureID != nil {
		query = query.Where("feature_id = ?", *featureID)
	}
	if subFeatureID != nil {
		query = query.Where("sub_feature_id = ?", *subFeatureID)
	}
	if from != nil {
		query = query.Where("date_to > ?", from)
	}
	if to != nil {
		query = query.Where("date_from < ?", to)
	}
	var archives []models.LogArchive
	if err := query.Order("date_from ASC, created_at ASC").Find(&archives).Error; err != nil {
		return nil, err
	}
	if len(archives) == 0 {
		return nil, errors.New("no verified archives matched the selected policy and date range")
	}

	temp, err := os.CreateTemp("", fmt.Sprintf("omnilogs-product-%d-env-%d-*.zip", productID, environmentID))
	if err != nil {
		return nil, err
	}
	removeOnError := true
	defer func() {
		if removeOnError {
			_ = temp.Close()
			_ = os.Remove(temp.Name())
		}
	}()
	archiveWriter := zip.NewWriter(temp)
	manifest := map[string]any{
		"product_id": productID, "environment_id": environmentID, "policy_id": policyID,
		"category_id": categoryID, "feature_id": featureID, "sub_feature_id": subFeatureID,
		"archive_count": len(archives), "archives": make([]map[string]any, 0, len(archives)),
	}
	manifestArchives := manifest["archives"].([]map[string]any)
	for _, archive := range archives {
		archivePath, cleanup, pathErr := s.archiveLocalPath(ctx, archive)
		if pathErr != nil {
			return nil, fmt.Errorf("materialize archive %s: %w", archive.ArchiveID, pathErr)
		}
		input, openErr := os.Open(archivePath)
		if openErr != nil {
			cleanup()
			return nil, openErr
		}
		ndjson, gzipErr := gzip.NewReader(input)
		if gzipErr != nil {
			_ = input.Close()
			cleanup()
			return nil, fmt.Errorf("open archive %s as gzip: %w", archive.ArchiveID, gzipErr)
		}
		dateLabel := "unknown-date"
		if archive.DateFrom != nil {
			dateLabel = archive.DateFrom.UTC().Format("2006-01-02")
		}
		featureLabel := "all"
		if archive.FeatureID != nil {
			featureLabel = strconv.Itoa(*archive.FeatureID)
		}
		subFeatureLabel := "all"
		if archive.SubFeatureID != nil {
			subFeatureLabel = strconv.Itoa(*archive.SubFeatureID)
		}
		// The portable ZIP contains plain NDJSON so users can open an extracted
		// file directly. The on-disk archive remains gzip-compressed.
		entryName := fmt.Sprintf("archives/feature-%s/sub-feature-%s/%s-%s.ndjson", featureLabel, subFeatureLabel, dateLabel, archive.ArchiveID)
		header := &zip.FileHeader{Name: entryName, Method: zip.Deflate}
		if archive.DateFrom != nil {
			header.SetModTime(archive.DateFrom.UTC())
		}
		entry, createErr := archiveWriter.CreateHeader(header)
		if createErr == nil {
			_, createErr = io.Copy(entry, ndjson)
		}
		_ = ndjson.Close()
		_ = input.Close()
		cleanup()
		if createErr != nil {
			return nil, createErr
		}
		manifestArchives = append(manifestArchives, map[string]any{
			"archive_id": archive.ArchiveID, "date_from": archive.DateFrom, "date_to": archive.DateTo,
			"category_id": archive.CategoryID, "feature_id": archive.FeatureID, "sub_feature_id": archive.SubFeatureID,
			"document_count": archive.DocumentCount, "status": archive.Status,
		})
	}
	manifest["archives"] = manifestArchives
	manifestEntry, err := archiveWriter.Create("manifest.json")
	if err != nil {
		return nil, err
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	if _, err := manifestEntry.Write(manifestBytes); err != nil {
		return nil, err
	}
	if err := archiveWriter.Close(); err != nil {
		return nil, err
	}
	if err := temp.Sync(); err != nil {
		return nil, err
	}
	info, err := temp.Stat()
	if err != nil {
		return nil, err
	}
	if _, err := temp.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	removeOnError = false
	return &DownloadedArchive{File: temp, Filename: fmt.Sprintf("omnilogs-product-%d-environment-%d.zip", productID, environmentID), Size: info.Size()}, nil
}

func (s *service) Import(ctx context.Context, productID, environmentID int, sourcePath string, categoryID, featureID, subFeatureID *int) (*dto.ImportArchiveResponse, error) {
	if err := s.authorize(ctx, productID, environmentID, "RESTORE"); err != nil {
		return nil, err
	}
	if strings.TrimSpace(sourcePath) == "" {
		return nil, errors.New("archive upload is required")
	}
	reader, err := zip.OpenReader(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("open archive ZIP: %w", err)
	}
	defer reader.Close()

	var packageManifest struct {
		ProductID     int  `json:"product_id"`
		EnvironmentID int  `json:"environment_id"`
		PolicyID      *int `json:"policy_id"`
	}
	manifestFound := false
	for _, entry := range reader.File {
		if entry.Name != "manifest.json" {
			continue
		}
		file, openErr := entry.Open()
		if openErr != nil {
			return nil, openErr
		}
		decodeErr := json.NewDecoder(io.LimitReader(file, 2<<20)).Decode(&packageManifest)
		_ = file.Close()
		if decodeErr != nil {
			return nil, fmt.Errorf("invalid package manifest: %w", decodeErr)
		}
		manifestFound = true
		break
	}
	if !manifestFound || packageManifest.ProductID != productID || packageManifest.EnvironmentID != environmentID {
		return nil, errors.New("archive package product/environment scope is invalid")
	}

	importPolicyID := packageManifest.PolicyID
	if importPolicyID != nil {
		var policy models.LogRetentionPolicy
		if err := s.db.WithContext(ctx).Where("policy_id = ? AND product_id = ? AND environment_id = ?", *importPolicyID, productID, environmentID).First(&policy).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
			// Policy configuration may legitimately be replaced while the ZIP
			// remains a durable copy of the logs. Scope is validated by both the
			// package and every per-file manifest, so detach a stale policy ID.
			importPolicyID = nil
		}
	}

	created := make([]models.LogArchive, 0)
	available := make([]models.LogArchive, 0)
	createdPaths := make([]string, 0)
	rollback := func() {
		if len(created) > 0 {
			ids := make([]string, 0, len(created))
			for _, value := range created {
				ids = append(ids, value.ArchiveID)
			}
			_ = s.db.WithContext(context.Background()).Where("archive_id IN ?", ids).Delete(&models.LogArchive{}).Error
		}
		for _, path := range createdPaths {
			_ = os.Remove(path)
		}
	}

	for _, entry := range reader.File {
		if entry.FileInfo().IsDir() || entry.Name == "manifest.json" {
			continue
		}
		if !strings.HasPrefix(entry.Name, "archives/") || (!strings.HasSuffix(entry.Name, ".ndjson.gz") && !strings.HasSuffix(entry.Name, ".ndjson")) {
			continue
		}
		if strings.Contains(entry.Name, "..") || strings.HasPrefix(entry.Name, "/") || entry.UncompressedSize64 > maxImportEntryBytes {
			rollback()
			return nil, errors.New("archive entry path or size is invalid")
		}
		rawPath, err := extractArchiveEntry(entry)
		if err != nil {
			rollback()
			return nil, err
		}
		gzipPath, normalizeErr := normalizeArchiveGZIP(rawPath)
		_ = os.Remove(rawPath)
		if normalizeErr != nil {
			rollback()
			return nil, normalizeErr
		}
		manifest, count, verifyErr := archive_utils.VerifyArchiveScope(gzipPath, productID, environmentID, categoryID, featureID, subFeatureID)
		if verifyErr != nil {
			_ = os.Remove(gzipPath)
			rollback()
			return nil, verifyErr
		}
		if manifest.DateTo.IsZero() || !manifest.DateTo.After(manifest.DateFrom) {
			_ = os.Remove(gzipPath)
			rollback()
			return nil, errors.New("archive date range is invalid")
		}
		var duplicates []models.LogArchive
		duplicateQuery := s.db.WithContext(ctx).Model(&models.LogArchive{}).
			Where("product_id = ? AND environment_id = ? AND date_from = ? AND date_to = ? AND document_count = ? AND status IN ?", productID, environmentID, manifest.DateFrom, manifest.DateTo, count, []string{"ARCHIVED", "VERIFIED", "SNAPSHOT_ONLY"})
		duplicateQuery = applyArchiveScopeQuery(duplicateQuery, manifest.CategoryID, manifest.FeatureID, manifest.SubFeatureID)
		if err := duplicateQuery.Order("verified_at DESC NULLS LAST, created_at DESC").Find(&duplicates).Error; err != nil {
			_ = os.Remove(gzipPath)
			rollback()
			return nil, err
		}
		for _, duplicate := range duplicates {
			if archiveFileAvailable(duplicate) {
				available = append(available, duplicate)
				_ = os.Remove(gzipPath)
				gzipPath = ""
				break
			}
		}
		if gzipPath == "" {
			continue
		}
		archiveID := newUUID()
		path := archive_utils.BuildScopedDailyArchiveVersionPath(storageRoot(), productID, environmentID, manifest.CategoryID, manifest.FeatureID, manifest.SubFeatureID, manifest.DateFrom, archiveID)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			_ = os.Remove(gzipPath)
			rollback()
			return nil, err
		}
		if err := copyFileExclusive(gzipPath, path); err != nil {
			_ = os.Remove(gzipPath)
			rollback()
			return nil, err
		}
		_ = os.Remove(gzipPath)
		createdPaths = append(createdPaths, path)
		checksum, checksumErr := archive_utils.ChecksumFile(path)
		if checksumErr != nil {
			rollback()
			return nil, checksumErr
		}
		format, compression, status := "NDJSON", "GZIP", "VERIFIED"
		archive := models.LogArchive{
			ArchiveID: archiveID, ProductID: productID, EnvironmentID: &environmentID, PolicyID: importPolicyID,
			CategoryID: manifest.CategoryID, FeatureID: manifest.FeatureID, SubFeatureID: manifest.SubFeatureID,
			ArchiveYear: manifest.DateFrom.Year(), ArchiveMonth: int(manifest.DateFrom.Month()), DateFrom: &manifest.DateFrom, DateTo: &manifest.DateTo,
			StorageProvider: "LOCAL", FilePath: &path, FileFormat: &format, Format: format, Compression: compression,
			Status: &status, DocumentCount: count, OriginalSizeBytes: manifest.OriginalBytes, Checksum: &checksum,
			ManifestValid: true, ExportedAt: timePtr(time.Now().UTC()), VerifiedAt: timePtr(time.Now().UTC()), TotalLogs: intPtr(int(count)),
		}
		if info, statErr := os.Stat(path); statErr == nil {
			compressed := info.Size()
			archive.CompressedSizeBytes = &compressed
		}
		if err := s.db.WithContext(ctx).Create(&archive).Error; err != nil {
			rollback()
			return nil, err
		}
		created = append(created, archive)
		available = append(available, archive)
	}
	if len(available) == 0 {
		return nil, errors.New("archive ZIP does not contain NDJSON archive entries")
	}
	return &dto.ImportArchiveResponse{ImportedCount: len(created), Archives: available}, nil
}

func archiveFileAvailable(archive models.LogArchive) bool {
	if archive.ObjectKey != nil && strings.TrimSpace(*archive.ObjectKey) != "" {
		return true
	}
	if archive.FilePath == nil || strings.TrimSpace(*archive.FilePath) == "" {
		return false
	}
	info, err := os.Stat(*archive.FilePath)
	return err == nil && info.Mode().IsRegular() && info.Size() > 0
}

const maxImportEntryBytes = uint64(10 << 30)

func extractArchiveEntry(entry *zip.File) (string, error) {
	input, err := entry.Open()
	if err != nil {
		return "", err
	}
	defer input.Close()
	file, err := os.CreateTemp("", "omnilogs-import-entry-*")
	if err != nil {
		return "", err
	}
	path := file.Name()
	removeOnError := true
	defer func() {
		_ = file.Close()
		if removeOnError {
			_ = os.Remove(path)
		}
	}()
	if _, err := io.Copy(file, io.LimitReader(input, int64(maxImportEntryBytes)+1)); err != nil {
		return "", err
	}
	if info, err := file.Stat(); err != nil || info.Size() > int64(maxImportEntryBytes) {
		return "", errors.New("archive entry exceeds import size limit")
	}
	if err := file.Close(); err != nil {
		return "", err
	}
	removeOnError = false
	return path, nil
}

func normalizeArchiveGZIP(sourcePath string) (string, error) {
	input, err := os.Open(sourcePath)
	if err != nil {
		return "", err
	}
	defer input.Close()
	header := make([]byte, 2)
	if _, err := io.ReadFull(input, header); err != nil {
		return "", err
	}
	if _, err := input.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	output, err := os.CreateTemp("", "omnilogs-import-gzip-*.ndjson.gz")
	if err != nil {
		return "", err
	}
	path := output.Name()
	removeOnError := true
	defer func() {
		_ = output.Close()
		if removeOnError {
			_ = os.Remove(path)
		}
	}()
	if bytes.Equal(header, []byte{0x1f, 0x8b}) {
		_, err = io.Copy(output, input)
	} else {
		writer := gzip.NewWriter(output)
		_, err = io.Copy(writer, input)
		if closeErr := writer.Close(); err == nil {
			err = closeErr
		}
	}
	if err != nil {
		return "", err
	}
	if err := output.Close(); err != nil {
		return "", err
	}
	removeOnError = false
	return path, nil
}

func copyFileExclusive(sourcePath, destinationPath string) error {
	input, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	removeOnError := true
	defer func() {
		_ = output.Close()
		if removeOnError {
			_ = os.Remove(destinationPath)
		}
	}()
	if _, err := io.Copy(output, input); err != nil {
		return err
	}
	if err := output.Close(); err != nil {
		return err
	}
	removeOnError = false
	return nil
}

func applyArchiveScopeQuery(query *gorm.DB, categoryID, featureID, subFeatureID *int) *gorm.DB {
	if categoryID == nil {
		query = query.Where("category_id IS NULL")
	} else {
		query = query.Where("category_id = ?", *categoryID)
	}
	if featureID == nil {
		query = query.Where("feature_id IS NULL")
	} else {
		query = query.Where("feature_id = ?", *featureID)
	}
	if subFeatureID == nil {
		query = query.Where("sub_feature_id IS NULL")
	} else {
		query = query.Where("sub_feature_id = ?", *subFeatureID)
	}
	return query
}

func (s *service) List(ctx context.Context, req dto.ListArchiveRequest) ([]models.LogArchive, error) {
	if err := s.authorize(ctx, req.ProductID, req.EnvironmentID, "READ"); err != nil {
		return nil, err
	}
	query := s.db.WithContext(ctx).Where("product_id = ? AND environment_id = ? AND document_count > 0", req.ProductID, req.EnvironmentID)
	if req.CategoryID != nil {
		query = query.Where("category_id = ?", *req.CategoryID)
	}
	if req.FeatureID != nil {
		query = query.Where("feature_id = ?", *req.FeatureID)
	}
	if req.SubFeatureID != nil {
		query = query.Where("sub_feature_id = ?", *req.SubFeatureID)
	}
	if req.DateFrom != nil {
		query = query.Where("date_to > ?", req.DateFrom.UTC())
	}
	if req.DateTo != nil {
		query = query.Where("date_from < ?", req.DateTo.UTC())
	}
	if value := strings.ToUpper(strings.TrimSpace(req.BackupType)); value != "" {
		if value != "POLICY" && value != "MANUAL" {
			return nil, errors.New("backup_type must be POLICY or MANUAL")
		}
		query = query.Where("backup_type = ?", value)
	}
	if value := strings.TrimSpace(req.BackupTag); value != "" {
		query = query.Where("backup_tag = ?", value)
	}
	var archives []models.LogArchive
	err := query.Order("date_from DESC, created_at DESC").Find(&archives).Error
	return archives, err
}

func (s *service) Get(ctx context.Context, productID, environmentID int, archiveID string) (*models.LogArchive, error) {
	var archive models.LogArchive
	err := s.db.WithContext(ctx).Where("archive_id = ? AND product_id = ? AND environment_id = ?", archiveID, productID, environmentID).First(&archive).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, responses.ErrNotFound
	}
	return &archive, err
}

func (s *service) ListJobs(ctx context.Context, productID, environmentID int) ([]models.LogArchiveJob, error) {
	if err := s.authorize(ctx, productID, environmentID, "READ"); err != nil {
		return nil, err
	}
	var jobs []models.LogArchiveJob
	err := s.db.WithContext(ctx).Where("product_id = ? AND environment_id = ?", productID, environmentID).Order("created_at DESC").Find(&jobs).Error
	return jobs, err
}

func (s *service) GetJob(ctx context.Context, productID, environmentID int, jobID string) (*models.LogArchiveJob, error) {
	if err := s.authorize(ctx, productID, environmentID, "READ"); err != nil {
		return nil, err
	}
	var job models.LogArchiveJob
	err := s.db.WithContext(ctx).Where("job_id = ? AND product_id = ? AND environment_id = ?", jobID, productID, environmentID).First(&job).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, responses.ErrNotFound
	}
	return &job, err
}

func (s *service) PurgeExpired(ctx context.Context, productID, environmentID, value int, unit string, neverDelete bool) error {
	if neverDelete || value <= 0 {
		return nil
	}
	cutoff, err := retentionhelper.Subtract(time.Now().UTC(), value, unit)
	if err != nil {
		return err
	}
	if err := s.authorize(ctx, productID, environmentID, "DELETE"); err != nil {
		return err
	}
	var archives []models.LogArchive
	if err := s.db.WithContext(ctx).Where("product_id = ? AND environment_id = ? AND status = ? AND date_to < ?", productID, environmentID, "VERIFIED", cutoff).Find(&archives).Error; err != nil {
		return err
	}
	groups := make(map[string][]models.LogArchive)
	for _, archive := range archives {
		key := "unscoped"
		if archive.PolicyID != nil {
			key = strconv.Itoa(*archive.PolicyID)
		}
		groups[key] = append(groups[key], archive)
	}
	for _, group := range groups {
		// A portable ZIP is created before the daily files are removed. This is
		// the durable "Ready to load" item shown by the UI.
		if err := s.createReadyDownload(ctx, productID, environmentID, group); err != nil {
			return err
		}
		for i := range group {
			if group[i].FilePath != nil && *group[i].FilePath != "" {
				if removeErr := os.Remove(*group[i].FilePath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
					return removeErr
				}
			}
			if group[i].ObjectKey != nil && strings.TrimSpace(*group[i].ObjectKey) != "" {
				if removeErr := s.r2.Delete(ctx, *group[i].ObjectKey); removeErr != nil {
					return removeErr
				}
			}
			status := "EXPIRED"
			now := time.Now().UTC()
			updates := map[string]any{"status": status, "deleted_at": now, "purged_at": now, "file_path": nil, "object_key": nil, "checksum": nil}
			if group[i].SnapshotName != nil && strings.TrimSpace(*group[i].SnapshotName) != "" {
				// The Elasticsearch snapshot remains available for restore after
				// the portable daily file has expired.
				updates["status"] = "SNAPSHOT_ONLY"
			}
			if err := s.db.WithContext(ctx).Model(&group[i]).Updates(updates).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *service) ListReadyDownloads(ctx context.Context, productID, environmentID int) ([]models.LogArchiveDownload, error) {
	if err := s.authorize(ctx, productID, environmentID, "READ"); err != nil {
		return nil, err
	}
	var downloads []models.LogArchiveDownload
	err := s.db.WithContext(ctx).
		Where("product_id = ? AND environment_id = ? AND status = ?", productID, environmentID, "READY_TO_LOAD").
		Order("created_at DESC").Find(&downloads).Error
	return downloads, err
}

func (s *service) DownloadReady(ctx context.Context, productID, environmentID int, downloadID string) (*DownloadedArchive, error) {
	if err := s.authorize(ctx, productID, environmentID, "READ"); err != nil {
		return nil, err
	}
	var download models.LogArchiveDownload
	if err := s.db.WithContext(ctx).
		Where("download_id = ? AND product_id = ? AND environment_id = ? AND status = ?", downloadID, productID, environmentID, "READY_TO_LOAD").
		First(&download).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, responses.ErrNotFound
		}
		return nil, err
	}
	file, err := os.Open(download.FilePath)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	now := time.Now().UTC()
	_ = s.db.WithContext(ctx).Model(&download).Update("downloaded_at", now).Error
	return &DownloadedArchive{File: file, Filename: download.FileName, Size: info.Size()}, nil
}

func (s *service) createReadyDownload(ctx context.Context, productID, environmentID int, archives []models.LogArchive) error {
	files := make([]models.LogArchive, 0, len(archives))
	for _, archive := range archives {
		if hasPortableArchive(archive) {
			files = append(files, archive)
		}
	}
	if len(files) == 0 {
		return nil
	}

	downloadID := newUUID()
	directory := filepath.Join(storageRoot(), "downloads", fmt.Sprintf("product-%d", productID), fmt.Sprintf("environment-%d", environmentID))
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}
	filename := fmt.Sprintf("omnilogs-product-%d-environment-%d-%s.zip", productID, environmentID, downloadID)
	path := filepath.Join(directory, filename)
	temp, err := os.CreateTemp(directory, ".ready-*.zip")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	removeTemp := true
	defer func() {
		_ = temp.Close()
		if removeTemp {
			_ = os.Remove(tempPath)
		}
	}()
	writer := zip.NewWriter(temp)
	manifest := map[string]any{"product_id": productID, "environment_id": environmentID, "archive_count": len(files), "archives": make([]map[string]any, 0, len(files))}
	manifestArchives := manifest["archives"].([]map[string]any)
	var dateFrom, dateTo *time.Time
	for _, archive := range files {
		archivePath, cleanup, pathErr := s.archiveLocalPath(ctx, archive)
		if pathErr != nil {
			return pathErr
		}
		input, openErr := os.Open(archivePath)
		if openErr != nil {
			cleanup()
			return openErr
		}
		entryName := fmt.Sprintf("archives/%s.ndjson.gz", archive.ArchiveID)
		entry, createErr := writer.Create(entryName)
		if createErr == nil {
			_, createErr = io.Copy(entry, input)
		}
		_ = input.Close()
		cleanup()
		if createErr != nil {
			return createErr
		}
		manifestArchives = append(manifestArchives, map[string]any{"archive_id": archive.ArchiveID, "date_from": archive.DateFrom, "date_to": archive.DateTo, "document_count": archive.DocumentCount})
		if dateFrom == nil || (archive.DateFrom != nil && archive.DateFrom.Before(*dateFrom)) {
			dateFrom = archive.DateFrom
		}
		if dateTo == nil || (archive.DateTo != nil && archive.DateTo.After(*dateTo)) {
			dateTo = archive.DateTo
		}
	}
	manifest["archives"] = manifestArchives
	manifestEntry, err := writer.Create("manifest.json")
	if err != nil {
		return err
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	if _, err := manifestEntry.Write(manifestBytes); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempPath, path); err != nil {
		return err
	}
	removeTemp = false
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	policyID := files[0].PolicyID
	status := "READY_TO_LOAD"
	created := time.Now().UTC()
	return s.db.WithContext(ctx).Create(&models.LogArchiveDownload{DownloadID: downloadID, ProductID: productID, EnvironmentID: environmentID, PolicyID: policyID, DateFrom: dateFrom, DateTo: dateTo, FileName: filename, FilePath: path, SizeBytes: info.Size(), ArchiveCount: len(files), Status: status, CreatedAt: &created}).Error
}

func (s *service) policy(ctx context.Context, productID, environmentID int, policyID *int) (*models.LogRetentionPolicy, error) {
	var policy models.LogRetentionPolicy
	query := s.db.WithContext(ctx).Where("product_id = ? AND environment_id = ? AND is_active = TRUE", productID, environmentID)
	if policyID != nil && *policyID > 0 {
		query = query.Where("policy_id = ?", *policyID)
	}
	err := query.Order("updated_at DESC NULLS LAST").First(&policy).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("active retention policy not found")
	}
	return &policy, err
}

func (s *service) deleteActiveEnabled(ctx context.Context, productID, environmentID int, archive *models.LogArchive, requestedPolicyID *int) (bool, error) {
	policyID := requestedPolicyID
	if policyID == nil && archive != nil {
		policyID = archive.PolicyID
	}
	var policy models.LogRetentionPolicy
	query := s.db.WithContext(ctx).Where("product_id = ? AND environment_id = ?", productID, environmentID)
	if policyID != nil && *policyID > 0 {
		query = query.Where("policy_id = ?", *policyID)
	} else {
		query = query.Where("is_active = TRUE").Order("updated_at DESC NULLS LAST")
	}
	if err := query.First(&policy).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Verification must remain non-destructive when the originating
			// policy has been removed. The archive is still valid and restorable.
			return false, nil
		}
		return false, err
	}
	return policy.DeleteActiveAfterArchive, nil
}

func (s *service) archiveRange(ctx context.Context, policy *models.LogRetentionPolicy, req dto.CreateArchiveRequest) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	location, err := retentionschedule.Location(policy.ScheduleTimezone)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	var from, to time.Time
	if req.DateFrom != nil {
		from = calendarDate(*req.DateFrom, location).UTC()
	}
	if req.DateTo != nil {
		// Explicit ranges are half-open [date_from, date_to). This avoids the
		// former double +1 day bug between the browser and backend.
		to = calendarDate(*req.DateTo, location).UTC()
	} else if policy.ArchiveAfterValue != nil && policy.ArchiveAfterUnit != "" {
		value, err := retentionhelper.Subtract(now.In(location), *policy.ArchiveAfterValue, policy.ArchiveAfterUnit)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		to = startOfDayIn(value, location).UTC()
	} else {
		return time.Time{}, time.Time{}, errors.New("date_to or archive_after policy is required")
	}
	if req.DateFrom == nil {
		if policy.ApplyToExistingLogs {
			oldest, err := s.provider.OldestTimestamp(ctx, req.ProductID, req.EnvironmentID, req.CategoryID, req.FeatureID, req.SubFeatureID)
			if err != nil {
				return time.Time{}, time.Time{}, err
			}
			if oldest != nil {
				from = startOfDayIn(oldest.In(location), location).UTC()
			}
		} else if policy.EffectiveFrom != nil {
			from = startOfDayIn(policy.EffectiveFrom.In(location), location).UTC()
		} else if policy.CreatedAt != nil {
			from = startOfDayIn(policy.CreatedAt.In(location), location).UTC()
		}
		if from.IsZero() || !to.After(from) {
			// No complete eligible day exists yet. Return a deterministic empty
			// window so the job completes successfully without creating files.
			from = to.AddDate(0, 0, -1)
		}
	}
	if !to.After(from) {
		return time.Time{}, time.Time{}, errors.New("date_to must be after date_from")
	}
	return from, to, nil
}

func (s *service) authorize(ctx context.Context, productID, environmentID int, action string) error {
	if productID <= 0 || environmentID <= 0 {
		return errors.New("product_id and environment_id are required")
	}
	var environmentCount int64
	if err := s.db.WithContext(ctx).Model(&models.ProductEnvironment{}).Where("product_id = ? AND environment_id = ? AND is_active = TRUE", productID, environmentID).Count(&environmentCount).Error; err != nil {
		return err
	}
	if environmentCount != 1 {
		return responses.ErrNotFound
	}
	actor, ok := actorFromContext(ctx)
	if !ok {
		return responses.ErrForbidden
	}
	if actor.PlatformAdmin {
		return nil
	}
	if s.authz == nil {
		return responses.ErrForbidden
	}
	product := productID
	environment := environmentID
	permission, err := s.authz.CheckPermission(globalauth.Actor{UserID: actor.UserID, PlatformAdmin: actor.PlatformAdmin}, dto.PermissionCheckRequest{ProductID: &product, EnvironmentID: &environment, ResourceType: "ARCHIVE", Action: action})
	if err != nil {
		return err
	}
	if !permission.Allowed {
		return responses.ErrForbidden
	}
	return nil
}

func (s *service) failJob(ctx context.Context, job *models.LogArchiveJob, err error) (*models.LogArchiveJob, []models.LogArchive, error) {
	message := err.Error()
	job.Status = "FAILED"
	job.ErrorMessage = &message
	_ = s.db.WithContext(ctx).Model(job).Updates(map[string]any{"status": job.Status, "error_message": message}).Error
	s.recordAudit(ctx, job.ProductID, "ARCHIVE_FAILED", nil, map[string]any{"environment_id": job.EnvironmentID, "error": message})
	return job, nil, err
}

func (s *service) recordAudit(ctx context.Context, productID int, action string, resourceID *string, metadata map[string]any) {
	actor, _ := actorFromContext(ctx)
	actorID := int64(actor.UserID)
	product := int64(productID)
	encoded, err := json.Marshal(metadata)
	if err != nil {
		encoded = []byte(`{}`)
	}
	result := models.AuditResultSuccess
	if strings.HasSuffix(action, "FAILED") {
		result = models.AuditResultFailed
	}
	createdAt := time.Now().UTC()
	value := &models.SystemAuditLog{AuditID: newUUID(), ActorUserID: &actorID, ProductID: &product, Action: action, ResourceType: "ARCHIVE", ResourceID: resourceID, Result: result, Metadata: encoded, CreatedAt: &createdAt}
	_ = s.db.WithContext(context.WithoutCancel(ctx)).Create(value).Error
}

func (s *service) setJobStatus(ctx context.Context, archiveID, status string) {
	var job models.LogArchiveJob
	needle := fmt.Sprintf(`["%s"]`, archiveID)
	if err := s.db.WithContext(ctx).Where("archive_ids @> ?", needle).Order("created_at DESC").First(&job).Error; err == nil {
		_ = s.db.WithContext(ctx).Model(&job).Update("status", status).Error
	}
}

func firstArchiveID(values []models.LogArchive) *string {
	if len(values) == 0 {
		return nil
	}
	value := values[0].ArchiveID
	return &value
}
func documentCount(values []models.LogArchive) int64 {
	var total int64
	for _, value := range values {
		total += value.DocumentCount
	}
	return total
}

func storageRoot() string {
	if value := strings.TrimSpace(os.Getenv("ARCHIVE_STORAGE_ROOT")); value != "" {
		return value
	}
	return "data"
}
func startOfDay(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func startOfDayIn(value time.Time, location *time.Location) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, location)
}

// calendarDate treats the Y-M-D sent by a date picker as a calendar value in
// the policy timezone, rather than shifting it because the JSON happened to
// carry a UTC offset.
func calendarDate(value time.Time, location *time.Location) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, location)
}
func firstIndex(values []string) *string {
	if len(values) != 1 {
		return nil
	}
	value := values[0]
	return &value
}
func intPtr(value int) *int              { return &value }
func stringPtr(value string) *string     { return &value }
func timePtr(value time.Time) *time.Time { return &value }

func buildSnapshotName(productID, environmentID int, day time.Time, archiveID string) string {
	suffix := strings.ReplaceAll(archiveID, "-", "")
	if len(suffix) > 12 {
		suffix = suffix[:12]
	}
	return fmt.Sprintf("omnilogs-p%d-e%d-%s-%s", productID, environmentID, day.UTC().Format("20060102"), suffix)
}

func archiveScopeMatches(archive models.LogArchive, req dto.CreateArchiveRequest) bool {
	// A retention policy is configuration, not the identity of the archived
	// logs. A policy can be deleted and recreated with a new policy_id while
	// the existing archive remains valid. Reuse an archive when its actual
	// log scope matches, regardless of which policy created it.
	return nullableIntEqual(archive.CategoryID, req.CategoryID) &&
		nullableIntEqual(archive.FeatureID, req.FeatureID) &&
		nullableIntEqual(archive.SubFeatureID, req.SubFeatureID)
}

func archiveNeedsRearchive(archive models.LogArchive) bool {
	// Restoring an empty archive is a successful no-op. There are no active
	// documents to archive again, so reuse the existing generation instead of
	// creating a guaranteed-empty replacement and reporting it as a failure.
	return archive.DocumentCount > 0 &&
		archive.RestoreStatus != nil &&
		strings.EqualFold(strings.TrimSpace(*archive.RestoreStatus), "RESTORED")
}

func archiveReusable(archive models.LogArchive, activeCount int64) bool {
	if archiveNeedsRearchive(archive) {
		return false
	}
	if archive.Status != nil && strings.EqualFold(*archive.Status, "ARCHIVED") {
		return true
	}
	if activeCount == 0 {
		return true
	}
	// A verified archive that intentionally retained active logs is already
	// complete. A verified archive whose active copy was deleted must not be
	// reused when new/late documents later appear in the same day.
	return archive.DeletedFromActiveAt == nil
}

func archiveDayQueryFrom(policy *models.LogRetentionPolicy, day, dayEnd time.Time, explicitRange bool) time.Time {
	if explicitRange || policy == nil || policy.ApplyToExistingLogs || policy.EffectiveFrom == nil {
		return day
	}
	effective := policy.EffectiveFrom.UTC()
	if effective.After(day) && effective.Before(dayEnd) {
		return effective
	}
	return day
}

func hasRestoredArchive(archives []models.LogArchive, req dto.CreateArchiveRequest) bool {
	for _, archive := range archives {
		if archiveScopeMatches(archive, req) && archiveNeedsRearchive(archive) {
			return true
		}
	}
	return false
}

func (s *service) markPreviousRestoresRearchived(ctx context.Context, archive *models.LogArchive) error {
	if archive == nil || archive.EnvironmentID == nil || archive.DateFrom == nil || archive.DateTo == nil {
		return nil
	}
	query := s.db.WithContext(ctx).
		Where("archive_id <> ? AND product_id = ? AND environment_id = ? AND date_from < ? AND date_to > ? AND restore_status = ?", archive.ArchiveID, archive.ProductID, *archive.EnvironmentID, archive.DateTo, archive.DateFrom, "RESTORED")
	query = applyArchiveScopeQuery(query, archive.CategoryID, archive.FeatureID, archive.SubFeatureID)
	var previous []models.LogArchive
	if err := query.Find(&previous).Error; err != nil {
		return err
	}
	if len(previous) == 0 {
		return nil
	}
	ids := make([]string, 0, len(previous))
	for _, item := range previous {
		ids = append(ids, item.ArchiveID)
	}
	if err := s.db.WithContext(ctx).Model(&models.LogArchive{}).Where("archive_id IN ?", ids).Updates(map[string]any{
		"status": "SUPERSEDED", "restore_status": "REARCHIVED", "file_path": nil, "object_key": nil, "checksum": nil,
	}).Error; err != nil {
		return err
	}
	for _, item := range previous {
		if item.FilePath != nil && strings.TrimSpace(*item.FilePath) != "" {
			_ = os.Remove(*item.FilePath)
		}
		if item.ObjectKey != nil && strings.TrimSpace(*item.ObjectKey) != "" {
			_ = s.r2.Delete(context.WithoutCancel(ctx), *item.ObjectKey)
		}
	}
	return nil
}

func nullableIntEqual(left, right *int) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func normalizeBackupType(value string) (string, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		value = "MANUAL"
	}
	if value != "POLICY" && value != "MANUAL" {
		return "", errors.New("backup_type must be POLICY or MANUAL")
	}
	return value, nil
}

func buildBackupTag(backupType string, policyID int, from, to time.Time) string {
	end := to.AddDate(0, 0, -1)
	return fmt.Sprintf("%s:p%d:%s_%s", strings.ToLower(backupType), policyID, from.Format("2006-01-02"), end.Format("2006-01-02"))
}

func archiveCoverageKey(productID, environmentID int, categoryID, featureID, subFeatureID *int, localDay time.Time) string {
	return fmt.Sprintf("p%d:e%d:c%s:f%s:s%s:d%s", productID, environmentID, nullableIntString(categoryID), nullableIntString(featureID), nullableIntString(subFeatureID), localDay.Format("2006-01-02"))
}

func nullableIntString(value *int) string {
	if value == nil {
		return "all"
	}
	return strconv.Itoa(*value)
}

func hasPortableArchive(archive models.LogArchive) bool {
	return (archive.FilePath != nil && strings.TrimSpace(*archive.FilePath) != "") ||
		(archive.ObjectKey != nil && strings.TrimSpace(*archive.ObjectKey) != "")
}

func (s *service) archiveLocalPath(ctx context.Context, archive models.LogArchive) (string, func(), error) {
	if archive.FilePath != nil && strings.TrimSpace(*archive.FilePath) != "" {
		return *archive.FilePath, func() {}, nil
	}
	if archive.ObjectKey != nil && strings.TrimSpace(*archive.ObjectKey) != "" {
		return s.r2.Materialize(ctx, *archive.ObjectKey)
	}
	return "", func() {}, errors.New("archive has no local file or R2 object")
}
func newUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "00000000-0000-4000-8000-000000000000"
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	out := make([]byte, 36)
	hex.Encode(out[0:8], b[0:4])
	out[8] = '-'
	hex.Encode(out[9:13], b[4:6])
	out[13] = '-'
	hex.Encode(out[14:18], b[6:8])
	out[18] = '-'
	hex.Encode(out[19:23], b[8:10])
	out[23] = '-'
	hex.Encode(out[24:36], b[10:16])
	return string(out)
}

package usecase

import (
	"context"
	"log/slog"
	"time"

	"omnilogs-api/dto"
	logarchiveservice "omnilogs-api/internal/log_archive/service"
	retentionschedule "omnilogs-api/internal/retention_policy/schedule"
	"omnilogs-api/models"

	"github.com/elastic/go-elasticsearch/v8"
	"gorm.io/gorm"
)

const (
	defaultPollInterval = 30 * time.Second
	defaultLeadTime     = 30 * time.Second
	retryDelay          = 5 * time.Minute
)

// Worker owns retention execution only. next_purge_at is the durable handoff
// from policy configuration (and survives process restarts), so no in-memory
// message is required between the ingestion and retention workers.
type Worker struct {
	db       *gorm.DB
	es       *elasticsearch.Client
	poll     time.Duration
	lead     time.Duration
	workerID string
}

func New(db *gorm.DB, es *elasticsearch.Client, pollInterval, leadTime time.Duration) *Worker {
	if pollInterval <= 0 {
		pollInterval = defaultPollInterval
	}
	if leadTime < 0 {
		leadTime = defaultLeadTime
	}
	return &Worker{db: db, es: es, poll: pollInterval, lead: leadTime, workerID: "retention-" + newID()}
}

func (w *Worker) Run(ctx context.Context) error {
	slog.Info("retention worker started", "worker_id", w.workerID, "poll_interval", w.poll, "lead_time", w.lead)
	for {
		if err := w.runDue(ctx); err != nil {
			slog.Error("retention worker cycle failed", "worker_id", w.workerID, "error", err)
		}
		wait := w.nextWait(ctx)
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}
	}
}

func (w *Worker) nextWait(ctx context.Context) time.Duration {
	var next *time.Time
	row := w.db.WithContext(ctx).Model(&models.LogRetentionPolicy{}).
		Where("is_active = TRUE AND archive_enabled = TRUE AND environment_id IS NOT NULL AND next_purge_at IS NOT NULL").
		Select("MIN(next_purge_at)").Scan(&next)
	if row.Error != nil || next == nil {
		return w.poll
	}
	wait := time.Until(next.UTC().Add(-w.lead))
	if wait < time.Second {
		return time.Second
	}
	if wait > w.poll {
		return w.poll
	}
	return wait
}

func (w *Worker) runDue(ctx context.Context) error {
	now := time.Now().UTC()
	var policies []models.LogRetentionPolicy
	if err := w.db.WithContext(ctx).
		Where("is_active = TRUE AND environment_id IS NOT NULL AND archive_enabled = TRUE").
		Where("next_purge_at IS NULL OR next_purge_at <= ?", now).
		Order("next_purge_at ASC NULLS FIRST").Find(&policies).Error; err != nil {
		return err
	}
	archiveService := logarchiveservice.New(w.db, w.es, nil)
	for _, policy := range policies {
		if err := w.runPolicy(ctx, archiveService, policy); err != nil {
			slog.Error("retention policy failed", "worker_id", w.workerID, "policy_id", policy.PolicyID, "error", err)
			// Prevent a bad policy or unavailable dependency from hot-looping.
			_ = w.db.WithContext(ctx).Model(&policy).Update("next_purge_at", time.Now().UTC().Add(retryDelay)).Error
		}
	}
	return nil
}

func (w *Worker) runPolicy(ctx context.Context, archiveService logarchiveservice.Service, policy models.LogRetentionPolicy) error {
	if policy.EnvironmentID == nil {
		return nil
	}
	actorCtx := logarchiveservice.WithActor(ctx, logarchiveservice.Actor{PlatformAdmin: true})
	policyID := policy.PolicyID
	job, archives, err := archiveService.Create(actorCtx, dto.CreateArchiveRequest{ProductID: policy.ProductID, EnvironmentID: *policy.EnvironmentID, PolicyID: &policyID, BackupType: "POLICY"})
	if err != nil {
		return err
	}
	for _, archive := range archives {
		if _, err := archiveService.Verify(actorCtx, policy.ProductID, *policy.EnvironmentID, archive.ArchiveID, &policyID); err != nil {
			return err
		}
	}
	if err := archiveService.PurgeExpired(actorCtx, policy.ProductID, *policy.EnvironmentID, valueOrZero(policy.ArchiveRetentionValue), policy.ArchiveRetentionUnit, policy.ArchiveNeverDelete); err != nil {
		return err
	}
	next, err := retentionschedule.Next(policy.CronSchedule, policy.ScheduleTimezone, time.Now().UTC())
	if err != nil {
		return err
	}
	if err := w.db.WithContext(ctx).Model(&policy).Updates(map[string]any{"last_purge_at": time.Now().UTC(), "next_purge_at": next}).Error; err != nil {
		return err
	}
	slog.Info("retention policy completed", "worker_id", w.workerID, "policy_id", policy.PolicyID, "job_id", job.JobID, "archives_count", len(archives), "next_purge_at", next)
	return nil
}

func valueOrZero(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func newID() string { return time.Now().UTC().Format("20060102150405.000000000") }

package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"omnilogs-api/configs"
	"omnilogs-api/internal/archive_utils"
	"omnilogs-api/internal/queue"
	workerrepo "omnilogs-api/internal/worker/repository"
	"omnilogs-api/models"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/nats-io/nats.go"
)

const DefaultPollInterval = 2 * time.Second

var ErrQueueDisabled = errors.New("nats queue is not configured")

type Usecase interface {
	Run(ctx context.Context) error
	RunOnce(ctx context.Context) (bool, error)
}

type usecase struct {
	repo               workerrepo.Repository
	esClient           *elasticsearch.Client
	workerID           string
	pollInterval       time.Duration
	encryptionKey      string
	lastRetentionCheck time.Time
	natsQueue          *configs.NATSQueue
	subscription       *nats.Subscription
}

func NewUsecase(repo workerrepo.Repository, esClient *elasticsearch.Client, encryptionKey string, natsQueue *configs.NATSQueue) Usecase {
	return &usecase{
		repo:               repo,
		esClient:           esClient,
		workerID:           "worker-" + newUUID(),
		pollInterval:       DefaultPollInterval,
		encryptionKey:      encryptionKey,
		lastRetentionCheck: time.Time{},
		natsQueue:          natsQueue,
	}
}

func (u *usecase) Run(ctx context.Context) error {
	if u.natsQueue == nil {
		return ErrQueueDisabled
	}

	ticker := time.NewTicker(u.pollInterval)
	defer ticker.Stop()

	for {
		u.runPeriodicMaintenance(ctx)

		if _, err := u.RunOnce(ctx); err != nil {
			slog.Error("jetstream batch processing failed",
				slog.String("worker_id", u.workerID),
				slog.Any("error", err),
			)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (u *usecase) RunOnce(ctx context.Context) (bool, error) {
	if u.natsQueue == nil {
		return false, ErrQueueDisabled
	}

	u.runPeriodicMaintenance(ctx)

	sub, err := u.getSubscription()
	if err != nil {
		return false, err
	}

	msgs, err := sub.Fetch(u.natsQueue.FetchBatchSize, nats.MaxWait(u.natsQueue.FetchMaxWait))
	if err != nil {
		if errors.Is(err, nats.ErrTimeout) {
			return false, nil
		}
		return false, err
	}
	if len(msgs) == 0 {
		return false, nil
	}

	return true, u.processFetchedMessages(ctx, msgs)
}

func (u *usecase) runPeriodicMaintenance(ctx context.Context) {
	if time.Since(u.lastRetentionCheck) > 30*time.Second {
		if err := u.runRetentionCheck(ctx); err != nil {
			slog.Error("retention check failed", "error", err)
		} else {
			u.lastRetentionCheck = time.Now()
		}
	}

}

func (u *usecase) getSubscription() (*nats.Subscription, error) {
	if u.subscription != nil {
		return u.subscription, nil
	}

	sub, err := u.natsQueue.PullSubscribe()
	if err != nil {
		return nil, err
	}
	u.subscription = sub
	return sub, nil
}

func (u *usecase) runRetentionCheck(ctx context.Context) error {
	policies, err := u.repo.GetActiveIndexPolicies(ctx)
	if err != nil {
		return err
	}

	for _, policy := range policies {
		if policy.RetentionDays == nil || *policy.RetentionDays <= 0 {
			continue
		}

		prefix := strings.ToLower(strings.TrimSpace(policy.IndexPrefix))
		if prefix == "" {
			prefix = fmt.Sprintf("omnilogs-product-%d", policy.ProductID)
		}

		targets := []string{prefix + "-*"}
		defaultPrefix := fmt.Sprintf("omnilogs-product-%d", policy.ProductID)
		if prefix != defaultPrefix {
			targets = append(targets, defaultPrefix+"-*")
		}

		res, err := u.esClient.Indices.Get(
			targets,
			u.esClient.Indices.Get.WithContext(ctx),
		)
		if err != nil {
			slog.Error("failed to get indices from ES", "prefix", prefix, "error", err)
			continue
		}

		func() {
			defer res.Body.Close()

			if res.IsError() {
				if res.StatusCode != 404 {
					slog.Error("ES returned error getting indices", "prefix", prefix, "status", res.StatusCode)
				}
				return
			}

			var info map[string]any
			if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
				slog.Error("failed to decode ES response", "error", err)
				return
			}

			retentionDays := *policy.RetentionDays
			cutoff := time.Now().UTC().AddDate(0, 0, -retentionDays)
			today := time.Now().UTC().Truncate(24 * time.Hour)

			for indexName := range info {
				parts := strings.Split(indexName, "-")
				if len(parts) < 2 {
					continue
				}

				dateStr := parts[len(parts)-1]
				date, err := time.Parse("2006.01.02", dateStr)
				if err != nil {
					continue
				}
				if !date.Before(today) || !date.Before(cutoff) {
					continue
				}

				slog.Info("index expired, moving to archive and deleting", "index", indexName, "cutoff", cutoff)

				var ingestPolicy models.LogIngestionPolicy
				hasIngestPolicy := true
				err = u.repo.DB().WithContext(ctx).
					Where("product_id = ? AND (environment_id = ? OR environment_id IS NULL)", policy.ProductID, policy.EnvironmentID).
					Order("environment_id DESC NULLS LAST").
					First(&ingestPolicy).Error
				if err != nil {
					hasIngestPolicy = false
				}

				archiveEnabled := false
				archiveFormat := "JSON"
				archiveStoragePath := fmt.Sprintf("data/archives/product-%d", policy.ProductID)

				if hasIngestPolicy {
					archiveEnabled = ingestPolicy.ArchiveEnabled
					if ingestPolicy.ArchiveFormat != nil && *ingestPolicy.ArchiveFormat != "" {
						archiveFormat = *ingestPolicy.ArchiveFormat
					}
					if ingestPolicy.ArchiveStoragePath != nil && *ingestPolicy.ArchiveStoragePath != "" {
						archiveStoragePath = *ingestPolicy.ArchiveStoragePath
					}
				}

				lockRes, lockErr := u.esClient.Indices.PutSettings(
					strings.NewReader(`{"index":{"blocks.write":true}}`),
					u.esClient.Indices.PutSettings.WithIndex(indexName),
					u.esClient.Indices.PutSettings.WithContext(ctx),
				)
				if lockErr == nil {
					lockRes.Body.Close()
				}

				if archiveEnabled {
					totalLogs, compressedSize, localPath, archiveErr := archive_utils.ArchiveIndex(ctx, u.repo.DB(), u.esClient, indexName, policy.ProductID, archiveFormat, archiveStoragePath)
					if archiveErr != nil {
						slog.Error("failed to archive index", "index", indexName, "error", archiveErr)
						unlockRes, unlockErr := u.esClient.Indices.PutSettings(
							strings.NewReader(`{"index":{"blocks.write":false}}`),
							u.esClient.Indices.PutSettings.WithIndex(indexName),
							u.esClient.Indices.PutSettings.WithContext(ctx),
						)
						if unlockErr == nil {
							unlockRes.Body.Close()
						}
						continue
					}

					now := time.Now()
					dateStart := date.UTC()
					dateEnd := date.UTC().Add(24*time.Hour - time.Second)
					status, provider := "COMPLETED", "LOCAL"

					archive := &models.LogArchive{
						ArchiveID:           newUUID(),
						ProductID:           policy.ProductID,
						EnvironmentID:       policy.EnvironmentID,
						ArchiveYear:         date.Year(),
						ArchiveMonth:        int(date.Month()),
						DateFrom:            &dateStart,
						DateTo:              &dateEnd,
						StorageProvider:     provider,
						FilePath:            &localPath,
						FileFormat:          &archiveFormat,
						TotalLogs:           &totalLogs,
						CompressedSizeBytes: &compressedSize,
						Status:              &status,
						ExportedAt:          &now,
					}

					if err := u.repo.CreateLogArchive(ctx, archive); err != nil {
						slog.Error("failed to create log archive record", "index", indexName, "error", err)
						continue
					}
				}

				delRes, err := u.esClient.Indices.Delete(
					[]string{indexName},
					u.esClient.Indices.Delete.WithContext(ctx),
				)
				if err != nil {
					slog.Error("failed to delete expired index from ES", "index", indexName, "error", err)
					continue
				}
				delRes.Body.Close()
			}
		}()
	}
	return nil
}

func decodeMessage(msg *nats.Msg) (queue.LogMessage, error) {
	var envelope queue.LogMessage
	if err := json.Unmarshal(msg.Data, &envelope); err != nil {
		return queue.LogMessage{}, err
	}
	return envelope, nil
}

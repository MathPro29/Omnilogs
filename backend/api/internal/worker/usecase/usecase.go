package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"omnilogs-api/configs"
	"omnilogs-api/dto"
	workerrepo "omnilogs-api/internal/worker/repository"

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
	repo          workerrepo.Repository
	esClient      *elasticsearch.Client
	workerID      string
	pollInterval  time.Duration
	encryptionKey string
	natsQueue     *configs.NATSQueue
	subscription  *nats.Subscription
}

func NewUsecase(repo workerrepo.Repository, esClient *elasticsearch.Client, encryptionKey string, natsQueue *configs.NATSQueue) Usecase {
	return &usecase{
		repo:          repo,
		esClient:      esClient,
		workerID:      "worker-" + newUUID(),
		pollInterval:  DefaultPollInterval,
		encryptionKey: encryptionKey,
		natsQueue:     natsQueue,
	}
}

func (u *usecase) Run(ctx context.Context) error {
	if u.natsQueue == nil {
		return ErrQueueDisabled
	}

	ticker := time.NewTicker(u.pollInterval)
	defer ticker.Stop()

	for {
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

func decodeMessage(msg *nats.Msg) (dto.LogMessage, error) {
	var envelope dto.LogMessage
	if err := json.Unmarshal(msg.Data, &envelope); err != nil {
		return dto.LogMessage{}, err
	}
	return envelope, nil
}

package worker

import (
	"context"

	"omnilogs-api/configs"
	workerhandler "omnilogs-api/internal/worker/handler"
	workerrepo "omnilogs-api/internal/worker/repository"
	workerusecase "omnilogs-api/internal/worker/usecase"

	"github.com/elastic/go-elasticsearch/v8"
	"gorm.io/gorm"
)

type Processor struct {
	handler *workerhandler.Handler
}

func NewProcessor(db *gorm.DB, esClient *elasticsearch.Client, encryptionKey string, natsQueue *configs.NATSQueue) *Processor {
	repo := workerrepo.NewRepository(db)
	usecase := workerusecase.NewUsecase(repo, esClient, encryptionKey, natsQueue)
	return &Processor{
		handler: workerhandler.NewHandler(usecase),
	}
}

func (p *Processor) Run(ctx context.Context) error {
	return p.handler.Run(ctx)
}

func (p *Processor) RunOnce(ctx context.Context) (bool, error) {
	return p.handler.RunOnce(ctx)
}

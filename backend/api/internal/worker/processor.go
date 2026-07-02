package worker

import (
	"context"

	workerhandler "omnilogs-api/internal/worker/handler"
	workerrepo "omnilogs-api/internal/worker/repository"
	workerusecase "omnilogs-api/internal/worker/usecase"

	"github.com/elastic/go-elasticsearch/v8"
	"gorm.io/gorm"
)

type Processor struct {
	handler *workerhandler.Handler
}

func NewProcessor(db *gorm.DB, esClient *elasticsearch.Client, encryptionKey string) *Processor {
	repo := workerrepo.NewRepository(db)
	usecase := workerusecase.NewUsecase(repo, esClient, encryptionKey)
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

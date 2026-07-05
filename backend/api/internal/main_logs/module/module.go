package mainlogs

import (
	auditrepo "omnilogs-api/internal/audit_logs/repository"
	auditusecase "omnilogs-api/internal/audit_logs/usecase"
	"omnilogs-api/internal/main_logs/handler"
	mainlogusecase "omnilogs-api/internal/main_logs/usecase"

	"github.com/elastic/go-elasticsearch/v8"
	"gorm.io/gorm"
)

func NewHandler(db *gorm.DB, esClient *elasticsearch.Client, encryptionKey string) *handler.Handler {
	repo := auditrepo.NewRepository(db)
	audits := auditusecase.NewUsecase(repo)
	mainLogs := mainlogusecase.NewUsecase(db, esClient, audits, encryptionKey)
	return handler.NewHandler(mainLogs)
}

package mainlogs

import (
	auditrepo "omnilogs-api/internal/audit_logs/repository"
	auditusecase "omnilogs-api/internal/audit_logs/usecase"
	"omnilogs-api/internal/main_logs/handler"
	mainlogrepo "omnilogs-api/internal/main_logs/repository"
	mainlogusecase "omnilogs-api/internal/main_logs/usecase"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"gorm.io/gorm"
)

func NewHandler(db *gorm.DB, esClient *elasticsearch.Client, encryptionKey string, queryTimeout, slowQueryThreshold time.Duration) *handler.Handler {
	auditRepo := auditrepo.NewRepository(db)
	audits := auditusecase.NewUsecase(auditRepo)

	mainLogRepo := mainlogrepo.NewRepository(db, esClient, queryTimeout, slowQueryThreshold)
	mainLogs := mainlogusecase.NewUsecase(mainLogRepo, audits, encryptionKey)
	return handler.NewHandler(mainLogs)
}

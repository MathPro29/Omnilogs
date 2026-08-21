package auditlogs

import (
	"omnilogs-api/internal/audit_logs/handler"
	"omnilogs-api/internal/audit_logs/repository"
	"omnilogs-api/internal/audit_logs/usecase"

	"gorm.io/gorm"
)

func NewHandler(db *gorm.DB) *handler.Handler {
	repo := repository.NewRepository(db)
	auditUsecase := usecase.NewUsecase(repo)
	return handler.NewHandler(auditUsecase)
}

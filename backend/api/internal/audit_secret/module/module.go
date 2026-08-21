package module

import (
	"omnilogs-api/internal/audit_secret/handler"
	"omnilogs-api/internal/audit_secret/repository"
	"omnilogs-api/internal/audit_secret/usecase"

	"gorm.io/gorm"
)

func NewHandler(db *gorm.DB, encryptionKey string) *handler.Handler {
	repo := repository.NewRepository(db)
	uc := usecase.NewUsecase(repo, encryptionKey)
	return handler.NewHandler(uc)
}

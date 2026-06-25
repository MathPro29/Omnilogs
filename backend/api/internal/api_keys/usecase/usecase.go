package usecase

import (
	"errors"

	"omnilogs-api/dto"
	"omnilogs-api/internal/api_keys/repository"
)

var (
	ErrNotFound  = errors.New("api key resource not found")
	ErrForbidden = errors.New("api key access denied")
	ErrConflict  = errors.New("api key resource already exists")
	ErrInvalid   = errors.New("invalid api key payload")
)

type Actor struct {
	UserID        int
	PlatformAdmin bool
}

type Usecase interface {
	CreateAPIKey(Actor, int, dto.CreateAPIKeyRequest) (*dto.CreateAPIKeyResponse, error)
	ListAPIKeys(Actor, int) ([]dto.APIKeyResponse, error)
	UpdateAPIKey(Actor, int, int, dto.UpdateAPIKeyRequest) (*dto.APIKeyResponse, error)
	RevokeAPIKey(Actor, int, int) error
}

type usecase struct {
	repository repository.Repository
}

func NewUsecase(repository repository.Repository) Usecase {
	return &usecase{repository: repository}
}

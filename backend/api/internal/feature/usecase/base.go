package usecase

import (
	"errors"

	"omnilogs-api/dto"
	"omnilogs-api/internal/feature/repository"
	"omnilogs-api/models"
)

var (
	ErrNotFound  = errors.New("feature resource not found")
	ErrForbidden = errors.New("feature access denied")
	ErrConflict  = errors.New("feature resource already exists")
	ErrInvalid   = errors.New("invalid feature relationship")
)

type Actor struct {
	UserID        int
	PlatformAdmin bool
}

type AccessTarget struct {
	ProductID  int
	ProjectID  *int
	CategoryID *int
}

type Usecase interface {
	CreateFeature(Actor, int, int, dto.CreateProjectFeatureRequest) (*models.ProjectFeature, error)
	ListFeatures(Actor, int, int) ([]models.ProjectFeature, error)
	UpdateFeature(Actor, int, int, int, dto.UpdateProjectFeatureRequest) (*models.ProjectFeature, error)
	DeleteFeature(Actor, int, int, int) error
}

type usecase struct{ repository repository.Repository }

func NewUsecase(repository repository.Repository) Usecase { return &usecase{repository: repository} }

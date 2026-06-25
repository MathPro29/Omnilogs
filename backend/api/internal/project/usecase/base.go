package usecase

import (
	"errors"

	"omnilogs-api/dto"
	"omnilogs-api/internal/project/repository"
	"omnilogs-api/models"
)

var (
	ErrNotFound  = errors.New("project resource not found")
	ErrForbidden = errors.New("project access denied")
	ErrConflict  = errors.New("project resource already exists")
	ErrInvalid   = errors.New("invalid project relationship")
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
	CreateProject(Actor, int, dto.CreateProjectRequest) (*models.Project, error)
	ListProjects(Actor, int) ([]models.Project, error)
	GetProject(Actor, int, int) (*models.Project, error)
	UpdateProject(Actor, int, int, dto.UpdateProjectRequest) (*models.Project, error)
}

type usecase struct{ repository repository.Repository }

func NewUsecase(repository repository.Repository) Usecase { return &usecase{repository: repository} }

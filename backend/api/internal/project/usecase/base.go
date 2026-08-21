package usecase

import (
	"omnilogs-api/dto"
	"omnilogs-api/internal/project/repository"
	"omnilogs-api/models"
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
	DeleteProject(Actor, int, int) error
}

type usecase struct{ repository repository.Repository }

func NewUsecase(repository repository.Repository) Usecase { return &usecase{repository: repository} }

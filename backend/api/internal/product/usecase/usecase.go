package usecase

import (
	"omnilogs-api/dto"
	"omnilogs-api/internal/product/repository"
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
	CreateProduct(Actor, dto.CreateProductRequest) (*models.Product, error)
	ListProducts(Actor) ([]models.Product, error)
	GetProduct(Actor, int) (*models.Product, error)
	UpdateProduct(Actor, int, dto.UpdateProductRequest) (*models.Product, error)
	DeleteProduct(Actor, int) error
	RestoreProduct(Actor, int) (*models.Product, error)
	BulkDeleteProducts(Actor, []int) error
	GetSetupStatus(Actor, int) (*dto.SetupStatusResponse, error)
	ValidateSetupStructure(Actor, int) (bool, []string, error)
	CompleteSetup(Actor, int) error
}

type usecase struct{ repository repository.Repository }

func NewUsecase(repository repository.Repository) Usecase { return &usecase{repository: repository} }

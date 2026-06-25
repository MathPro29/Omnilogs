package usecase

import (
	"errors"

	"omnilogs-api/dto"
	"omnilogs-api/internal/product/repository"
	"omnilogs-api/models"
)

var (
	ErrNotFound  = errors.New("product resource not found")
	ErrForbidden = errors.New("product access denied")
	ErrConflict  = errors.New("product resource already exists")
	ErrInvalid   = errors.New("invalid product relationship")
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
}

type usecase struct{ repository repository.Repository }

func NewUsecase(repository repository.Repository) Usecase { return &usecase{repository: repository} }

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
	/// Product
	CreateProduct(Actor, dto.CreateProductRequest) (*models.Product, error)
	ListProducts(Actor) ([]models.Product, error)
	GetProduct(Actor, int) (*models.Product, error)
	UpdateProduct(Actor, int, dto.UpdateProductRequest) (*models.Product, error)
	/// API Key
	CreateAPIKey(Actor, int, dto.CreateAPIKeyRequest) (*dto.CreateAPIKeyResponse, error)
	ListAPIKeys(Actor, int) ([]dto.APIKeyResponse, error)
	UpdateAPIKey(Actor, int, int, dto.UpdateAPIKeyRequest) (*dto.APIKeyResponse, error)
	RevokeAPIKey(Actor, int, int) error
	/// Project
	CreateProject(Actor, int, dto.CreateProjectRequest) (*models.Project, error)
	ListProjects(Actor, int) ([]models.Project, error)
	GetProject(Actor, int, int) (*models.Project, error)
	UpdateProject(Actor, int, int, dto.UpdateProjectRequest) (*models.Project, error)
	/// Feature
	CreateFeature(Actor, int, int, dto.CreateProjectFeatureRequest) (*models.ProjectFeature, error)
	ListFeatures(Actor, int, int) ([]models.ProjectFeature, error)
	UpdateFeature(Actor, int, int, int, dto.UpdateProjectFeatureRequest) (*models.ProjectFeature, error)
	DeleteFeature(Actor, int, int, int) error
	/// Role
	CreateRole(Actor, int, dto.CreateProductRoleRequest) (*models.ProductRole, error)
	ListRoles(Actor, int) ([]models.ProductRole, error)
	UpdateRole(Actor, int, int, dto.UpdateRoleRequest) (*models.ProductRole, error)
	DeleteRole(Actor, int, int) error
	/// Membership
	CreateMembership(Actor, int, dto.CreateProductMembershipRequest) (*models.ProductMembership, error)
	ListMemberships(Actor, int) ([]models.ProductMembership, error)
	UpdateMembership(Actor, int, int, dto.UpdateProductMembershipRequest) (*models.ProductMembership, error)
	DeleteMembership(Actor, int, int) error
	/// Permission
	CreatePermissionRule(Actor, dto.CreatePermissionRuleRequest) (*models.UserRolePermissionRule, error)
	ListPermissionRules(Actor, int) ([]models.UserRolePermissionRule, error)
	UpdatePermissionRule(Actor, int, dto.UpdatePermissionRuleRequest) (*models.UserRolePermissionRule, error)
	CheckPermission(Actor, dto.PermissionCheckRequest) (*dto.PermissionCheckResponse, error)
}

type usecase struct{ repository repository.Repository }

func NewUsecase(repository repository.Repository) Usecase { return &usecase{repository: repository} }

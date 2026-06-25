package usecase

import (
	"errors"

	"omnilogs-api/dto"
	"omnilogs-api/internal/global_auth/repository"
	"omnilogs-api/models"
)

var (
	ErrNotFound  = errors.New("access resource not found")
	ErrForbidden = errors.New("access denied")
	ErrConflict  = errors.New("access resource already exists")
	ErrInvalid   = errors.New("invalid access relationship")
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
	CreateRole(Actor, int, dto.CreateProductRoleRequest) (*models.ProductRole, error)
	ListRoles(Actor, int) ([]models.ProductRole, error)
	UpdateRole(Actor, int, int, dto.UpdateRoleRequest) (*models.ProductRole, error)
	DeleteRole(Actor, int, int) error
	CreateMembership(Actor, int, dto.CreateProductMembershipRequest) (*models.ProductMembership, error)
	ListMemberships(Actor, int) ([]models.ProductMembership, error)
	UpdateMembership(Actor, int, int, dto.UpdateProductMembershipRequest) (*models.ProductMembership, error)
	DeleteMembership(Actor, int, int) error
	CreatePermissionRule(Actor, dto.CreatePermissionRuleRequest) (*models.UserRolePermissionRule, error)
	ListPermissionRules(Actor, int) ([]models.UserRolePermissionRule, error)
	UpdatePermissionRule(Actor, int, dto.UpdatePermissionRuleRequest) (*models.UserRolePermissionRule, error)
	CheckPermission(Actor, dto.PermissionCheckRequest) (*dto.PermissionCheckResponse, error)
}

type usecase struct{ repository repository.Repository }

func NewUsecase(repository repository.Repository) Usecase { return &usecase{repository: repository} }

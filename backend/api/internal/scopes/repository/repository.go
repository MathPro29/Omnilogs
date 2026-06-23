package repository

import (
	"omnilogs-api/models"

	"gorm.io/gorm"
)

type Repository interface {
	CreateScope(*models.ProductMembershipScope) error
	ListScopes(productID, membershipID int) ([]models.ProductMembershipScope, error)
	UpdateScope(*models.ProductMembershipScope) error
	DeleteScope(*models.ProductMembershipScope) error
	GetScopeByID(id int) (*models.ProductMembershipScope, error)
}

type repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateScope(scope *models.ProductMembershipScope) error {
	return r.db.Create(scope).Error
}

func (r *repository) ListScopes(productID, membershipID int) ([]models.ProductMembershipScope, error) {
	var scopes []models.ProductMembershipScope
	return scopes, r.db.Where("product_id = ? AND membership_id = ?", productID, membershipID).Order("scope_id").Find(&scopes).Error
}

func (r *repository) UpdateScope(scope *models.ProductMembershipScope) error {
	return r.db.Save(scope).Error
}

func (r *repository) DeleteScope(scope *models.ProductMembershipScope) error {
	return r.db.Delete(scope).Error
}

func (r *repository) GetScopeByID(id int) (*models.ProductMembershipScope, error) {
	var scope models.ProductMembershipScope
	return &scope, r.db.First(&scope, id).Error
}
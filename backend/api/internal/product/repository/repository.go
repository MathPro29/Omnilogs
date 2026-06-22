package repository

import "gorm.io/gorm"

// Repository intentionally exposes transaction boundaries to the product
// usecase. Product operations span roles, memberships, scopes, projects and
// features and must be committed atomically.
type Repository interface {
	DB() *gorm.DB
	Transaction(func(*gorm.DB) error) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) DB() *gorm.DB { return r.db }

func (r *repository) Transaction(fn func(*gorm.DB) error) error {
	return r.db.Transaction(fn)
}

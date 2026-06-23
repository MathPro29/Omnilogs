package repository

import "gorm.io/gorm"

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

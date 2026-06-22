package repository

import (
	"omnilogs-api/models"

	"gorm.io/gorm"
)

type Repository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(userID uint) (*models.User, error)
	ListAllUsers() ([]models.User, error)
	UpdateRoleID(userID uint, roleID uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(user *models.User) error {
	if err := r.db.Create(user).Error; err != nil {

		return err
	}
	return r.db.Preload("Role").First(user, user.ID).Error
}

func (r *repository) FindByEmail(email string) (*models.User, error) {
	var user models.User

	if err := r.db.Preload("Role").Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindByID(userID uint) (*models.User, error) {
	var user models.User
	if err := r.db.Preload("Role").First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) UpdateRoleID(userID uint, roleID uint) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("role_id", roleID).Error
}

func (r *repository) ListAllUsers() ([]models.User, error) {
	var users []models.User
	return users, r.db.Preload("Role").Find(&users).Error
}

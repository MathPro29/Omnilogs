package repository

import (
	"omnilogs-api/models"
)

func (r *repository) Create(user *models.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return err
	}

	defaultRoleID, err := r.getOrCreateRegistrationRoleID()
	if err != nil {
		return err
	}

	membership := models.PlatformMembership{
		UserID:         user.UserID,
		PlatformRoleID: defaultRoleID,
		IsActive:       true,
	}
	if err := r.db.Create(&membership).Error; err != nil {
		return err
	}

	return r.populateUserRole(user)
}

func (r *repository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	if err := r.populateUserRole(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindByIdentifier(identifier string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ? OR username = ?", identifier, identifier).First(&user).Error; err != nil {
		return nil, err
	}
	if err := r.populateUserRole(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	if err := r.populateUserRole(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindByID(userID uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	if err := r.populateUserRole(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) ListAllUsers() ([]models.User, error) {
	var users []models.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	for i := range users {
		if err := r.populateUserRole(&users[i]); err != nil {
			return nil, err
		}
	}
	return users, nil
}

func (r *repository) UpdateUserField(userID uint, fieldName string, value any) error {
	return r.db.Model(&models.User{}).
		Where("user_id = ?", userID).
		Update(fieldName, value).Error
}

func (r *repository) DeleteUserByID(userID uint) error {
	return r.db.Delete(&models.User{}, userID).Error
}

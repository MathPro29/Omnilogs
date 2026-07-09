package repository

import (
	"time"

	"omnilogs-api/models"
)

func (r *repository) CreatePasswordResetToken(token *models.PasswordResetToken) error {
	return r.db.Create(token).Error
}

func (r *repository) FindActivePasswordResetToken(hash string) (*models.PasswordResetToken, error) {
	var token models.PasswordResetToken
	if err := r.db.Where("token_hash = ? AND used_at IS NULL", hash).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *repository) MarkPasswordResetTokenUsed(id int, usedAt time.Time) error {
	return r.db.Model(&models.PasswordResetToken{}).
		Where("password_reset_token_id = ?", id).
		Update("used_at", usedAt).Error
}

func (r *repository) UpdatePassword(userID uint, passwordHash string) error {
	return r.db.Model(&models.User{}).
		Where("user_id = ?", userID).
		Update("password_hash", passwordHash).Error
}

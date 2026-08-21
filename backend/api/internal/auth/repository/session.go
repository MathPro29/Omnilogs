package repository

import (
	"time"

	"omnilogs-api/models"
)

func (r *repository) CreateSession(session *models.AuthSession) error {
	return r.db.Create(session).Error
}

func (r *repository) FindActiveSessionByRefreshTokenHash(hash string) (*models.AuthSession, error) {
	var session models.AuthSession
	if err := r.db.Where("refresh_token_hash = ? AND revoked_at IS NULL", hash).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *repository) RevokeSessionByRefreshTokenHash(hash string, revokedAt time.Time) error {
	return r.db.Model(&models.AuthSession{}).
		Where("refresh_token_hash = ? AND revoked_at IS NULL", hash).
		Updates(map[string]any{"revoked_at": revokedAt, "last_used_at": revokedAt}).Error
}

func (r *repository) TouchSession(sessionID int, lastUsedAt time.Time) error {
	return r.db.Model(&models.AuthSession{}).
		Where("session_id = ?", sessionID).
		Update("last_used_at", lastUsedAt).Error
}

func (r *repository) RevokeAllUserSessions(userID uint, revokedAt time.Time) error {
	return r.db.Model(&models.AuthSession{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Updates(map[string]any{"revoked_at": revokedAt, "last_used_at": revokedAt}).Error
}

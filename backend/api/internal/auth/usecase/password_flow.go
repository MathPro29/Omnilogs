package usecase

import (
	"errors"
	"time"

	"omnilogs-api/dto"
	"omnilogs-api/models"
	"omnilogs-api/responses"
	"omnilogs-api/utils"

	"gorm.io/gorm"
)

func (u *usecase) ForgotPassword(req dto.ForgotPasswordRequest) (*dto.ForgotPasswordResponse, error) {
	user, err := u.repo.FindByEmail(req.Email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &dto.ForgotPasswordResponse{Message: "if the email exists, a reset token has been issued"}, nil
	}
	if err != nil {
		return nil, err
	}

	rawToken, err := u.opaqueTokenService.NewToken(32)
	if err != nil {
		return nil, err
	}
	token := &models.PasswordResetToken{
		UserID:    user.UserID,
		TokenHash: utils.SHA256Hex(rawToken),
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
	if err := u.repo.CreatePasswordResetToken(token); err != nil {
		return nil, err
	}
	return &dto.ForgotPasswordResponse{
		Message:    "password reset token created",
		ResetToken: rawToken,
	}, nil
}

func (u *usecase) ResetPassword(req dto.ResetPasswordRequest) error {
	token, err := u.repo.FindActivePasswordResetToken(utils.SHA256Hex(req.Token))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return responses.ErrorUserCode["INVALID_RESET_TOKEN"]
	}
	if err != nil {
		return err
	}
	if token.UsedAt != nil || token.ExpiresAt.Before(time.Now()) {
		return responses.ErrorUserCode["INVALID_RESET_TOKEN"]
	}

	passwordHash, err := u.passwordService.Hash(req.NewPassword)
	if err != nil {
		return err
	}
	if err := u.repo.UpdatePassword(uint(token.UserID), passwordHash); err != nil {
		return err
	}
	if err := u.repo.MarkPasswordResetTokenUsed(token.PasswordResetTokenID, time.Now()); err != nil {
		return err
	}
	return u.repo.RevokeAllUserSessions(uint(token.UserID), time.Now())
}

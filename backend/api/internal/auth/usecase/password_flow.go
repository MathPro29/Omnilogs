package usecase

import (
	"errors"
	"time"

	"omnilogs-api/dto"
	"omnilogs-api/responses"

	"gorm.io/gorm"
)

func (u *usecase) ForgotPassword(req dto.ForgotPasswordRequest) (*dto.ForgotPasswordResponse, error) {
	_, err := u.repo.FindByEmail(req.Email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &dto.ForgotPasswordResponse{Message: "if the email exists, password reset is available"}, nil
	}
	if err != nil {
		return nil, err
	}
	return &dto.ForgotPasswordResponse{Message: "password reset is available"}, nil
}

func (u *usecase) ResetPassword(req dto.ResetPasswordRequest) error {
	user, err := u.repo.FindByEmail(req.Email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return responses.ErrorUserCode["USER_NOT_FOUND"]
	}
	if err != nil {
		return err
	}
	passwordHash, err := u.passwordService.Hash(req.NewPassword)
	if err != nil {
		return err
	}
	if err := u.repo.UpdatePassword(uint(user.UserID), passwordHash); err != nil {
		return err
	}
	return u.repo.RevokeAllUserSessions(uint(user.UserID), time.Now())
}

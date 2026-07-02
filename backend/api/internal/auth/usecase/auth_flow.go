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

func (u *usecase) Register(req dto.RegisterRequest) (*dto.UserResponse, error) {
	if req.Username != nil && *req.Username != "" {
		if err := u.CheckUserExistsByUsername(*req.Username); err != nil {
			return nil, err
		}
	}

	if _, err := u.repo.FindByEmail(req.Email); err == nil {
		return nil, responses.ErrorUserCode["EMAIL_ALREADY_EXISTS"]
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	PasswordHash, err := u.passwordService.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Username:     req.Username,
		Email:        req.Email,
		RoleID:       defaultUserRoleID,
		PasswordHash: PasswordHash,
		PhoneNumber:  req.PhoneNumber,
	}
	if err := u.repo.Create(user); err != nil {
		return nil, err
	}
	return toUserResponse(user), nil
}

func (u *usecase) Login(req dto.LoginRequest) (*dto.AuthTokenResponse, error) {
	user, err := u.authenticate(req)
	if err != nil {
		return nil, err
	}

	return u.buildAuthTokens(user)
}

func (u *usecase) authenticate(req dto.LoginRequest) (*models.User, error) {
	user, err := u.repo.FindByIdentifier(req.Identifier)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, responses.ErrorUserCode["INVALID_CREDENTIAL"]
	}
	if err != nil {
		return nil, err
	}
	if !u.passwordService.Check(req.Password, user.PasswordHash) {
		return nil, responses.ErrorUserCode["INVALID_CREDENTIAL"]
	}

	return user, nil
}

func (u *usecase) buildAuthTokens(user *models.User) (*dto.AuthTokenResponse, error) {
	roleName := getUserRoleName(user)
	accessToken, err := u.tokenManager.GenerateAccessToken(uint(user.UserID), user.Email, roleName, user.RoleID, u.accessTokenExpiresIn)
	if err != nil {
		return nil, err
	}
	refreshToken, err := u.tokenManager.GenerateRefreshToken(uint(user.UserID), user.Email, roleName, user.RoleID, u.refreshTokenExpiresIn)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	session := &models.AuthSession{
		UserID:           user.UserID,
		RefreshTokenHash: utils.SHA256Hex(refreshToken),
		ExpiresAt:        now.Add(u.refreshTokenExpiresIn),
		LastUsedAt:       &now,
	}
	if err := u.repo.CreateSession(session); err != nil {
		return nil, err
	}

	return &dto.AuthTokenResponse{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		TokenType:        "Bearer",
		Role:             roleName,
		UserID:           uint(user.UserID),
		ExpiresIn:        u.accessTokenExpiresInSec,
		RefreshExpiresIn: u.refreshTokenExpiresInSec,
	}, nil
}

func (u *usecase) RefreshToken(refreshToken string) (*dto.AuthTokenResponse, error) {
	claims, err := u.tokenManager.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, responses.ErrorUserCode["INVALID_REFRESH_TOKEN"]
	}
	session, err := u.repo.FindActiveSessionByRefreshTokenHash(utils.SHA256Hex(refreshToken))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, responses.ErrorUserCode["INVALID_REFRESH_TOKEN"]
	}
	if err != nil {
		return nil, err
	}
	if session.ExpiresAt.Before(time.Now()) || session.RevokedAt != nil {
		return nil, responses.ErrorUserCode["INVALID_REFRESH_TOKEN"]
	}
	if err := u.repo.RevokeSessionByRefreshTokenHash(utils.SHA256Hex(refreshToken), time.Now()); err != nil {
		return nil, err
	}

	user, err := u.repo.FindByID(claims.UserID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, responses.ErrorUserCode["USER_NOT_FOUND"]
	}
	if err != nil {
		return nil, err
	}

	return u.buildAuthTokens(user)
}

func (u *usecase) Logout(refreshToken string) error {
	hash := utils.SHA256Hex(refreshToken)
	session, err := u.repo.FindActiveSessionByRefreshTokenHash(hash)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return responses.ErrorUserCode["INVALID_REFRESH_TOKEN"]
	}
	if err != nil {
		return err
	}
	if session.RevokedAt != nil {
		return responses.ErrorUserCode["INVALID_REFRESH_TOKEN"]
	}
	return u.repo.RevokeSessionByRefreshTokenHash(hash, time.Now())
}

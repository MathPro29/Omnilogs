package usecase

import (
	"errors"
	"time"

	"omnilogs-api/dto"
	authrepo "omnilogs-api/internal/auth/repository"
	"omnilogs-api/models"
	"omnilogs-api/responses"
	"omnilogs-api/utils"

	"gorm.io/gorm"
)

const defaultUserRoleID uint = 4

type Usecase interface {
	Register(req dto.RegisterRequest) (*dto.UserResponse, error)
	Login(req dto.LoginRequest) (*dto.AuthTokenResponse, error)
	ListAllUsers() ([]dto.UserResponse, error)
	RefreshToken(refreshToken string) (*dto.AuthTokenResponse, error)
	GetMe(userID uint) (*dto.UserResponse, error)
	GiveAdminAccess(userID uint, adminID uint, roleID uint) error
}

type usecase struct {
	repo                     authrepo.Repository
	jwtSecret                string
	accessTokenExpiresIn     time.Duration
	refreshTokenExpiresIn    time.Duration
	accessTokenExpiresInSec  int64
	refreshTokenExpiresInSec int64
}

func NewUsecase(repo authrepo.Repository, jwtSecret string, accessTokenExpiresInSec int, refreshTokenExpiresInSec int) Usecase {
	return &usecase{
		repo:                     repo,
		jwtSecret:                jwtSecret,
		accessTokenExpiresIn:     time.Duration(accessTokenExpiresInSec) * time.Second,
		refreshTokenExpiresIn:    time.Duration(refreshTokenExpiresInSec) * time.Second,
		accessTokenExpiresInSec:  int64(accessTokenExpiresInSec),
		refreshTokenExpiresInSec: int64(refreshTokenExpiresInSec),
	}
}

func (u *usecase) Register(req dto.RegisterRequest) (*dto.UserResponse, error) {
	if _, err := u.repo.FindByEmail(req.Email); err == nil {
		return nil, responses.ErrorUserCode["EMAIL_ALREADY_EXISTS"]
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	PasswordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
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

	roleName := ""
	if user.Role != nil {
		roleName = user.Role.RoleName
	}
	return u.buildAuthTokens(user, roleName)
}

func (u *usecase) authenticate(req dto.LoginRequest) (*models.User, error) {
	user, err := u.repo.FindByEmail(req.Email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, responses.ErrorUserCode["INVALID_CREDENTIAL"]
	}
	if err != nil {
		return nil, err
	}
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, responses.ErrorUserCode["INVALID_CREDENTIAL"]
	}

	return user, nil
}

func (u *usecase) buildAuthTokens(user *models.User, roleName string) (*dto.AuthTokenResponse, error) {
	accessToken, err := utils.GenerateToken(uint(user.UserID), user.Email, roleName, user.RoleID, utils.TokenTypeAccess, u.jwtSecret, u.accessTokenExpiresIn)
	if err != nil {
		return nil, err
	}
	refreshToken, err := utils.GenerateToken(uint(user.UserID), user.Email, roleName, user.RoleID, utils.TokenTypeRefresh, u.jwtSecret, u.refreshTokenExpiresIn)
	if err != nil {
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
	claims, err := utils.ParseToken(refreshToken, u.jwtSecret)
	if err != nil {
		return nil, responses.ErrorUserCode["INVALID_REFRESH_TOKEN"]
	}
	if err := utils.RequireTokenType(claims, utils.TokenTypeRefresh); err != nil {
		return nil, responses.ErrorUserCode["INVALID_REFRESH_TOKEN"]
	}

	accessToken, err := utils.GenerateToken(claims.UserID, claims.Email, claims.Role, claims.RoleID, utils.TokenTypeAccess, u.jwtSecret, u.accessTokenExpiresIn)
	if err != nil {
		return nil, err
	}

	return &dto.AuthTokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		UserID:      claims.UserID,
		Role:        claims.Role,
		ExpiresIn:   u.accessTokenExpiresInSec,
	}, nil
}

func (u *usecase) GetMe(userID uint) (*dto.UserResponse, error) {
	user, err := u.repo.FindByID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, responses.ErrorUserCode["USER_NOT_FOUND"]
	}
	if err != nil {
		return nil, err
	}
	return toUserResponse(user), nil
}

func toUserResponse(user *models.User) *dto.UserResponse {
	roleName := ""
	if user.Role != nil {
		roleName = user.Role.RoleName
	}

	return &dto.UserResponse{
		ID:          uint(user.UserID),
		UserID:      user.UserID,
		Username:    user.Username,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		IsActive:    user.IsActive,
		Role:        roleName,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

func (u *usecase) GiveAdminAccess(userID uint, adminID uint, roleID uint) error {
	// 1. First check if the adminID user exists
	user, err := u.repo.FindByID(adminID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return responses.ErrorUserCode["USER_NOT_FOUND"]
	}
	if err != nil {
		return err
	}

	// 2. Update the role for that existing user
	user.RoleID = roleID
	if err := u.repo.UpdateRoleID(uint(user.UserID), roleID); err != nil {
		return err // Return the database error if the update fails
	}

	return nil // Success
}

func (u *usecase) ListAllUsers() ([]dto.UserResponse, error) {
	users, err := u.repo.ListAllUsers()
	if err != nil {
		return nil, err
	}
	var userResponses []dto.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, *toUserResponse(&user))
	}
	return userResponses, nil
}

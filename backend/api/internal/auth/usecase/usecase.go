package usecase

import (
	"errors"
	"time"

	"omnilogs-api/dto"
	authrepo "omnilogs-api/internal/auth/repository"
	authservice "omnilogs-api/internal/auth/service"
	"omnilogs-api/models"
)

const defaultUserRoleID uint = 4

var (
	ErrForbiddenRoleAssignment = errors.New("forbidden role assignment")
	ErrInvalidRoleAssignment   = errors.New("invalid role assignment")
)

type Usecase interface {
	Register(req dto.RegisterRequest) (*dto.UserResponse, error)
	Login(req dto.LoginRequest) (*dto.AuthTokenResponse, error)
	ListAllUsers() ([]dto.UserResponse, error)
	RefreshToken(refreshToken string) (*dto.AuthTokenResponse, error)
	Logout(refreshToken string) error
	ForgotPassword(req dto.ForgotPasswordRequest) (*dto.ForgotPasswordResponse, error)
	ResetPassword(req dto.ResetPasswordRequest) error
	GetMe(userID uint) (*dto.UserResponse, error)
	GiveAdminAccess(userID uint, adminID uint, roleID uint) (*dto.GiveAdminAccessResponse, error)
	CheckUserExistsByUsername(username string) error
}

type usecase struct {
	repo                     authrepo.Repository
	tokenManager             authservice.TokenManager
	passwordService          authservice.PasswordService
	opaqueTokenService       authservice.OpaqueTokenService
	accessTokenExpiresIn     time.Duration
	refreshTokenExpiresIn    time.Duration
	accessTokenExpiresInSec  int64
	refreshTokenExpiresInSec int64
}

func NewUsecase(
	repo authrepo.Repository,
	tokenManager authservice.TokenManager,
	passwordService authservice.PasswordService,
	opaqueTokenService authservice.OpaqueTokenService,
	accessTokenExpiresIn time.Duration,
	refreshTokenExpiresIn time.Duration,
) Usecase {
	return &usecase{
		repo:                     repo,
		tokenManager:             tokenManager,
		passwordService:          passwordService,
		opaqueTokenService:       opaqueTokenService,
		accessTokenExpiresIn:     accessTokenExpiresIn,
		refreshTokenExpiresIn:    refreshTokenExpiresIn,
		accessTokenExpiresInSec:  int64(accessTokenExpiresIn / time.Second),
		refreshTokenExpiresInSec: int64(refreshTokenExpiresIn / time.Second),
	}
}

func getUserRoleName(user *models.User) string {
	if user != nil && user.Role != nil {
		return user.Role.RoleName
	}
	return ""
}

func toUserResponse(user *models.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:          uint(user.UserID),
		UserID:      user.UserID,
		Username:    user.Username,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		IsActive:    user.IsActive,
		Role:        getUserRoleName(user),
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

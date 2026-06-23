package usecase

import (
	"errors"
	"time"

	"omnilogs-api/dto"
	authrepo "omnilogs-api/internal/auth/repository"
	authservice "omnilogs-api/internal/auth/service"
	"omnilogs-api/models"
	"omnilogs-api/responses"
	"omnilogs-api/utils"

	"gorm.io/gorm"
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

func (u *usecase) Register(req dto.RegisterRequest) (*dto.UserResponse, error) {
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

func (u *usecase) buildAuthTokens(user *models.User, roleName string) (*dto.AuthTokenResponse, error) {
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

	roleName := ""
	if user.Role != nil {
		roleName = user.Role.RoleName
	}
	return u.buildAuthTokens(user, roleName)
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

func (u *usecase) GiveAdminAccess(userID uint, adminID uint, roleID uint) (*dto.GiveAdminAccessResponse, error) {
	actor, err := u.repo.FindByID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, responses.ErrorUserCode["USER_NOT_FOUND"]
	}
	if err != nil {
		return nil, err
	}

	target, err := u.repo.FindByID(adminID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, responses.ErrorUserCode["USER_NOT_FOUND"]
	}
	if err != nil {
		return nil, err
	}

	actorRole := ""
	if actor.Role != nil {
		actorRole = actor.Role.RoleName
	}
	switch actorRole {
	case "god":
		// GOD may assign any platform role.
	case "owner":
		if roleID == 1 {
			return nil, ErrForbiddenRoleAssignment
		}
		allowed, err := u.repo.CanOwnerManageUser(userID, adminID)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, ErrForbiddenRoleAssignment
		}
	case "superadmin":
		return nil, ErrForbiddenRoleAssignment
	default:
		return nil, ErrForbiddenRoleAssignment
	}

	previousRoleID := target.RoleID
	previousRoleCode := ""
	if target.Role != nil {
		previousRoleCode = target.Role.RoleName
	}

	newRole, err := u.repo.FindPlatformRoleByID(roleID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidRoleAssignment
	}
	if err != nil {
		return nil, err
	}

	target.RoleID = roleID
	if err := u.repo.UpdateRoleID(uint(target.UserID), roleID); err != nil {
		return nil, err
	}

	return &dto.GiveAdminAccessResponse{
		UserID:           uint(target.UserID),
		StorageTable:     "platform_memberships",
		PreviousRoleID:   previousRoleID,
		PreviousRoleCode: previousRoleCode,
		NewRoleID:        uint(newRole.PlatformRoleID),
		NewRoleCode:      newRole.RoleCode,
		Message:          "platform role updated successfully",
	}, nil
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

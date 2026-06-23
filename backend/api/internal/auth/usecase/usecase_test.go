package usecase

import (
	"errors"
	"testing"
	"time"

	"omnilogs-api/dto"
	authservice "omnilogs-api/internal/auth/service"
	"omnilogs-api/models"

	"gorm.io/gorm"
)

type authRepoStub struct {
	users             map[uint]*models.User
	roles             map[uint]*models.PlatformRole
	ownerCanManage    bool
	updateRoleUserID  uint
	updateRoleRoleID  uint
	updateRoleInvoked bool
	sessions          map[string]*models.AuthSession
	resetTokens       map[string]*models.PasswordResetToken
}

func newTestUsecase(repo *authRepoStub) Usecase {
	return NewUsecase(
		repo,
		authservice.NewJWTTokenManager("secret"),
		authservice.NewBCryptPasswordService(),
		authservice.NewRandomOpaqueTokenService(),
		900*time.Second,
		604800*time.Second,
	)
}

func (s *authRepoStub) Create(user *models.User) error { return nil }
func (s *authRepoStub) FindByEmail(email string) (*models.User, error) {
	for _, user := range s.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (s *authRepoStub) FindByIdentifier(identifier string) (*models.User, error) {
	for _, user := range s.users {
		if user.Email == identifier {
			return user, nil
		}
		if user.Username != nil && *user.Username == identifier {
			return user, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (s *authRepoStub) FindByID(userID uint) (*models.User, error) {
	user, ok := s.users[userID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return user, nil
}
func (s *authRepoStub) ListAllUsers() ([]models.User, error) { return nil, nil }
func (s *authRepoStub) UpdateRoleID(userID uint, roleID uint) error {
	s.updateRoleUserID = userID
	s.updateRoleRoleID = roleID
	s.updateRoleInvoked = true
	return nil
}
func (s *authRepoStub) CanOwnerManageUser(ownerUserID uint, targetUserID uint) (bool, error) {
	return s.ownerCanManage, nil
}
func (s *authRepoStub) FindPlatformRoleByID(roleID uint) (*models.PlatformRole, error) {
	role, ok := s.roles[roleID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return role, nil
}
func (s *authRepoStub) CreateSession(session *models.AuthSession) error {
	if s.sessions == nil {
		s.sessions = map[string]*models.AuthSession{}
	}
	s.sessions[session.RefreshTokenHash] = session
	return nil
}
func (s *authRepoStub) FindActiveSessionByRefreshTokenHash(hash string) (*models.AuthSession, error) {
	session, ok := s.sessions[hash]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return session, nil
}
func (s *authRepoStub) RevokeSessionByRefreshTokenHash(hash string, revokedAt time.Time) error {
	session, ok := s.sessions[hash]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	session.RevokedAt = &revokedAt
	return nil
}
func (s *authRepoStub) TouchSession(sessionID int, lastUsedAt time.Time) error { return nil }
func (s *authRepoStub) RevokeAllUserSessions(userID uint, revokedAt time.Time) error {
	for _, session := range s.sessions {
		if session.UserID == int(userID) {
			session.RevokedAt = &revokedAt
		}
	}
	return nil
}
func (s *authRepoStub) CreatePasswordResetToken(token *models.PasswordResetToken) error {
	if s.resetTokens == nil {
		s.resetTokens = map[string]*models.PasswordResetToken{}
	}
	s.resetTokens[token.TokenHash] = token
	return nil
}
func (s *authRepoStub) FindActivePasswordResetToken(hash string) (*models.PasswordResetToken, error) {
	token, ok := s.resetTokens[hash]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return token, nil
}
func (s *authRepoStub) MarkPasswordResetTokenUsed(id int, usedAt time.Time) error {
	for _, token := range s.resetTokens {
		if token.PasswordResetTokenID == id {
			token.UsedAt = &usedAt
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}
func (s *authRepoStub) UpdatePassword(userID uint, passwordHash string) error {
	user, ok := s.users[userID]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	user.PasswordHash = passwordHash
	return nil
}

func TestGiveAdminAccessRules(t *testing.T) {
	makeUser := func(id uint, role string) *models.User {
		return &models.User{
			UserID: int(id),
			Role:   &models.UserRole{RoleName: role},
		}
	}

	t.Run("god can assign any role", func(t *testing.T) {
		repo := &authRepoStub{
			users: map[uint]*models.User{
				1: makeUser(1, "god"),
				2: makeUser(2, "user"),
			},
			roles: map[uint]*models.PlatformRole{
				1: {PlatformRoleID: 1, RoleCode: "god"},
			},
		}
		usecase := newTestUsecase(repo)

		result, err := usecase.GiveAdminAccess(1, 2, 1)
		if err != nil {
			t.Fatalf("expected god assignment to pass, got %v", err)
		}
		if result.StorageTable != "platform_memberships" {
			t.Fatalf("expected storage table to be platform_memberships, got %s", result.StorageTable)
		}
		if !repo.updateRoleInvoked || repo.updateRoleRoleID != 1 {
			t.Fatal("expected role update to be invoked for god assignment")
		}
	})

	t.Run("owner cannot assign god", func(t *testing.T) {
		repo := &authRepoStub{
			users: map[uint]*models.User{
				1: makeUser(1, "owner"),
				2: makeUser(2, "user"),
			},
			roles: map[uint]*models.PlatformRole{
				1: {PlatformRoleID: 1, RoleCode: "god"},
			},
			ownerCanManage: true,
		}
		usecase := newTestUsecase(repo)

		_, err := usecase.GiveAdminAccess(1, 2, 1)
		if !errors.Is(err, ErrForbiddenRoleAssignment) {
			t.Fatalf("expected forbidden assignment, got %v", err)
		}
	})

	t.Run("owner can assign only within managed products", func(t *testing.T) {
		repo := &authRepoStub{
			users: map[uint]*models.User{
				1: makeUser(1, "owner"),
				2: makeUser(2, "user"),
			},
			roles: map[uint]*models.PlatformRole{
				4: {PlatformRoleID: 4, RoleCode: "user"},
			},
			ownerCanManage: false,
		}
		usecase := newTestUsecase(repo)

		_, err := usecase.GiveAdminAccess(1, 2, 4)
		if !errors.Is(err, ErrForbiddenRoleAssignment) {
			t.Fatalf("expected forbidden assignment, got %v", err)
		}
	})

	t.Run("superadmin cannot assign any role", func(t *testing.T) {
		repo := &authRepoStub{
			users: map[uint]*models.User{
				1: makeUser(1, "superadmin"),
				2: makeUser(2, "user"),
			},
			roles: map[uint]*models.PlatformRole{
				4: {PlatformRoleID: 4, RoleCode: "user"},
			},
		}
		usecase := newTestUsecase(repo)

		_, err := usecase.GiveAdminAccess(1, 2, 4)
		if !errors.Is(err, ErrForbiddenRoleAssignment) {
			t.Fatalf("expected forbidden assignment, got %v", err)
		}
	})
}

func TestForgotPasswordCreatesResetToken(t *testing.T) {
	repo := &authRepoStub{
		users: map[uint]*models.User{
			1: {UserID: 1, Email: "user@example.com"},
		},
	}
	usecase := newTestUsecase(repo)

	result, err := usecase.ForgotPassword(dto.ForgotPasswordRequest{Email: "user@example.com"})
	if err != nil {
		t.Fatalf("expected forgot password to pass, got %v", err)
	}
	if result.ResetToken == "" {
		t.Fatal("expected reset token to be returned")
	}
}

var _ interface {
	Create(*models.User) error
	FindByEmail(string) (*models.User, error)
	FindByID(uint) (*models.User, error)
	ListAllUsers() ([]models.User, error)
	UpdateRoleID(uint, uint) error
	CanOwnerManageUser(uint, uint) (bool, error)
	FindPlatformRoleByID(uint) (*models.PlatformRole, error)
	CreateSession(*models.AuthSession) error
	FindActiveSessionByRefreshTokenHash(string) (*models.AuthSession, error)
	RevokeSessionByRefreshTokenHash(string, time.Time) error
	TouchSession(int, time.Time) error
	RevokeAllUserSessions(uint, time.Time) error
	CreatePasswordResetToken(*models.PasswordResetToken) error
	FindActivePasswordResetToken(string) (*models.PasswordResetToken, error)
	MarkPasswordResetTokenUsed(int, time.Time) error
	UpdatePassword(uint, string) error
} = (*authRepoStub)(nil)

package service

import (
	"time"

	"omnilogs-api/utils"
)

type TokenManager interface {
	GenerateAccessToken(userID uint, email string, role string, roleID uint, expiresIn time.Duration) (string, error)
	GenerateRefreshToken(userID uint, email string, role string, roleID uint, expiresIn time.Duration) (string, error)
	ParseRefreshToken(token string) (*utils.Claims, error)
}

type JWTTokenManager struct {
	secret string
}

func NewJWTTokenManager(secret string) *JWTTokenManager {
	return &JWTTokenManager{secret: secret}
}

func (m *JWTTokenManager) GenerateAccessToken(userID uint, email string, role string, roleID uint, expiresIn time.Duration) (string, error) {
	return utils.GenerateToken(userID, email, role, roleID, utils.TokenTypeAccess, m.secret, expiresIn)
}

func (m *JWTTokenManager) GenerateRefreshToken(userID uint, email string, role string, roleID uint, expiresIn time.Duration) (string, error) {
	return utils.GenerateToken(userID, email, role, roleID, utils.TokenTypeRefresh, m.secret, expiresIn)
}

func (m *JWTTokenManager) ParseRefreshToken(token string) (*utils.Claims, error) {
	claims, err := utils.ParseToken(token, m.secret)
	if err != nil {
		return nil, err
	}
	if err := utils.RequireTokenType(claims, utils.TokenTypeRefresh); err != nil {
		return nil, err
	}
	return claims, nil
}

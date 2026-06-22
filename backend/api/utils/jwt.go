package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

var ErrInvalidTokenType = errors.New("invalid token type")

type Claims struct {
	UserID    uint   `json:"userId"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	RoleID    uint   `json:"roleId"`
	TokenType string `json:"tokenType"`
	jwt.RegisteredClaims
}

func GenerateToken(userID uint, email string, role string, roleID uint, tokenType string, secret string, expiresIn time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		RoleID:    roleID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseToken(tokenText string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenText, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func RequireTokenType(claims *Claims, tokenType string) error {
	if claims.TokenType != tokenType {
		return ErrInvalidTokenType
	}
	return nil
}

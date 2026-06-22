package dto

import "time"

type RegisterRequest struct {
	FirstName   string  `json:"firstName" binding:"required"`
	LastName    string  `json:"lastName" binding:"required"`
	Email       string  `json:"email" binding:"required,email"`
	PhoneNumber *string `json:"phoneNumber" binding:"omitempty,max=10"`
	Provider    string  `json:"-" binding:"omitempty"`
	ProviderID  uint    `json:"-" binding:"omitempty"`
	Password    string  `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type UserResponse struct {
	ID        uint      `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type AuthTokenResponse struct {
	AccessToken      string `json:"accessToken"`
	RefreshToken     string `json:"refreshToken,omitempty"`
	TokenType        string `json:"tokenType"`
	Role             string `json:"role"`
	UserID           uint   `json:"userId"`
	ExpiresIn        int64  `json:"expiresIn"`
	RefreshExpiresIn int64  `json:"refreshExpiresIn,omitempty"`
}

type GiveAdminAccessRequest struct {
	ID     uint `json:"id" binding:"required"`
	RoleID uint `json:"roleId " binding:"required"`
}

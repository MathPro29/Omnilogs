package dto

import "time"

type RegisterRequest struct {
	Username        *string `json:"username,omitempty"`
	FirstName       *string `json:"first_name" binding:"required"`
	LastName        *string `json:"last_name" binding:"required"`
	Email           string  `json:"email" binding:"required,email"`
	PhoneNumber     *string `json:"phone_number,omitempty" binding:"omitempty,max=20"`
	Password        string  `json:"password" binding:"required,min=8"`
	ConfirmPassword string  `json:"confirm_password" binding:"required,eqfield=Password"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token           string `json:"token" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}
type GiveAdminAccessRequest struct {
	ID     uint `json:"id" binding:"required,gt=0"`
	RoleID uint `json:"role_id" binding:"required,oneof=1 2 3 4"`
}

type GiveAdminAccessResponse struct {
	UserID           uint   `json:"user_id"`
	StorageTable     string `json:"storage_table"`
	PreviousRoleID   uint   `json:"previous_role_id"`
	PreviousRoleCode string `json:"previous_role_code"`
	NewRoleID        uint   `json:"new_role_id"`
	NewRoleCode      string `json:"new_role_code"`
	Message          string `json:"message"`
}

type UserDetailResponse struct {
	UserID           int       `json:"user_id" binding:"required,gt=0"`
	Username         string    `json:"username" binding:"required"`
	FirstName        string    `json:"first_name" binding:"required"`
	LastName         string    `json:"last_name" binding:"required"`
	Email            string    `json:"email" binding:"required,email"`
	PhoneNumber      *string   `json:"phone_number,omitempty"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	PlatformRoleID   int       `json:"platform_role_id" binding:"required,gt=0"`
	PlatformRoleName string    `json:"platform_role_name" binding:"required"`
	ProductRoleID    int       `json:"product_role_id" binding:"required,gt=0"`
	ProductRoleName  string    `json:"product_role_name" binding:"required"`
	PermissionLevel  int       `json:"permission_level" binding:"required,gt=0"`
}

type UserListFilterRequest struct {
	PlatformRoleID  *int   `form:"platform_role_id" binding:"omitempty"`
	ProductRoleID   *int   `form:"product_role_id" binding:"omitempty"`
	PermissionLevel *int   `form:"permission_level" binding:"omitempty"`
	IsActive        *bool  `form:"is_active" binding:"omitempty"`
	Page            int    `form:"page" binding:"omitempty"`
	Limit           int    `form:"limit" binding:"omitempty"`
	Sort            string `form:"sort" binding:"omitempty"`
	Order           string `form:"order" binding:"omitempty"`
	Search          string `form:"search" binding:"omitempty"`
}

type UpdateUserRequest struct {
	Username    *string `json:"username,omitempty"`
	FirstName   *string `json:"first_name,omitempty"`
	LastName    *string `json:"last_name,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

type UserResponse struct {
	ID          uint       `json:"id"`
	UserID      int        `json:"user_id,omitempty"`
	Username    *string    `json:"username,omitempty"`
	FirstName   *string    `json:"first_name,omitempty"`
	LastName    *string    `json:"last_name,omitempty"`
	Email       string     `json:"email"`
	PhoneNumber *string    `json:"phone_number,omitempty"`
	IsActive    bool       `json:"is_active"`
	Role        string     `json:"role,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type AuthTokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	TokenType        string `json:"token_type"`
	Role             string `json:"role"`
	UserID           uint   `json:"user_id"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in,omitempty"`
}

type ForgotPasswordResponse struct {
	Message    string `json:"message"`
	ResetToken string `json:"reset_token,omitempty"`
}

type EditUserRoleRequest struct {
	ID              uint   `json:"id" binding:"required,gt=0"`
	FirstName       string `json:"first_name,omitempty"`
	LastName        string `json:"last_name,omitempty"`
	Email           string `json:"email,omitempty,email"`
	PlatformRoleID  uint   `json:"platform_role_id,omitempty,gt=0"`
	ProductRoleID   uint   `json:"product_role_id,omitempty,gt=0"`
	EnvironmentID   uint   `json:"environment_id,omitempty,gt=0"`
	PermissionLevel uint   `json:"permission_level,omitempty,gt=0"`
	IsActive        *bool  `json:"is_active,omitempty"`
}

type CreateUserRequest struct {
	Username    *string `json:"username,omitempty"`
	FullName    string  `json:"fullName" binding:"required"`
	Email       string  `json:"email" binding:"required,email"`
	PhoneNumber *string `json:"phone,omitempty"`
	Password    string  `json:"password" binding:"required,min=8"`
	Role        string  `json:"role" binding:"required"`
}

type UpdateUserAdminRequest struct {
	FullName    string  `json:"fullName"`
	Email       string  `json:"email" binding:"email"`
	PhoneNumber *string `json:"phone,omitempty"`
	Role        string  `json:"role"`
	Status      string  `json:"status"` // 'active' | 'inactive'
}

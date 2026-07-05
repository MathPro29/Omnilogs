package handler

import (
	"errors"
	"fmt"
	"net/http"
	_ "strconv"

	"omnilogs-api/dto"
	authusecase "omnilogs-api/internal/auth/usecase"
	"omnilogs-api/middleware"
	"omnilogs-api/responses"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	usecase authusecase.Usecase
}

func NewHandler(usecase authusecase.Usecase) *Handler {
	return &Handler{usecase: usecase}
}

func (h *Handler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, "INVALID_REQUEST", responses.ErrorCode["INVALID_REQUEST"], err)
		return
	}

	// check if username is provided
	if req.Username == nil || *req.Username == "" {
		responses.Error(c, "INVALID_REQUEST", "username is required", nil)
		return
	}

	// check if username already exists
	if err := h.usecase.CheckUserExistsByUsername(*req.Username); err != nil {
		responses.Error(c, "USERNAME_ALREADY_EXISTS", responses.ErrorCode["USERNAME_ALREADY_EXISTS"], nil)
		return
	}

	// check phone number if > 10 return invalid phone number
	if req.PhoneNumber != nil && len(*req.PhoneNumber) > 10 {
		responses.Error(c, "INVALID_PHONE_NUMBER", responses.ErrorCode["INVALID_PHONE_NUMBER"], nil)
		return
	}

	// check phone number must contain only number
	if req.PhoneNumber != nil {
		for _, r := range *req.PhoneNumber {
			if r < '0' || r > '9' {
				responses.BadRequest(c, "invalid phone number")
				return
			}
		}
	}

	// phone number at least 8 character
	if req.PhoneNumber != nil && len(*req.PhoneNumber) < 8 {
		responses.Error(c, "INVALID_PHONE_NUMBER", responses.ErrorCode["INVALID_PHONE_NUMBER"], nil)
		return
	}

	user, err := h.usecase.Register(req)
	if errors.Is(err, responses.ErrorUserCode["EMAIL_ALREADY_EXISTS"]) {
		responses.Error(c, "EMAIL_ALREADY_EXISTS", responses.ErrorCode["EMAIL_ALREADY_EXISTS"], nil)
		return
	}
	if err != nil {
		responses.Error(c, "INTERNAL_ERROR", responses.ErrorCode["INTERNAL_ERROR"], err)
		return
	}

	c.Set(middleware.ContextUserID, user.ID)
	c.Set("audit_reason", fmt.Sprintf("Registered new user account for email '%s'", user.Email))
	c.Set("audit_payload", gin.H{
		"user_id":     user.UserID,
		"email":       user.Email,
		"username":    user.Username,
		"is_active":   user.IsActive,
		"activity":    "REGISTER",
	})
	responses.Success(c, http.StatusCreated, "USER_REGISTERED", user)
}

func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationError(c, err)
		return
	}

	tokens, err := h.usecase.Login(req)
	if errors.Is(err, responses.ErrorUserCode["INVALID_CREDENTIAL"]) {
		responses.Unauthorized(c, "invalid username/email or password")
		return
	}
	if err != nil {
		responses.InternalError(c)
		return
	}

	c.Set(middleware.ContextUserID, tokens.UserID)
	c.Set("audit_reason", fmt.Sprintf("User login succeeded for identifier '%s'", req.Identifier))
	c.Set("audit_payload", gin.H{
		"user_id":    tokens.UserID,
		"role":       tokens.Role,
		"activity":   "LOGIN",
		"identifier": req.Identifier,
	})
	responses.Success(c, http.StatusOK, "USER_LOGIN", tokens)
}

func (h *Handler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationError(c, err)
		return
	}

	tokens, err := h.usecase.RefreshToken(req.RefreshToken)
	if errors.Is(err, responses.ErrorUserCode["INVALID_REFRESH_TOKEN"]) {
		responses.Unauthorized(c, "invalid refresh token")
		return
	}
	if err != nil {
		responses.InternalError(c)
		return
	}

	c.Set(middleware.ContextUserID, tokens.UserID)
	c.Set("audit_reason", fmt.Sprintf("User refreshed session for user ID %d", tokens.UserID))
	c.Set("audit_payload", gin.H{
		"user_id":  tokens.UserID,
		"role":     tokens.Role,
		"activity": "REFRESH_TOKEN",
	})
	responses.Success(c, http.StatusOK, "USER_REFRESH", tokens)
}

func (h *Handler) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationError(c, err)
		return
	}
	if err := h.usecase.Logout(req.RefreshToken); err != nil {
		if errors.Is(err, responses.ErrorUserCode["INVALID_REFRESH_TOKEN"]) {
			responses.Unauthorized(c, "invalid refresh token")
			return
		}
		responses.InternalError(c)
		return
	}
	c.Set("audit_reason", "User logout succeeded")
	c.Set("audit_payload", gin.H{
		"activity": "LOGOUT",
	})
	responses.Success(c, http.StatusOK, "USER_LOGOUT", gin.H{"message": "logged out successfully"})
}

func (h *Handler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationError(c, err)
		return
	}
	result, err := h.usecase.ForgotPassword(req)
	if err != nil {
		responses.InternalError(c)
		return
	}
	c.Set("audit_reason", fmt.Sprintf("Password reset requested for email '%s'", req.Email))
	c.Set("audit_payload", gin.H{
		"email":    req.Email,
		"activity": "FORGOT_PASSWORD",
	})
	responses.Success(c, http.StatusOK, "USER_UPDATE", result)
}

func (h *Handler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationError(c, err)
		return
	}
	if err := h.usecase.ResetPassword(req); err != nil {
		if errors.Is(err, responses.ErrorUserCode["INVALID_RESET_TOKEN"]) {
			responses.BadRequest(c, "invalid or expired reset token")
			return
		}
		responses.InternalError(c)
		return
	}
	c.Set("audit_reason", "Password reset completed successfully")
	c.Set("audit_payload", gin.H{
		"activity": "RESET_PASSWORD",
	})
	responses.Success(c, http.StatusOK, "USER_UPDATE", gin.H{"message": "password reset successfully"})
}

func (h *Handler) Me(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		responses.Unauthorized(c, "invalid authenticated user")
		return
	}

	user, err := h.usecase.GetMe(userID)
	if errors.Is(err, responses.ErrorUserCode["USER_NOT_FOUND"]) {
		responses.NotFound(c, "user not found")
		return
	}
	if err != nil {
		responses.InternalError(c)
		return
	}

	responses.Success(c, http.StatusOK, "USER_GET", user)
}

//// ADMIN ZONES

func (h *Handler) GiveAdminAccess(c *gin.Context) {
	// check user is logged in
	if _, ok := middleware.CurrentUserID(c); !ok {
		responses.Unauthorized(c, "invalid authenticated user")
		return
	}

	if !middleware.HasAdminPlatformRole(c) {
		responses.Forbidden(c, "forbidden")
		return
	}

	var req dto.GiveAdminAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationError(c, err)
		return
	}

	/*
		ROLE ID LIST
		ID 1 = GOD
		ID 2 = Owner
		ID 3 = Superadmin
		ID 4 = User
	*/

	// check role id is valid
	if req.RoleID != 1 && req.RoleID != 2 && req.RoleID != 3 && req.RoleID != 4 {
		responses.BadRequest(c, "invalid role id")
		return
	}
	currentUserID, ok := middleware.CurrentUserID(c)
	if !ok {
		responses.Unauthorized(c, "invalid authenticated user")
		return
	}

	// make sure admin don't change his own role
	if currentUserID == req.ID {
		responses.Error(c, "FORBIDDEN", "You cannot change your own role", nil)
		return
	}

	// Call usecase for update role
	result, err := h.usecase.GiveAdminAccess(currentUserID, req.ID, req.RoleID)
	if err != nil {
		if errors.Is(err, responses.ErrorUserCode["USER_NOT_FOUND"]) {
			responses.NotFound(c, "user not found")
			return
		}
		if errors.Is(err, authusecase.ErrForbiddenRoleAssignment) {
			responses.Forbidden(c, "forbidden role assignment")
			return
		}
		if errors.Is(err, authusecase.ErrInvalidRoleAssignment) {
			responses.BadRequest(c, "invalid role assignment")
			return
		}
		responses.Error(c, "INTERNAL_ERROR", "failed to update role", err)
		return
	}
	c.Set("audit_reason", fmt.Sprintf("Changed platform role for user %d to role %d", req.ID, req.RoleID))
	c.Set("audit_payload", result)

	responses.Success(c, http.StatusOK, "USER_UPDATE", result)
}

func (h *Handler) ListAllUsers(c *gin.Context) {
	if _, ok := middleware.CurrentUserID(c); !ok {
		responses.Unauthorized(c, "invalid authenticated user")
		return
	}

	if !middleware.HasAdminPlatformRole(c) {
		responses.Forbidden(c, "forbidden")
		return
	}

	// check if not get user list access
	if !middleware.HasUserListAccess(c) {
		responses.Forbidden(c, "forbidden")
		return
	}

	users, err := h.usecase.ListAllUsers()
	if err != nil {
		responses.InternalError(c)
		return
	}
	c.Set("audit_reason", fmt.Sprintf("Listed all users (%d records)", len(users)))
	c.Set("audit_payload", gin.H{
		"count":    len(users),
		"activity": "LIST_ALL_USERS",
	})

	responses.Success(c, http.StatusOK, "USER_LIST", users)
}

package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"omnilogs-api/dto"
	authusecase "omnilogs-api/internal/auth/usecase"
	"omnilogs-api/middleware"
	"omnilogs-api/responses"

	"github.com/gin-gonic/gin"
)

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

func (h *Handler) EditUserRole(c *gin.Context) {
	currentUserID, ok := middleware.CurrentUserID(c)
	if !ok {
		responses.Unauthorized(c, "invalid authenticated user")
		return
	}

	if !middleware.HasAdminPlatformRole(c) {
		responses.Forbidden(c, "forbidden")
		return
	}

	var req dto.EditUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationError(c, err)
		return
	}

	// Call usecase for update role
	err := h.usecase.EditUserRole(currentUserID, req)
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
	c.Set("audit_reason", fmt.Sprintf("Changed platform role/permissions for user %d", req.ID))
	c.Set("audit_payload", gin.H{
		"user_id":    req.ID,
		"activity":   "EDIT_USER_ROLE",
		"changed_by": currentUserID,
	})

	responses.Success(c, http.StatusOK, "USER_UPDATE", gin.H{"message": "role updated successfully"})
}

func (h *Handler) CreateUser(c *gin.Context) {
	if _, ok := middleware.CurrentUserID(c); !ok {
		responses.Unauthorized(c, "invalid authenticated user")
		return
	}

	if !middleware.HasAdminPlatformRole(c) {
		responses.Forbidden(c, "forbidden")
		return
	}

	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationError(c, err)
		return
	}

	user, err := h.usecase.CreateUser(req)
	if err != nil {
		if errors.Is(err, responses.ErrorUserCode["EMAIL_ALREADY_EXISTS"]) {
			responses.Error(c, "EMAIL_ALREADY_EXISTS", responses.ErrorCode["EMAIL_ALREADY_EXISTS"], nil)
			return
		}
		if errors.Is(err, responses.ErrorUserCode["USERNAME_ALREADY_EXISTS"]) {
			responses.Error(c, "USERNAME_ALREADY_EXISTS", responses.ErrorCode["USERNAME_ALREADY_EXISTS"], nil)
			return
		}
		responses.Error(c, "INTERNAL_ERROR", responses.ErrorCode["INTERNAL_ERROR"], err)
		return
	}

	c.Set("audit_reason", fmt.Sprintf("Created new user account for email '%s'", user.Email))
	c.Set("audit_payload", gin.H{
		"user_id":   user.UserID,
		"email":     user.Email,
		"username":  user.Username,
		"is_active": user.IsActive,
		"activity":  "CREATE_USER",
	})
	responses.Success(c, http.StatusCreated, "USER_CREATED", user)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	if _, ok := middleware.CurrentUserID(c); !ok {
		responses.Unauthorized(c, "invalid authenticated user")
		return
	}

	if !middleware.HasAdminPlatformRole(c) {
		responses.Forbidden(c, "forbidden")
		return
	}

	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		responses.BadRequest(c, "invalid user id")
		return
	}

	var req dto.UpdateUserAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationError(c, err)
		return
	}

	user, err := h.usecase.UpdateUser(uint(userID), req)
	if err != nil {
		if errors.Is(err, responses.ErrorUserCode["USER_NOT_FOUND"]) {
			responses.NotFound(c, "user not found")
			return
		}
		if errors.Is(err, responses.ErrorUserCode["EMAIL_ALREADY_EXISTS"]) {
			responses.Error(c, "EMAIL_ALREADY_EXISTS", responses.ErrorCode["EMAIL_ALREADY_EXISTS"], nil)
			return
		}
		responses.Error(c, "INTERNAL_ERROR", "failed to update user", err)
		return
	}

	c.Set("audit_reason", fmt.Sprintf("Updated user account %d", userID))
	c.Set("audit_payload", gin.H{
		"user_id":  userID,
		"activity": "UPDATE_USER",
	})
	responses.Success(c, http.StatusOK, "USER_UPDATE", user)
}

func (h *Handler) GetUserByID(c *gin.Context) {
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

	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		responses.BadRequest(c, "invalid user id")
		return
	}

	user, err := h.usecase.GetUserByID(uint(userID))
	if err != nil {
		if errors.Is(err, responses.ErrorUserCode["USER_NOT_FOUND"]) {
			responses.NotFound(c, "user not found")
			return
		}
		responses.InternalError(c)
		return
	}
	c.Set("audit_reason", fmt.Sprintf("Viewed user details of ID %d", userID))
	c.Set("audit_payload", gin.H{
		"user_id":  userID,
		"activity": "GET_USER_BY_ID",
	})

	responses.Success(c, http.StatusOK, "USER_GET", user)
}

func (h *Handler) DeleteUserByID(c *gin.Context) {
	if _, ok := middleware.CurrentUserID(c); !ok {
		responses.Unauthorized(c, "invalid authenticated user")
		return
	}

	if !middleware.HasAdminPlatformRole(c) {
		responses.Forbidden(c, "forbidden")
		return
	}

	// check if not delete user access
	if !middleware.HasDeleteUserAccess(c) {
		responses.Forbidden(c, "forbidden")
		return
	}

	// Get user id from path parameter
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		responses.BadRequest(c, "invalid user id")
		return
	}

	// Call usecase for delete user
	err = h.usecase.DeleteUserByID(uint(userID))
	if err != nil {
		if errors.Is(err, responses.ErrorUserCode["USER_NOT_FOUND"]) {
			responses.NotFound(c, "user not found")
			return
		}
		responses.Error(c, "INTERNAL_ERROR", "failed to delete user", err)
		return
	}
	c.Set("audit_reason", fmt.Sprintf("Deleted user %d", userID))
	c.Set("audit_payload", gin.H{
		"user_id":  userID,
		"activity": "DELETE_USER",
	})

	responses.Success(c, http.StatusOK, "USER_DELETE", gin.H{"message": "user deleted successfully"})
}

package middleware

import (
	"errors"
	"strings"

	"omnilogs-api/responses"
	"omnilogs-api/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	ContextUserID = "userId"
	ContextEmail  = "email"
	ContextRole   = "role"
	ContextRoleID = "roleId"
)

func IsSessionExpired(c *gin.Context, jwtSecret string) bool {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return false
	}

	_, err := utils.ParseToken(parts[1], jwtSecret)
	return errors.Is(err, jwt.ErrTokenExpired)
}

func UserAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			token := c.Query("token")
			if token != "" {
				authHeader = "Bearer " + token
			}
		}
		if authHeader == "" {
			responses.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			responses.Unauthorized(c, "invalid authorization header")
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(parts[1], jwtSecret)
		if err != nil {
			if IsSessionExpired(c, jwtSecret) {
				responses.Error(c, "SESSION_EXPIRED", "session expired", nil)
				c.Abort()
				return
			}
			responses.Unauthorized(c, "invalid token")
			c.Abort()
			return
		}
		if err := utils.RequireTokenType(claims, utils.TokenTypeAccess); err != nil {
			responses.Unauthorized(c, "access token is required")
			c.Abort()
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextEmail, claims.Email)
		c.Set(ContextRole, claims.Role)
		c.Set(ContextRoleID, claims.RoleID)
		c.Next()
	}
}

func RequireAdminPlatformRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := CurrentUserID(c); !ok {
			responses.Unauthorized(c, "unauthorized")
			c.Abort()
			return
		}
		if !HasAdminPlatformRole(c) {
			responses.Forbidden(c, "Forbidden or Permission Denied")
			c.Abort()
			return
		}
		c.Next()
	}
}

func CurrentUserID(c *gin.Context) (uint, bool) {
	value, ok := c.Get(ContextUserID)
	if !ok {
		return 0, false
	}
	userID, ok := value.(uint)
	return userID, ok
}

// HasAdminPlatformRole reports whether the authenticated platform role may
// enter the platform administration routes.
func HasAdminPlatformRole(c *gin.Context) bool {
	if _, ok := CurrentUserID(c); !ok {
		return false
	}
	role, ok := c.Get(ContextRole)
	if !ok {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(toString(role))) {
	case "god", "owner", "superadmin", "super_admin":
		return true
	default:
		return false
	}
}

// if don't have access to user list
func HasUserListAccess(c *gin.Context) bool {
	return true
}

// if don't have access to delete user
func HasDeleteUserAccess(c *gin.Context) bool {
	return true
}

// user can only access ticket which is created by him
func CanAccessTicket(c *gin.Context, ticketUserID uint) bool {
	if HasAdminPlatformRole(c) {
		return true
	}

	userID, ok := CurrentUserID(c)
	if !ok {
		return false
	}

	role, ok := c.Get(ContextRole)
	if !ok {
		return false
	}

	return toString(role) == "user" && userID == ticketUserID
}

func toString(value interface{}) string {
	str, _ := value.(string)
	return str
}

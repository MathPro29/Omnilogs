package middleware

import (
	"errors"
	"strings"

	"ticket-system-api/responses"
	"ticket-system-api/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	ContextUserID = "userId"
	ContextEmail  = "email"
	ContextRole   = "role"
	ContextRoleID = "roleId"
)

func IsUser(c *gin.Context) bool {
	role, ok := c.Get(ContextRole)
	if !ok {
		return false
	}
	return toString(role) == "user"
}

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

// check access for only admin and superadmin role
func AdminAuthMiddleware(adminJWTSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if IsUser(c) {
			responses.Forbidden(c, "Access Denied")
			c.Abort()
			return
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

		claims, err := utils.ParseToken(parts[1], adminJWTSecret)
		if err != nil {
			if IsSessionExpired(c, adminJWTSecret) {
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

func RequireUserOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := CurrentUserID(c); !ok {
			responses.Unauthorized(c, "unauthorized")
			c.Abort()
			return
		}
		if !IsUser(c) {
			responses.Forbidden(c, "forbidden user route only")
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireAdminOrSuperadmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := CurrentUserID(c); !ok {
			responses.Unauthorized(c, "unauthorized")
			c.Abort()
			return
		}
		if !IsAdminOrSuperadmin(c) {
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

// admin and superadmin can access all tickets

func IsAdminOrSuperadmin(c *gin.Context) bool {
	role, ok := c.Get(ContextRole)
	if !ok {
		return false
	}
	if toString(role) == "user" {
		return false
	}
	return toString(role) == "admin" || toString(role) == "superadmin"
}

// user can only access ticket which is created by him
func CanAccessTicket(c *gin.Context, ticketUserID uint) bool {
	if IsAdminOrSuperadmin(c) {
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

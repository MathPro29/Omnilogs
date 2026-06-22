package responses

import (
	"net/http"
	"omnilogs-api/utils"

	"github.com/gin-gonic/gin"
)

func Error(c *gin.Context, code string, message string, err error) {
	status := http.StatusBadRequest
	switch code {
	case "INTERNAL_ERROR", "INTERNAL_SERVER_ERROR":
		status = http.StatusInternalServerError
	case "FORBIDDEN", "ACCESS_DENIED":
		status = http.StatusForbidden
	case "SESSION_EXPIRED", "UNAUTHORIZED":
		status = http.StatusUnauthorized
	case "NOT_FOUND":
		status = http.StatusNotFound
	}
	var details interface{}
	if err != nil {
		details = err.Error()
	}
	utils.ErrorWithDetails(c, status, code, message, details)
}

func BadRequest(c *gin.Context, message string) {
	utils.BadRequest(c, message)
}

func ValidationError(c *gin.Context, err error) {
	utils.ValidationError(c, err)
}

func Unauthorized(c *gin.Context, message string) {
	utils.Unauthorized(c, message)
}

func Forbidden(c *gin.Context, message string) {
	utils.Error(c, http.StatusForbidden, "FORBIDDEN", message)
}

func NotFound(c *gin.Context, message string) {
	utils.NotFound(c, message)
}

func InternalError(c *gin.Context) {
	utils.InternalError(c)
}

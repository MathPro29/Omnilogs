package responses

import (
	"net/http"
	"omnilogs-api/utils"

	"github.com/gin-gonic/gin"
)

func Success(c *gin.Context, status int, code string, data interface{}) {
	message, exists := SuccessCode[code]
	if !exists {
		message = "success"
	}
	if data == nil {
		data = gin.H{}
	}
	c.JSON(status, utils.ResponseBody{
		Success: true,
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func SuccessWithMeta(c *gin.Context, status int, code string, data interface{}, meta utils.PaginationMeta) {
	message, exists := SuccessCode[code]
	if !exists {
		message = "success"
	}
	if data == nil {
		data = gin.H{}
	}
	c.JSON(status, utils.ResponseBody{
		Success: true,
		Code:    code,
		Message: message,
		Data:    data,
		Meta:    &meta,
	})
}

func Error(c *gin.Context, code string, message string, err error) {
	if message == "" {
		if msg, exists := ErrorCode[code]; exists {
			message = msg
		} else {
			message = code
		}
	}
	status := http.StatusBadRequest
	switch code {
	case "INTERNAL_ERROR", "INTERNAL_SERVER_ERROR":
		status = http.StatusInternalServerError
	case "TIMEOUT":
		status = http.StatusGatewayTimeout
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

package utils

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ResponseBody struct {
	Success bool            `json:"success"`
	Data    interface{}     `json:"data,omitempty"`
	Error   *ErrorBody      `json:"error,omitempty"`
	Meta    *PaginationMeta `json:"meta,omitempty"` // list[]
}

type ErrorBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Status  int         `json:"status"`
	Details interface{} `json:"details,omitempty"`
}

type PaginationMeta struct {
	Page      int   `json:"page"`
	PerPage   int   `json:"perPage"`
	Total     int64 `json:"total"`
	TotalPage int   `json:"totalPage"`
}

type FieldError struct {
	Field   string `json:"field"`
	Rule    string `json:"rule,omitempty"`
	Message string `json:"message"`
}

func Success(c *gin.Context, status int, data interface{}) {
	if data == nil {
		data = gin.H{}
	}
	c.JSON(status, ResponseBody{
		Success: true,
		Data:    data,
	})
}

func SuccessWithMeta(c *gin.Context, status int, data interface{}, meta PaginationMeta) {
	if data == nil {
		data = gin.H{}
	}
	c.JSON(status, ResponseBody{
		Success: true,
		Data:    data,
		Meta:    &meta,
	})
}

func Error(c *gin.Context, status int, code string, message string) {
	ErrorWithDetails(c, status, code, message, nil)
}

func ErrorWithDetails(c *gin.Context, status int, code string, message string, details interface{}) {
	c.JSON(status, ResponseBody{
		Success: false,
		Error: &ErrorBody{
			Code:    code,
			Message: message,
			Status:  status,
			Details: details,
		},
	})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

func ValidationError(c *gin.Context, err error) {
	ErrorWithDetails(c, http.StatusBadRequest, "VALIDATION_ERROR", "request validation failed", validationDetails(err))
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, "NOT_FOUND", message)
}

func InternalError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
}

func NewPaginationMeta(page int, perPage int, total int64) PaginationMeta {
	totalPage := 0
	if perPage > 0 {
		totalPage = int((total + int64(perPage) - 1) / int64(perPage))
	}

	return PaginationMeta{
		Page:      page,
		PerPage:   perPage,
		Total:     total,
		TotalPage: totalPage,
	}
}

func validationDetails(err error) []FieldError {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		details := make([]FieldError, 0, len(validationErrors))
		for _, fieldError := range validationErrors {
			field := toLowerCamel(fieldError.Field())
			details = append(details, FieldError{
				Field:   field,
				Rule:    fieldError.Tag(),
				Message: validationMessage(field, fieldError),
			})
		}
		return details
	}

	return []FieldError{
		{
			Field:   "body",
			Message: "invalid request body",
		},
	}
}

func validationMessage(field string, fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email"
	case "min":
		return field + " must be at least " + fieldError.Param() + " characters"
	default:
		return field + " is invalid"
	}
}

func toLowerCamel(value string) string {
	if value == "" {
		return value
	}
	if len(value) == 1 {
		return strings.ToLower(value)
	}
	return strings.ToLower(value[:1]) + value[1:]
}

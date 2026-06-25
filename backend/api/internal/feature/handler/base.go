package handler

import (
	"errors"
	"net/http"
	"strconv"

	"omnilogs-api/internal/feature/usecase"
	"omnilogs-api/middleware"
	"omnilogs-api/responses"
	"omnilogs-api/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct{ usecase usecase.Usecase }

func NewHandler(usecase usecase.Usecase) *Handler { return &Handler{usecase: usecase} }

func actor(c *gin.Context) (usecase.Actor, bool) {
	id, ok := middleware.CurrentUserID(c)
	return usecase.Actor{UserID: int(id), PlatformAdmin: middleware.HasAdminPlatformRole(c)}, ok
}

func idParam(c *gin.Context, name string) (int, bool) {
	value, err := strconv.Atoi(c.Param(name))
	if err != nil || value <= 0 {
		responses.BadRequest(c, "invalid "+name)
		return 0, false
	}
	return value, true
}

func bind(c *gin.Context, value any) bool {
	if err := c.ShouldBindJSON(value); err != nil {
		responses.ValidationError(c, err)
		return false
	}
	return true
}

func fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrNotFound):
		responses.NotFound(c, "feature resource not found")
	case errors.Is(err, usecase.ErrForbidden):
		responses.Forbidden(c, "feature permission denied")
	case errors.Is(err, usecase.ErrConflict):
		utils.Error(c, http.StatusConflict, "CONFLICT", "feature resource already exists")
	case errors.Is(err, usecase.ErrInvalid):
		responses.BadRequest(c, "invalid feature relationship or payload")
	default:
		responses.Error(c, "INTERNAL_ERROR", "internal server error", err)
	}
}

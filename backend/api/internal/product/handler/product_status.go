package handler

import (
	"errors"
	"net/http"
	"strconv"

	"omnilogs-api/internal/product/usecase"
	"omnilogs-api/middleware"
	"omnilogs-api/responses"
	"omnilogs-api/utils"

	"github.com/gin-gonic/gin"
)

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
	var scopedConflict *responses.ScopedConflictError
	if errors.As(err, &scopedConflict) {
		utils.ErrorWithDetails(c, http.StatusConflict, "CONFLICT", scopedConflict.Error(), gin.H{"field": scopedConflict.Field, "value": scopedConflict.Value, "scope": scopedConflict.Scope})
		return
	}
	switch {
	case errors.Is(err, responses.ErrNotFound):
		responses.NotFound(c, "product resource not found")
	case errors.Is(err, responses.ErrForbidden):
		responses.Forbidden(c, "product permission denied")
	case errors.Is(err, responses.ErrConflict):
		utils.Error(c, http.StatusConflict, "CONFLICT", "product resource already exists")
	case errors.Is(err, responses.ErrInvalid):
		responses.BadRequest(c, "invalid product relationship or payload")
	default:
		responses.Error(c, "INTERNAL_ERROR", "internal server error", err)
	}
}

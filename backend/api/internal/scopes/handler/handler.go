package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"omnilogs-api/dto"
	"omnilogs-api/internal/scopes/usecase"
	"omnilogs-api/middleware"
	"omnilogs-api/responses"
	"omnilogs-api/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	usecase usecase.Usecase
}

func NewHandler(usecase usecase.Usecase) *Handler {
	return &Handler{usecase: usecase}
}

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
	case errors.Is(err, responses.ErrorScopeCode["NOT_FOUND"]):
		responses.NotFound(c, "scope resource not found")
	case errors.Is(err, responses.ErrorScopeCode["FORBIDDEN"]):
		responses.Forbidden(c, "scope permission denied")
	case errors.Is(err, responses.ErrorScopeCode["CONFLICT"]):
		utils.Error(c, http.StatusConflict, "CONFLICT", "scope resource already exists")
	case errors.Is(err, responses.ErrorScopeCode["INVALID_SCOPE"]):
		responses.BadRequest(c, "invalid scope relationship or payload")
	default:
		responses.Error(c, "INTERNAL_ERROR", "internal server error", err)
	}
}

func (h *Handler) CreateScope(c *gin.Context) {
	a, ok := actor(c)
	if !ok {
		responses.Unauthorized(c, "unauthorized")
		return
	}
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	membershipID, ok := idParam(c, "membershipId")
	if !ok {
		return
	}
	var req dto.CreateMembershipScopeRequest
	if !bind(c, &req) {
		return
	}
	value, err := h.usecase.CreateScope(a, productID, membershipID, req)
	if err != nil {
		fail(c, err)
		return
	}
	var projIDVal, catIDVal int
	if value.ProjectID != nil {
		projIDVal = *value.ProjectID
	}
	if value.CategoryID != nil {
		catIDVal = *value.CategoryID
	}
	c.Set("audit_reason", fmt.Sprintf("Created access scope: Level: %s (ProjectID: %d, CategoryID: %d) for membership %d",
		value.ScopeLevel, projIDVal, catIDVal, value.MembershipID))
	c.Set("audit_payload", value)
	utils.Success(c, http.StatusCreated, value)
}

func (h *Handler) ListScopes(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	membershipID, ok := idParam(c, "membershipId")
	if !ok {
		return
	}
	values, err := h.usecase.ListScopes(a, productID, membershipID)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, values)
}

func (h *Handler) UpdateScope(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	membershipID, ok := idParam(c, "membershipId")
	if !ok {
		return
	}
	id, ok := idParam(c, "scopeId")
	if !ok {
		return
	}
	var req dto.UpdateMembershipScopeRequest
	if !bind(c, &req) {
		return
	}
	value, err := h.usecase.UpdateScope(a, productID, membershipID, id, req)
	if err != nil {
		fail(c, err)
		return
	}
	var projIDVal, catIDVal int
	if value.ProjectID != nil {
		projIDVal = *value.ProjectID
	}
	if value.CategoryID != nil {
		catIDVal = *value.CategoryID
	}
	c.Set("audit_reason", fmt.Sprintf("Updated access scope ID %d: Level: %s (ProjectID: %d, CategoryID: %d)",
		value.ScopeID, value.ScopeLevel, projIDVal, catIDVal))
	c.Set("audit_payload", value)
	utils.Success(c, http.StatusOK, value)
}

func (h *Handler) DeleteScope(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	membershipID, ok := idParam(c, "membershipId")
	if !ok {
		return
	}
	id, ok := idParam(c, "scopeId")
	if !ok {
		return
	}
	value, err := h.usecase.DeleteScope(a, productID, membershipID, id)
	if err != nil {
		fail(c, err)
		return
	}
	var projIDVal, catIDVal int
	if value.ProjectID != nil {
		projIDVal = *value.ProjectID
	}
	if value.CategoryID != nil {
		catIDVal = *value.CategoryID
	}
	c.Set("audit_reason", fmt.Sprintf("Deleted access scope ID %d: Level: %s (ProjectID: %d, CategoryID: %d)",
		value.ScopeID, value.ScopeLevel, projIDVal, catIDVal))
	c.Set("audit_payload", value)
	utils.Success(c, http.StatusOK, gin.H{"scope_id": id, "status": "deleted"})
}

package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"omnilogs-api/dto"
	"omnilogs-api/internal/sensitive_log/usecase"
	"omnilogs-api/responses"
	"omnilogs-api/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct{ usecase usecase.Usecase }

func NewHandler(usecase usecase.Usecase) *Handler { return &Handler{usecase: usecase} }

func actor(c *gin.Context) (usecase.Actor, bool) {
	id, ok := c.Get("userId")
	role, _ := c.Get("role")
	roleStr, _ := role.(string)
	isAdmin := roleStr == "god" || roleStr == "owner" || roleStr == "superadmin"
	
	val, _ := id.(uint)
	return usecase.Actor{UserID: int(val), PlatformAdmin: isAdmin}, ok
}

func idParam(c *gin.Context, name string) (int, bool) {
	value, err := strconv.Atoi(c.Param(name))
	if err != nil || value <= 0 {
		responses.BadRequest(c, "invalid "+name)
		return 0, false
	}
	return value, true
}

func (h *Handler) CreateAccessRequest(c *gin.Context) {
	act, ok := actor(c)
	if !ok {
		responses.Unauthorized(c, "unauthorized")
		return
	}
	var req dto.CreateSensitiveLogAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationError(c, err)
		return
	}
	
	res, err := h.usecase.CreateRequest(c.Request.Context(), act, req)
	if err != nil {
		responses.Error(c, "INVALID_REQUEST", err.Error(), err)
		return
	}
	var logIDStr string
	if req.LogID != nil {
		logIDStr = *req.LogID
	}
	var fieldPathStr string
	if req.FieldPath != nil {
		fieldPathStr = *req.FieldPath
	}
	c.Set("audit_reason", fmt.Sprintf("Created access request to sensitive log ID: '%s', field path: '%s'", logIDStr, fieldPathStr))
	utils.Success(c, http.StatusCreated, res)
}

func (h *Handler) ReviewAccessRequest(c *gin.Context) {
	act, ok := actor(c)
	if !ok {
		responses.Unauthorized(c, "unauthorized")
		return
	}
	requestID := c.Param("requestId")
	var req dto.ReviewSensitiveLogAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationError(c, err)
		return
	}
	
	res, err := h.usecase.ReviewRequest(c.Request.Context(), act, requestID, req)
	if err != nil {
		if errors.Is(err, responses.ErrForbidden) {
			responses.Forbidden(c, "only superadmin, owner or god roles can approve requests")
			return
		}
		responses.Error(c, "INVALID_REQUEST", err.Error(), err)
		return
	}
	c.Set("audit_reason", fmt.Sprintf("Reviewed access request %s as %s", requestID, req.ApprovalStatus))
	utils.Success(c, http.StatusOK, res)
}

func (h *Handler) ListAccessRequests(c *gin.Context) {
	act, ok := actor(c)
	if !ok {
		responses.Unauthorized(c, "unauthorized")
		return
	}
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	
	res, err := h.usecase.ListRequests(c.Request.Context(), act, productID)
	if err != nil {
		responses.Error(c, "INVALID_REQUEST", err.Error(), err)
		return
	}
	utils.Success(c, http.StatusOK, res)
}

func (h *Handler) RevealSensitiveValue(c *gin.Context) {
	act, ok := actor(c)
	if !ok {
		responses.Unauthorized(c, "unauthorized")
		return
	}
	var req dto.RevealSensitiveLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationError(c, err)
		return
	}
	
	res, err := h.usecase.RevealValue(c.Request.Context(), act, req)
	if err != nil {
		if errors.Is(err, responses.ErrForbidden) {
			responses.Forbidden(c, "permission denied")
			return
		}
		responses.Error(c, "INVALID_REQUEST", err.Error(), err)
		return
	}
	c.Set("audit_reason", fmt.Sprintf("Revealed decrypted value for request ID %s", req.RequestID))
	utils.Success(c, http.StatusOK, res)
}

func (h *Handler) ListAccessHistory(c *gin.Context) {
	act, ok := actor(c)
	if !ok {
		responses.Unauthorized(c, "unauthorized")
		return
	}
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	
	res, err := h.usecase.ListHistory(c.Request.Context(), act, productID)
	if err != nil {
		if errors.Is(err, responses.ErrForbidden) {
			responses.Forbidden(c, "permission denied")
			return
		}
		responses.Error(c, "INVALID_REQUEST", err.Error(), err)
		return
	}
	utils.Success(c, http.StatusOK, res)
}

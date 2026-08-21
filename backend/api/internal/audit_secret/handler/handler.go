package handler

import (
	"errors"
	"fmt"
	"net/http"

	"omnilogs-api/dto"
	"omnilogs-api/internal/audit_secret/usecase"
	"omnilogs-api/middleware"
	"omnilogs-api/responses"
	"omnilogs-api/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct{ usecase usecase.Usecase }

func NewHandler(usecase usecase.Usecase) *Handler { return &Handler{usecase: usecase} }

func actor(c *gin.Context) (usecase.Actor, bool) {
	id, ok := c.Get("userId")
	isAdmin := middleware.HasAdminPlatformRole(c)
	val, _ := id.(uint)
	return usecase.Actor{UserID: int(val), PlatformAdmin: isAdmin}, ok
}

func (h *Handler) CreateRequest(c *gin.Context) {
	act, ok := actor(c)
	if !ok {
		responses.Unauthorized(c, "unauthorized")
		return
	}
	auditID := c.Param("auditId")
	var req dto.CreateAuditSecretAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationError(c, err)
		return
	}
	res, err := h.usecase.CreateRequest(c.Request.Context(), act, auditID, req)
	if err != nil {
		responses.Error(c, "INVALID_REQUEST", err.Error(), err)
		return
	}
	c.Set("audit_reason", fmt.Sprintf("Created access request to audit secret %s for audit %s", req.SecretID, auditID))
	utils.Success(c, http.StatusCreated, res)
}

func (h *Handler) ReviewRequest(c *gin.Context) {
	act, ok := actor(c)
	if !ok {
		responses.Unauthorized(c, "unauthorized")
		return
	}
	auditID := c.Param("auditId")
	requestID := c.Param("requestId")
	var req dto.ReviewAuditSecretAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationError(c, err)
		return
	}
	res, err := h.usecase.ReviewRequest(c.Request.Context(), act, auditID, requestID, req)
	if err != nil {
		if errors.Is(err, responses.ErrForbidden) {
			responses.Forbidden(c, "only superadmin, owner or god roles can approve requests")
			return
		}
		responses.Error(c, "INVALID_REQUEST", err.Error(), err)
		return
	}
	c.Set("audit_reason", fmt.Sprintf("Reviewed audit secret request %s as %s", requestID, req.ApprovalStatus))
	utils.Success(c, http.StatusOK, res)
}

func (h *Handler) ListRequests(c *gin.Context) {
	auditID := c.Param("auditId")
	res, err := h.usecase.ListRequests(c.Request.Context(), auditID)
	if err != nil {
		responses.Error(c, "INVALID_REQUEST", err.Error(), err)
		return
	}
	utils.Success(c, http.StatusOK, res)
}

func (h *Handler) ListPendingRequests(c *gin.Context) {
	act, ok := actor(c)
	if !ok {
		responses.Unauthorized(c, "unauthorized")
		return
	}
	res, err := h.usecase.ListPendingRequests(c.Request.Context(), act)
	if err != nil {
		responses.Error(c, "INVALID_REQUEST", err.Error(), err)
		return
	}
	utils.Success(c, http.StatusOK, res)
}

func (h *Handler) ListSecrets(c *gin.Context) {
	auditID := c.Param("auditId")
	res, err := h.usecase.ListSecrets(c.Request.Context(), auditID)
	if err != nil {
		responses.Error(c, "INVALID_REQUEST", err.Error(), err)
		return
	}
	utils.Success(c, http.StatusOK, res)
}

func (h *Handler) RevealValue(c *gin.Context) {
	act, ok := actor(c)
	if !ok {
		responses.Unauthorized(c, "unauthorized")
		return
	}
	auditID := c.Param("auditId")
	var req dto.RevealAuditSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ValidationError(c, err)
		return
	}
	res, err := h.usecase.RevealValue(c.Request.Context(), act, auditID, req)
	if err != nil {
		if errors.Is(err, responses.ErrForbidden) {
			responses.Forbidden(c, "permission denied")
			return
		}
		responses.Error(c, "INVALID_REQUEST", err.Error(), err)
		return
	}
	c.Set("audit_reason", fmt.Sprintf("Revealed audit secret for request ID %s", req.RequestID))
	utils.Success(c, http.StatusOK, res)
}

func (h *Handler) ListHistory(c *gin.Context) {
	auditID := c.Param("auditId")
	res, err := h.usecase.ListHistory(c.Request.Context(), auditID)
	if err != nil {
		responses.Error(c, "INVALID_REQUEST", err.Error(), err)
		return
	}
	utils.Success(c, http.StatusOK, res)
}

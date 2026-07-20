package handler

import (
	"errors"
	"net/http"

	"omnilogs-api/dto"
	"omnilogs-api/internal/audit_logs/usecase"
	"omnilogs-api/middleware"
	"omnilogs-api/responses"
	"omnilogs-api/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	usecase usecase.Usecase
}

func NewHandler(usecase usecase.Usecase) *Handler {
	return &Handler{usecase: usecase}
}

func (h *Handler) List(c *gin.Context) {
	var req dto.AuditLogFilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		responses.ValidationError(c, err)
		return
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	perPage := req.PerPage
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}
	req.Page = page
	req.PerPage = perPage
	global, productIDs := middleware.AuditScope(c)
	if !global {
		req.AllowedProductIDs = productIDs
	}

	values, total, err := h.usecase.List(c.Request.Context(), req)
	if err != nil {
		responses.Error(c, "INTERNAL_ERROR", "failed to query audit logs", err)
		return
	}

	responses.SuccessWithMeta(c, http.StatusOK, "AUDIT_LOGS_RETRIEVED", values, utils.NewPaginationMeta(page, perPage, total))
}

func (h *Handler) GetByID(c *gin.Context) {
	value, err := h.usecase.FindByID(c.Request.Context(), c.Param("auditId"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			responses.NotFound(c, "audit log not found")
			return
		}
		responses.Error(c, "INTERNAL_ERROR", "failed to retrieve audit log", err)
		return
	}
	if !middleware.CanReadAuditProduct(c, value.ProductID) {
		responses.Forbidden(c, "audit log access denied")
		return
	}
	responses.Success(c, http.StatusOK, "AUDIT_LOG_RETRIEVED", value)
}

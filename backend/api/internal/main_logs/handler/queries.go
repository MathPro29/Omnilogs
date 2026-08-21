package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"omnilogs-api/internal/main_logs/usecase"
	"omnilogs-api/middleware"
	"omnilogs-api/responses"
	"omnilogs-api/utils"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Search(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		responses.Unauthorized(c, "unauthorized")
		return
	}

	productID, err := strconv.ParseInt(c.Query("product_id"), 10, 64)
	if err != nil || productID <= 0 {
		responses.BadRequest(c, "product_id is required")
		return
	}
	environmentID := parseOptionalInt64(c.Query("environment_id"))

	page := parseIntWithDefault(c.Query("page"), 1)
	perPage := parseIntWithDefault(c.Query("per_page"), 20)
	if perPage > 100 {
		perPage = 100
	}

	input := usecase.SearchInput{
		ActorUserID:      userID,
		PlatformAdmin:    middleware.HasAdminPlatformRole(c),
		ProductID:        productID,
		Page:             page,
		PerPage:          perPage,
		EnvironmentID:    environmentID,
		ProjectID:        parseOptionalInt64(c.Query("project_id")),
		ProjectIDs:       parseCSVInt64(c.Query("project_ids")),
		CategoryID:       parseOptionalInt64(c.Query("category_id")),
		CategoryIDs:      parseCSVInt64(c.Query("category_ids")),
		Level:            stringParam(c.Query("level")),
		LogType:          stringParam(c.Query("log_type")),
		RequestID:        stringParam(c.Query("request_id")),
		TraceID:          stringParam(c.Query("trace_id")),
		DateFrom:         parseOptionalTime(c.Query("date_from")),
		StatusCode:       parseOptionalInt64(c.Query("status_code")),
		SourceID:         parseOptionalInt64(c.Query("source_id")),
		CustomFieldPath:  stringParam(c.Query("custom_field_path")),
		CustomFieldValue: stringParam(c.Query("custom_field_value")),
		SortField:        stringParam(c.Query("sort_field")),
		SortOrder:        c.Query("sort_order"),
		Keyword:          stringParam(c.Query("keyword")),
	}

	result, err := h.usecase.Search(c.Request.Context(), input, requestID(c), traceID(c), ipAddress(c), userAgent(c))
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			responses.Error(c, "TIMEOUT", "log search timed out", err)
			return
		}
		if errors.Is(err, usecase.ErrInvalidSearchFilter) {
			responses.BadRequest(c, err.Error())
			return
		}
		if errors.Is(err, usecase.ErrRestoredArchiveUnavailable) {
			responses.NotFound(c, "restored archive is not available for exploration")
			return
		}
		if errors.Is(err, responses.ErrForbidden) {
			responses.Forbidden(c, "forbidden")
			return
		}
		responses.Error(c, "INTERNAL_ERROR", "failed to search logs", err)
		return
	}

	responses.SuccessWithMeta(c, http.StatusOK, "LOGS_RETRIEVED", result.Items, utils.NewPaginationMeta(page, perPage, result.Total))
}

func (h *Handler) SearchV2(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		responses.Unauthorized(c, "unauthorized")
		return
	}

	var input usecase.DynamicSearchInput
	if err := c.ShouldBindJSON(&input); err != nil {
		responses.BadRequest(c, err.Error())
		return
	}

	if input.Scope.ProductID <= 0 {
		responses.BadRequest(c, "product_id is required in scope")
		return
	}

	input.ActorUserID = userID
	input.PlatformAdmin = middleware.HasAdminPlatformRole(c)

	page := input.Page
	if page <= 0 {
		page = 1
	}
	perPage := input.PageSize
	if perPage <= 0 {
		perPage = 20
	} else if perPage > 100 {
		perPage = 100
	}
	input.Page = page
	input.PageSize = perPage

	result, err := h.usecase.SearchV2(c.Request.Context(), input, requestID(c), traceID(c), ipAddress(c), userAgent(c))
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			responses.Error(c, "TIMEOUT", "log search timed out", err)
			return
		}
		if errors.Is(err, usecase.ErrInvalidSearchFilter) {
			responses.BadRequest(c, err.Error())
			return
		}
		if errors.Is(err, usecase.ErrRestoredArchiveUnavailable) {
			responses.NotFound(c, "restored archive is not available for exploration")
			return
		}
		if errors.Is(err, responses.ErrForbidden) {
			responses.Forbidden(c, "forbidden")
			return
		}
		responses.Error(c, "INTERNAL_ERROR", "failed to search logs", err)
		return
	}

	responses.SuccessWithMeta(c, http.StatusOK, "LOGS_RETRIEVED", result.Items, utils.NewPaginationMeta(page, perPage, result.Total))
}

func (h *Handler) GetByID(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		responses.Unauthorized(c, "unauthorized")
		return
	}

	productID, err := strconv.ParseInt(c.Query("product_id"), 10, 64)
	if err != nil || productID <= 0 {
		responses.BadRequest(c, "product_id is required")
		return
	}

	value, err := h.usecase.FindByID(c.Request.Context(), userID, middleware.HasAdminPlatformRole(c), productID, c.Param("logId"), requestID(c), traceID(c), ipAddress(c), userAgent(c))
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			responses.Error(c, "TIMEOUT", "log detail query timed out", err)
		case errors.Is(err, responses.ErrForbidden):
			responses.Forbidden(c, "forbidden")
		case errors.Is(err, usecase.ErrMainLogNotFound):
			responses.NotFound(c, "main log not found")
		default:
			responses.Error(c, "INTERNAL_ERROR", "failed to retrieve main log", err)
		}
		return
	}

	responses.Success(c, http.StatusOK, "LOG_RETRIEVED", value)
}

func (h *Handler) GetByAudit(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		responses.Unauthorized(c, "unauthorized")
		return
	}

	audit, value, reason, err := h.usecase.FindByAudit(c.Request.Context(), userID, middleware.HasAdminPlatformRole(c), c.Param("auditId"), requestID(c), traceID(c), ipAddress(c), userAgent(c))
	if err != nil {
		if errors.Is(err, responses.ErrForbidden) {
			responses.Forbidden(c, "forbidden")
			return
		}
		responses.Error(c, "INTERNAL_ERROR", "failed to retrieve linked main log", err)
		return
	}

	data := gin.H{
		"audit": audit,
		"main_log": gin.H{
			"available": value != nil,
		},
	}
	if value != nil {
		data["main_log"] = gin.H{"available": true, "data": value}
	} else if reason != "" {
		data["main_log"] = gin.H{"available": false, "reason": reason}
	}

	responses.Success(c, http.StatusOK, "AUDIT_MAIN_LOG_RETRIEVED", data)
}

func parseOptionalTime(value string) *time.Time {
	if value == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}
	return &parsed
}

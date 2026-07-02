package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"omnilogs-api/internal/main_logs/usecase"
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

	page := parseIntWithDefault(c.Query("page"), 1)
	perPage := parseIntWithDefault(c.Query("per_page"), 20)
	if perPage > 100 {
		perPage = 100
	}

	input := usecase.SearchInput{
		ActorUserID: userID,
		PlatformAdmin: middleware.HasAdminPlatformRole(c),
		ProductID:   productID,
		Page:        page,
		PerPage:     perPage,
		EnvironmentID: parseOptionalInt64(c.Query("environment_id")),
		ProjectID:   parseOptionalInt64(c.Query("project_id")),
		ProjectIDs:  parseCSVInt64(c.Query("project_ids")),
		CategoryID:  parseOptionalInt64(c.Query("category_id")),
		CategoryIDs: parseCSVInt64(c.Query("category_ids")),
		Level:       stringParam(c.Query("level")),
		LogType:     stringParam(c.Query("log_type")),
		RequestID:   stringParam(c.Query("request_id")),
		TraceID:     stringParam(c.Query("trace_id")),
		Keyword:     stringParam(c.Query("keyword")),
	}

	result, err := h.usecase.Search(c.Request.Context(), input, requestID(c), traceID(c), ipAddress(c), userAgent(c))
	if err != nil {
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

func parseIntWithDefault(raw string, fallback int) int {
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func parseOptionalInt64(raw string) *int64 {
	if raw == "" {
		return nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return nil
	}
	return &value
}

func stringParam(raw string) *string {
	if raw == "" {
		return nil
	}
	return &raw
}

func requestID(c *gin.Context) *string {
	value, ok := c.Get(middleware.ContextRequestID)
	if !ok {
		return nil
	}
	raw, ok := value.(string)
	if !ok || raw == "" {
		return nil
	}
	return &raw
}

func traceID(c *gin.Context) *string {
	raw := c.GetHeader("X-Trace-ID")
	if raw == "" {
		raw = c.GetHeader("Traceparent")
	}
	if raw == "" {
		return nil
	}
	return &raw
}

func ipAddress(c *gin.Context) *string {
	raw := c.ClientIP()
	if raw == "" {
		return nil
	}
	return &raw
}

func userAgent(c *gin.Context) *string {
	raw := c.Request.UserAgent()
	if raw == "" {
		return nil
	}
	return &raw
}

func parseCSVInt64(raw string) []int64 {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]int64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if v, err := strconv.ParseInt(part, 10, 64); err == nil && v > 0 {
			result = append(result, v)
		}
	}
	return result
}

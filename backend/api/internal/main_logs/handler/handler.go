package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"omnilogs-api/configs"
	"omnilogs-api/internal/main_logs/usecase"
	"omnilogs-api/middleware"
	"omnilogs-api/responses"
	"omnilogs-api/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	usecase   usecase.Usecase
	natsQueue *configs.NATSQueue
}

func NewHandler(usecase usecase.Usecase, natsQueue *configs.NATSQueue) *Handler {
	return &Handler{usecase: usecase, natsQueue: natsQueue}
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
		ActorUserID:   userID,
		PlatformAdmin: middleware.HasAdminPlatformRole(c),
		ProductID:     productID,
		Page:          page,
		PerPage:       perPage,
		EnvironmentID: parseOptionalInt64(c.Query("environment_id")),
		ProjectID:     parseOptionalInt64(c.Query("project_id")),
		ProjectIDs:    parseCSVInt64(c.Query("project_ids")),
		CategoryID:    parseOptionalInt64(c.Query("category_id")),
		CategoryIDs:   parseCSVInt64(c.Query("category_ids")),
		Level:         stringParam(c.Query("level")),
		LogType:       stringParam(c.Query("log_type")),
		RequestID:     stringParam(c.Query("request_id")),
		TraceID:       stringParam(c.Query("trace_id")),
		Keyword:       stringParam(c.Query("keyword")),
	}

	result, err := h.usecase.Search(c.Request.Context(), input, requestID(c), traceID(c), ipAddress(c), userAgent(c))
	if err != nil {
		if errors.Is(err, responses.ErrForbidden) {
			responses.Forbidden(c, "forbidden")
			return
		}
		slog.ErrorContext(c.Request.Context(), "failed to search logs", "error", err, "product_id", productID)
		responses.Error(c, "INTERNAL_ERROR", "failed to search logs", nil)
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
			slog.ErrorContext(c.Request.Context(), "failed to retrieve main log", "error", err, "product_id", productID, "log_id", c.Param("logId"))
			responses.Error(c, "INTERNAL_ERROR", "failed to retrieve main log", nil)
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
		slog.ErrorContext(c.Request.Context(), "failed to retrieve linked main log", "error", err, "audit_id", c.Param("auditId"))
		responses.Error(c, "INTERNAL_ERROR", "failed to retrieve linked main log", nil)
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

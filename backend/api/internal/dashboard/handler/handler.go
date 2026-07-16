package handler

import (
	"context"
	"errors"
	"net/http"

	"omnilogs-api/dto"
	"omnilogs-api/internal/dashboard/usecase"
	"omnilogs-api/middleware"
	"omnilogs-api/responses"

	"github.com/gin-gonic/gin"
)

type Handler interface {
	GetLogs(c *gin.Context)
	GetLogStats(c *gin.Context)
	GetLogDetail(c *gin.Context)
	GetAuditLogs(c *gin.Context)
}

type handler struct {
	uc usecase.Usecase
}

func NewHandler(uc usecase.Usecase) Handler {
	return &handler{
		uc: uc,
	}
}

func (h *handler) GetLogs(c *gin.Context) {
	var query dto.LogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query = usecase.NormalizeLogQuery(query)

	result, err := h.uc.GetLogs(c.Request.Context(), query)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			responses.Error(c, "TIMEOUT", "log query timed out", err)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *handler) GetLogStats(c *gin.Context) {
	var query dto.LogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query = usecase.NormalizeLogQuery(query)

	result, err := h.uc.GetLogStats(c.Request.Context(), query)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			responses.Error(c, "TIMEOUT", "statistics query timed out", err)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *handler) GetLogDetail(c *gin.Context) {
	indexName := c.Param("index")
	logID := c.Param("logId")

	result, err := h.uc.GetLogDetail(c.Request.Context(), indexName, logID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			responses.Error(c, "TIMEOUT", "log detail query timed out", err)
			return
		}
		if err.Error() == "log not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "log not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *handler) GetAuditLogs(c *gin.Context) {
	var query dto.AuditLogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query = usecase.NormalizeAuditLogQuery(query)
	global, productIDs := middleware.AuditScope(c)
	if !global {
		query.AllowedProductIDs = productIDs
	}

	auditLogs, total, err := h.uc.GetAuditLogs(c.Request.Context(), query)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			responses.Error(c, "TIMEOUT", "audit log query timed out", err)
			return
		}
		responses.Error(c, "INTERNAL_SERVER_ERROR", "failed to query audit logs", err)
		return
	}

	responses.Success(c, http.StatusOK, "AUDIT_LOGS_RETRIEVED", gin.H{
		"total":  total,
		"limit":  query.Limit,
		"offset": query.Offset,
		"data":   auditLogs,
	})
}

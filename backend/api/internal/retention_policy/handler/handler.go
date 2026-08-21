package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"omnilogs-api/dto"
	"omnilogs-api/internal/retention_policy/repository"
	"omnilogs-api/internal/retention_policy/usecase"
	"omnilogs-api/middleware"
	"omnilogs-api/responses"

	"github.com/gin-gonic/gin"
)

type Handler interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	GetByID(c *gin.Context)
	ToggleActive(c *gin.Context)
	TriggerNow(c *gin.Context)
	Simulate(c *gin.Context)
	GetStats(c *gin.Context)
	Historical(c *gin.Context)
	Preview(c *gin.Context)
}

type handler struct {
	usecase usecase.Usecase
}

func NewHandler(usecase usecase.Usecase) Handler {
	return &handler{usecase: usecase}
}

func (h *handler) Create(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var req dto.CreateRetentionPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ProductID = productID
	c.Set("audit_action", "RETENTION_POLICY_CREATED")

	res, err := h.usecase.Create(withActor(c), req)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (h *handler) Update(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	policyID, err := strconv.Atoi(c.Param("policyId"))
	if err != nil || policyID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}

	var req dto.UpdateRetentionPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ProductID = productID
	req.PolicyID = policyID
	c.Set("audit_action", "RETENTION_POLICY_UPDATED")

	res, err := h.usecase.Update(withActor(c), req)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *handler) Delete(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	policyID, err := strconv.Atoi(c.Param("policyId"))
	if err != nil || policyID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}

	c.Set("audit_action", "RETENTION_POLICY_DELETED")
	if err := h.usecase.Delete(withActor(c), policyID, productID); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "retention policy deleted successfully"})
}

func (h *handler) List(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var envID *int
	if val := c.Query("environment_id"); val != "" {
		if id, convErr := strconv.Atoi(val); convErr == nil && id > 0 {
			envID = &id
		}
	}
	if envID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "environment_id query parameter is required"})
		return
	}

	policies, err := h.usecase.List(withActor(c), productID, envID)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, policies)
}

func (h *handler) GetByID(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	policyID, err := strconv.Atoi(c.Param("policyId"))
	if err != nil || policyID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}

	policy, err := h.usecase.GetByID(withActor(c), policyID, productID)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, policy)
}

func (h *handler) ToggleActive(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	policyID, err := strconv.Atoi(c.Param("policyId"))
	if err != nil || policyID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}

	var body struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.usecase.ToggleActive(withActor(c), policyID, productID, body.IsActive); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "policy active status updated"})
}

func (h *handler) TriggerNow(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	policyID, err := strconv.Atoi(c.Param("policyId"))
	if err != nil || policyID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}
	c.Set("audit_action", "RETENTION_POLICY_TRIGGERED")
	result, err := h.usecase.TriggerNow(withActor(c), policyID, productID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, result)
}

func (h *handler) Simulate(c *gin.Context) {
	var req dto.SimulateRetentionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.usecase.Simulate(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *handler) GetStats(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var environmentID *int
	if value := c.Query("environment_id"); value != "" {
		parsed, parseErr := strconv.Atoi(value)
		if parseErr != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid environment id"})
			return
		}
		environmentID = &parsed
	}

	stats, err := h.usecase.GetStats(withActor(c), productID, environmentID)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *handler) Historical(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	var req dto.RetentionHistoricalRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ProductID = productID
	result, err := h.usecase.Historical(withActor(c), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *handler) Preview(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	var req dto.RetentionPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ProductID = productID
	result, err := h.usecase.Preview(withActor(c), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func withActor(c *gin.Context) context.Context {
	id, _ := middleware.CurrentUserID(c)
	return usecase.WithActor(c.Request.Context(), usecase.Actor{UserID: int(id), PlatformAdmin: middleware.HasAdminPlatformRole(c)})
}

func writeError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	if errors.Is(err, repository.ErrActivePolicyExists) {
		status = http.StatusConflict
	}
	if errors.Is(err, responses.ErrForbidden) {
		status = http.StatusForbidden
	}
	if errors.Is(err, responses.ErrNotFound) {
		status = http.StatusNotFound
	}
	if status == http.StatusInternalServerError && (errors.Is(err, responses.ErrInvalid) || strings.Contains(strings.ToLower(err.Error()), "required") || strings.Contains(strings.ToLower(err.Error()), "invalid") || strings.Contains(strings.ToLower(err.Error()), "must be") || strings.Contains(strings.ToLower(err.Error()), "does not belong")) {
		status = http.StatusBadRequest
	}
	c.JSON(status, gin.H{"error": err.Error()})
}

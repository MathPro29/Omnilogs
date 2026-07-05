package handler

import (
	"fmt"
	"net/http"
	"omnilogs-api/dto"
	"omnilogs-api/internal/elastic_index_policy/usecase"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	GetByID(c *gin.Context)
	PushToArchives(c *gin.Context)
	ClearAllLogs(c *gin.Context)
}

type handler struct {
	usecase usecase.Usecase
}

func NewHandler(usecase usecase.Usecase) Handler {
	return &handler{usecase: usecase}
}

func (h *handler) Create(c *gin.Context) {
	request := dto.CreateElasticIndexPolicyRequest{}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	request.ProductID = productID

	// validation rollovertType
	if request.RolloverType != "size" && request.RolloverType != "age" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rollover type"})
		return
	}

	res, err := h.usecase.Create(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *handler) Update(c *gin.Context) {
	request := dto.UpdateElasticIndexPolicyRequest{}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.usecase.Update(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *handler) Delete(c *gin.Context) {
	request := dto.DeleteElasticIndexPolicyRequest{}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.usecase.Delete(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *handler) List(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	query := dto.ListElasticIndexPolicyRequest{ProductID: productID}
	if value := c.Query("environment_id"); value != "" {
		environmentID, err := strconv.Atoi(value)
		if err != nil || environmentID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid environment id"})
			return
		}
		query.EnvironmentID = &environmentID
	}

	policies, err := h.usecase.List(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, policies)
}

func (h *handler) GetByID(c *gin.Context) {
	request := dto.GetElasticIndexPolicyRequest{}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.usecase.GetByID(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *handler) PushToArchives(c *gin.Context) {
	roleValue, ok := c.Get("role")
	role := strings.ToLower(fmt.Sprint(roleValue))
	if !ok || (role != "god" && role != "owner") {
		c.JSON(http.StatusForbidden, gin.H{"error": "only god or owner can push logs to archives"})
		return
	}
	request := dto.PushToArchivesRequest{}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	policyID, err := strconv.Atoi(c.Param("elasticPolicyId"))
	if err != nil || policyID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}
	request.ProductID = productID
	request.ElasticPolicyID = policyID

	res, err := h.usecase.PushToArchives(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *handler) ClearAllLogs(c *gin.Context) {
	roleValue, ok := c.Get("role")
	role := strings.ToLower(fmt.Sprint(roleValue))
	if !ok || (role != "god" && role != "owner") {
		c.JSON(http.StatusForbidden, gin.H{"error": "only god or owner can clear logs"})
		return
	}
	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	err = h.usecase.ClearAllLogs(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully cleared all logs and indices"})
}

package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"omnilogs-api/middleware"
	"omnilogs-api/models"
)

type Handler struct{ db *gorm.DB }

func New(db *gorm.DB) *Handler { return &Handler{db: db} }

type routeInput struct {
	EnvironmentID int    `json:"environment_id"`
	RouteKey      string `json:"route_key"`
	ProjectID     int    `json:"project_id"`
	CategoryID    *int   `json:"category_id"`
	Priority      int    `json:"priority"`
}

func (h *Handler) allowed(c *gin.Context, productID int) bool {
	if middleware.HasAdminPlatformRole(c) {
		return true
	}
	uid, ok := middleware.CurrentUserID(c)
	if !ok {
		return false
	}
	var count int64
	h.db.Table("product_memberships").Where("product_id = ? AND user_id = ? AND is_active = TRUE", productID, uid).Count(&count)
	return count > 0
}
func (h *Handler) List(c *gin.Context) {
	productID := c.Param("productId")
	var rows []models.LogRoute
	if !h.allowed(c, atoi(productID)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}
	if err := h.db.Where("product_id = ? AND is_active = TRUE", atoi(productID)).Order("environment_id, priority DESC, route_id").Find(&rows).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}
func (h *Handler) Create(c *gin.Context) {
	productID := atoi(c.Param("productId"))
	if !h.allowed(c, productID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}
	var input routeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "invalid route payload"})
		return
	}
	input.RouteKey = strings.TrimSpace(input.RouteKey)
	if productID <= 0 || input.EnvironmentID <= 0 || input.ProjectID <= 0 || input.RouteKey == "" {
		c.JSON(400, gin.H{"error": "route key, environment and project are required"})
		return
	}
	var env models.ProductEnvironment
	if err := h.db.Where("environment_id = ? AND product_id = ? AND is_active = TRUE", input.EnvironmentID, productID).First(&env).Error; err != nil {
		c.JSON(400, gin.H{"error": "environment does not belong to product"})
		return
	}
	var project models.Project
	if err := h.db.Where("project_id = ? AND product_id = ? AND is_active = TRUE", input.ProjectID, productID).First(&project).Error; err != nil {
		c.JSON(400, gin.H{"error": "project does not belong to product"})
		return
	}
	if input.CategoryID != nil {
		var feature models.ProjectFeature
		if err := h.db.Where("category_id = ? AND project_id = ? AND product_id = ? AND is_active = TRUE", *input.CategoryID, input.ProjectID, productID).First(&feature).Error; err != nil {
			c.JSON(400, gin.H{"error": "feature does not belong to selected project"})
			return
		}
	}
	var existing models.LogRoute
	if err := h.db.Where("product_id = ? AND environment_id = ? AND LOWER(BTRIM(route_key)) = LOWER(BTRIM(?))", productID, input.EnvironmentID, input.RouteKey).First(&existing).Error; err == nil {
		if existing.IsActive {
			c.JSON(409, gin.H{"error": "route_key already exists for this environment"})
			return
		}
		existing.ProjectID = input.ProjectID
		existing.CategoryID = input.CategoryID
		existing.Priority = input.Priority
		existing.IsActive = true
		if err := h.db.Save(&existing).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": existing})
		return
	}
	row := models.LogRoute{ProductID: productID, EnvironmentID: input.EnvironmentID, RouteKey: input.RouteKey, ProjectID: input.ProjectID, CategoryID: input.CategoryID, Priority: input.Priority, IsActive: true}
	if err := h.db.Create(&row).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": row})
}
func (h *Handler) Delete(c *gin.Context) {
	productID := atoi(c.Param("productId"))
	routeID := atoi(c.Param("routeId"))
	if !h.allowed(c, productID) {
		c.JSON(403, gin.H{"error": "permission denied"})
		return
	}
	result := h.db.Model(&models.LogRoute{}).
		Where("route_id = ? AND product_id = ? AND is_active = TRUE", routeID, productID).
		Update("is_active", false)
	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "route mapping not found"})
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"route_id": routeID, "is_active": false}})
}
func atoi(value string) int {
	n := 0
	for _, r := range value {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
	}
	return n
}

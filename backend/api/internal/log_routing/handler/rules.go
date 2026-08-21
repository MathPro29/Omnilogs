package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"omnilogs-api/models"
)

type routingConditionInput struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
	Min      any    `json:"min,omitempty"`
	Max      any    `json:"max,omitempty"`
}

type routingRuleInput struct {
	EnvironmentID    *int                    `json:"environment_id"`
	RuleName         string                  `json:"rule_name"`
	Priority         int                     `json:"priority"`
	Conditions       []routingConditionInput `json:"conditions"`
	TargetProjectID  int                     `json:"target_project_id"`
	TargetCategoryID *int                    `json:"target_category_id"`
}

var routingFieldPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*(\.[a-zA-Z][a-zA-Z0-9_]*)*$`)

// ListRules ส่งคืน Mapping ตามข้อมูลที่ connector ส่งมา เช่น service, request_path และ request_method
func (h *Handler) ListRules(c *gin.Context) {
	productID := atoi(c.Param("productId"))
	if !h.allowed(c, productID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	var rows []models.LogRoutingRule
	if err := h.db.Where("product_id = ? AND is_active = TRUE", productID).Order("priority DESC, rule_id").Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

func (h *Handler) CreateRule(c *gin.Context) {
	productID := atoi(c.Param("productId"))
	if !h.allowed(c, productID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	var input routingRuleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid routing rule payload"})
		return
	}
	input.RuleName = strings.TrimSpace(input.RuleName)
	if productID <= 0 || input.RuleName == "" || input.TargetProjectID <= 0 || len(input.Conditions) == 0 || !validRoutingConditions(input.Conditions) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rule name, target project and valid conditions are required"})
		return
	}
	if input.EnvironmentID != nil {
		var environment models.ProductEnvironment
		if err := h.db.Where("environment_id = ? AND product_id = ? AND is_active = TRUE", *input.EnvironmentID, productID).First(&environment).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "environment does not belong to product"})
			return
		}
	}
	if !h.validRouteTarget(productID, input.TargetProjectID, input.TargetCategoryID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target project or feature does not belong to product"})
		return
	}

	conditions, err := json.Marshal(input.Conditions)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid routing conditions"})
		return
	}
	row := models.LogRoutingRule{ProductID: productID, EnvironmentID: input.EnvironmentID, RuleName: input.RuleName, Priority: input.Priority, Conditions: conditions, TargetProjectID: input.TargetProjectID, TargetCategoryID: input.TargetCategoryID, IsActive: true}
	if err := h.db.Create(&row).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "routing rule already exists or could not be created"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": row})
}

func (h *Handler) UpdateRule(c *gin.Context) {
	productID := atoi(c.Param("productId"))
	ruleID := atoi(c.Param("ruleId"))
	if !h.allowed(c, productID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	var input routingRuleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid routing rule payload"})
		return
	}
	input.RuleName = strings.TrimSpace(input.RuleName)
	if productID <= 0 || ruleID <= 0 || input.RuleName == "" || input.TargetProjectID <= 0 || len(input.Conditions) == 0 || !validRoutingConditions(input.Conditions) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rule name, target project and valid conditions are required"})
		return
	}
	if input.EnvironmentID != nil {
		var environment models.ProductEnvironment
		if err := h.db.Where("environment_id = ? AND product_id = ? AND is_active = TRUE", *input.EnvironmentID, productID).First(&environment).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "environment does not belong to product"})
			return
		}
	}
	if !h.validRouteTarget(productID, input.TargetProjectID, input.TargetCategoryID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target project or feature does not belong to product"})
		return
	}
	conditions, err := json.Marshal(input.Conditions)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid routing conditions"})
		return
	}
	var rule models.LogRoutingRule
	if err := h.db.Where("rule_id = ? AND product_id = ? AND is_active = TRUE", ruleID, productID).First(&rule).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "routing rule not found"})
		return
	}
	rule.EnvironmentID = input.EnvironmentID
	rule.RuleName = input.RuleName
	rule.Priority = input.Priority
	rule.Conditions = conditions
	rule.TargetProjectID = input.TargetProjectID
	rule.TargetCategoryID = input.TargetCategoryID
	if err := h.db.Save(&rule).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "routing rule could not be updated"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rule})
}
func (h *Handler) DeleteRule(c *gin.Context) {
	productID := atoi(c.Param("productId"))
	ruleID := atoi(c.Param("ruleId"))
	if !h.allowed(c, productID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	result := h.db.Model(&models.LogRoutingRule{}).
		Where("rule_id = ? AND product_id = ? AND is_active = TRUE", ruleID, productID).
		Update("is_active", false)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "routing rule not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"rule_id": ruleID}})
}

func (h *Handler) validRouteTarget(productID, projectID int, categoryID *int) bool {
	var project models.Project
	if err := h.db.Where("project_id = ? AND product_id = ? AND is_active = TRUE", projectID, productID).First(&project).Error; err != nil {
		return false
	}
	if categoryID == nil {
		return true
	}

	var feature models.ProjectFeature
	return h.db.Where("category_id = ? AND project_id = ? AND product_id = ? AND is_active = TRUE", *categoryID, projectID, productID).First(&feature).Error == nil
}

func validRoutingConditions(conditions []routingConditionInput) bool {
	allowedOperators := map[string]bool{
		"equals": true, "not_equals": true, "starts_with": true,
		"ends_with": true, "contains": true, "in": true,
		"not_in": true, "regex": true, "range": true,
	}
	for _, condition := range conditions {
		condition.Field = strings.TrimSpace(condition.Field)
		condition.Operator = strings.ToLower(strings.TrimSpace(condition.Operator))
		if !routingFieldPattern.MatchString(condition.Field) || !allowedOperators[condition.Operator] {
			return false
		}
		if condition.Operator == "range" {
			if condition.Min == nil || condition.Max == nil {
				return false
			}
			continue
		}
		if condition.Value == nil {
			return false
		}
	}
	return true
}

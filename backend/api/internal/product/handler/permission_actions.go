package handler

import (
	"net/http"

	"omnilogs-api/dto"
	"omnilogs-api/responses"
	"omnilogs-api/utils"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreatePermissionRule(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	var req dto.CreatePermissionRuleRequest
	if !bind(c, &req) {
		return
	}
	if req.ProductID != nil && *req.ProductID != productID {
		responses.BadRequest(c, "product_id does not match route")
		return
	}
	req.ProductID = &productID
	value, err := h.usecase.CreatePermissionRule(a, req)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, value)
}
func (h *Handler) ListPermissionRules(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	values, err := h.usecase.ListPermissionRules(a, productID)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, values)
}
func (h *Handler) UpdatePermissionRule(c *gin.Context) {
	a, _ := actor(c)
	id, ok := idParam(c, "ruleId")
	if !ok {
		return
	}
	var req dto.UpdatePermissionRuleRequest
	if !bind(c, &req) {
		return
	}
	value, err := h.usecase.UpdatePermissionRule(a, id, req)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, value)
}
func (h *Handler) CheckPermission(c *gin.Context) {
	a, _ := actor(c)
	var req dto.PermissionCheckRequest
	if !bind(c, &req) {
		return
	}
	value, err := h.usecase.CheckPermission(a, req)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, value)
}

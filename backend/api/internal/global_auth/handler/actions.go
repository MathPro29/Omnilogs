package handler

import (
	"net/http"

	"omnilogs-api/dto"
	"omnilogs-api/responses"
	"omnilogs-api/utils"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateRole(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	var req dto.CreateProductRoleRequest
	if !bind(c, &req) {
		return
	}
	value, err := h.usecase.CreateRole(a, productID, req)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, value)
}

func (h *Handler) ListRoles(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	values, err := h.usecase.ListRoles(a, productID)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, values)
}

func (h *Handler) UpdateRole(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	id, ok := idParam(c, "roleId")
	if !ok {
		return
	}
	var req dto.UpdateRoleRequest
	if !bind(c, &req) {
		return
	}
	value, err := h.usecase.UpdateRole(a, productID, id, req)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, value)
}

func (h *Handler) DeleteRole(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	id, ok := idParam(c, "roleId")
	if !ok {
		return
	}
	if err := h.usecase.DeleteRole(a, productID, id); err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"role_id": id, "status": "deleted"})
}

func (h *Handler) CreateMembership(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	var req dto.CreateProductMembershipRequest
	if !bind(c, &req) {
		return
	}
	value, err := h.usecase.CreateMembership(a, productID, req)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, value)
}

func (h *Handler) ListMemberships(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	values, err := h.usecase.ListMemberships(a, productID)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, values)
}

func (h *Handler) UpdateMembership(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	id, ok := idParam(c, "membershipId")
	if !ok {
		return
	}
	var req dto.UpdateProductMembershipRequest
	if !bind(c, &req) {
		return
	}
	value, err := h.usecase.UpdateMembership(a, productID, id, req)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, value)
}

func (h *Handler) DeleteMembership(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	id, ok := idParam(c, "membershipId")
	if !ok {
		return
	}
	if err := h.usecase.DeleteMembership(a, productID, id); err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"membership_id": id, "status": "deleted"})
}

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

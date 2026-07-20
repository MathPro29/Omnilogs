package handler

import (
	"fmt"
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
	c.Set("audit_reason", fmt.Sprintf("Created product role '%s' (Code: %s)", value.RoleName, value.RoleCode))
	c.Set("audit_payload", value)
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
	c.Set("audit_reason", fmt.Sprintf("Updated product role '%s' (ID: %d)", value.RoleName, value.RoleID))
	c.Set("audit_payload", value)
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
	role, err := h.usecase.DeleteRole(a, productID, id)
	if err != nil {
		fail(c, err)
		return
	}
	c.Set("audit_reason", fmt.Sprintf("Deleted product role '%s' (ID: %d)", role.RoleName, role.RoleID))
	c.Set("audit_payload", role)
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

func (h *Handler) CreateMemberships(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	var req dto.CreateBulkProductMembershipRequest
	if !bind(c, &req) {
		return
	}
	values, err := h.usecase.CreateMemberships(a, productID, req)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, values)
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
	c.Set("audit_reason", fmt.Sprintf("Granted %s permission on %s to user %d", req.Action, req.ResourceType, req.UserID))
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
	c.Set("audit_reason", fmt.Sprintf("Updated permission rule ID %d", id))
	utils.Success(c, http.StatusOK, value)
}

func (h *Handler) GetProductAccessOverview(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	value, err := h.usecase.GetProductAccessOverview(a, productID)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, value)
}

func (h *Handler) UpsertProductAccess(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	userID, ok := idParam(c, "userId")
	if !ok {
		return
	}
	var req dto.UpsertProductAccessRequest
	if !bind(c, &req) {
		return
	}
	value, err := h.usecase.UpsertProductAccess(a, productID, userID, req)
	if err != nil {
		fail(c, err)
		return
	}
	c.Set("audit_reason", fmt.Sprintf("Upserted centralized product access for user %d in product %d", userID, productID))
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

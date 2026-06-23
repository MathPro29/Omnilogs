package handler

import (
	"net/http"

	"omnilogs-api/dto"
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

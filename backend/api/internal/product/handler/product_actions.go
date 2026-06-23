package handler

import (
	"net/http"

	"omnilogs-api/dto"
	"omnilogs-api/responses"
	"omnilogs-api/utils"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateProduct(c *gin.Context) {
	a, ok := actor(c)
	if !ok {
		responses.Unauthorized(c, "unauthorized")
		return
	}
	var req dto.CreateProductRequest
	if !bind(c, &req) {
		return
	}
	value, err := h.usecase.CreateProduct(a, req)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, value)
}
func (h *Handler) ListProducts(c *gin.Context) {
	a, _ := actor(c)
	values, err := h.usecase.ListProducts(a)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, values)
}
func (h *Handler) GetProduct(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	value, err := h.usecase.GetProduct(a, productID)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, value)
}
func (h *Handler) UpdateProduct(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	var req dto.UpdateProductRequest
	if !bind(c, &req) {
		return
	}
	value, err := h.usecase.UpdateProduct(a, productID, req)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, value)
}

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

func (h *Handler) DeleteProduct(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	if err := h.usecase.DeleteProduct(a, productID); err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"product_id": productID, "status": "deleted"})
}

func (h *Handler) RestoreProduct(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	value, err := h.usecase.RestoreProduct(a, productID)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, value)
}

// bulk delete product
func (h *Handler) BulkDeleteProducts(c *gin.Context) {
	a, _ := actor(c)
	var req dto.BulkDeleteProductRequest
	if !bind(c, &req) {
		return
	}
	if err := h.usecase.BulkDeleteProducts(a, req.ProductIDs); err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"product_ids": req.ProductIDs, "status": "deleted"})
}

func (h *Handler) GetSetupStatus(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	res, err := h.usecase.GetSetupStatus(a, productID)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, res)
}

func (h *Handler) ValidateSetupStructure(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	valid, errs, err := h.usecase.ValidateSetupStructure(a, productID)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{
		"valid":             valid,
		"validation_errors": errs,
	})
}

func (h *Handler) CompleteSetup(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	if err := h.usecase.CompleteSetup(a, productID); err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{
		"product_id":   productID,
		"setup_status": "ACTIVE",
		"message":      "Product setup completed successfully",
	})
}

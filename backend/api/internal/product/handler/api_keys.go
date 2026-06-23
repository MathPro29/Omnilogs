package handler

import (
	"net/http"

	"omnilogs-api/dto"
	"omnilogs-api/utils"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateAPIKey(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	var req dto.CreateAPIKeyRequest
	if !bind(c, &req) {
		return
	}
	value, err := h.usecase.CreateAPIKey(a, productID, req)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, value)
}
func (h *Handler) ListAPIKeys(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	values, err := h.usecase.ListAPIKeys(a, productID)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, values)
}
func (h *Handler) UpdateAPIKey(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	id, ok := idParam(c, "keyId")
	if !ok {
		return
	}
	var req dto.UpdateAPIKeyRequest
	if !bind(c, &req) {
		return
	}
	value, err := h.usecase.UpdateAPIKey(a, productID, id, req)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, value)
}
func (h *Handler) RevokeAPIKey(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	id, ok := idParam(c, "keyId")
	if !ok {
		return
	}
	if err := h.usecase.RevokeAPIKey(a, productID, id); err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"key_id": id, "status": "revoked"})
}

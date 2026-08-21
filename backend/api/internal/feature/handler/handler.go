package handler

import (
	"net/http"

	"omnilogs-api/dto"
	"omnilogs-api/utils"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateFeature(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	projectID, ok := idParam(c, "projectId")
	if !ok {
		return
	}
	var req dto.CreateProjectFeatureRequest
	if !bind(c, &req) {
		return
	}
	value, err := h.usecase.CreateFeature(a, productID, projectID, req)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, value)
}
func (h *Handler) ListFeatures(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	projectID, ok := idParam(c, "projectId")
	if !ok {
		return
	}
	values, err := h.usecase.ListFeatures(a, productID, projectID)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, values)
}
func (h *Handler) UpdateFeature(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	projectID, ok := idParam(c, "projectId")
	if !ok {
		return
	}
	id, ok := idParam(c, "featureId")
	if !ok {
		return
	}
	var req dto.UpdateProjectFeatureRequest
	if !bind(c, &req) {
		return
	}
	value, err := h.usecase.UpdateFeature(a, productID, projectID, id, req)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, value)
}
func (h *Handler) DeleteFeature(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	projectID, ok := idParam(c, "projectId")
	if !ok {
		return
	}
	id, ok := idParam(c, "featureId")
	if !ok {
		return
	}
	if err := h.usecase.DeleteFeature(a, productID, projectID, id); err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"feature_id": id, "status": "deleted"})
}

package handler

import (
	"net/http"

	"omnilogs-api/dto"
	"omnilogs-api/utils"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateProject(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	var req dto.CreateProjectRequest
	if !bind(c, &req) {
		return
	}
	value, err := h.usecase.CreateProject(a, productID, req)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, value)
}
func (h *Handler) ListProjects(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	values, err := h.usecase.ListProjects(a, productID)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, values)
}
func (h *Handler) GetProject(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	id, ok := idParam(c, "projectId")
	if !ok {
		return
	}
	value, err := h.usecase.GetProject(a, productID, id)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, value)
}
func (h *Handler) UpdateProject(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	id, ok := idParam(c, "projectId")
	if !ok {
		return
	}
	var req dto.UpdateProjectRequest
	if !bind(c, &req) {
		return
	}
	value, err := h.usecase.UpdateProject(a, productID, id, req)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, value)
}
func (h *Handler) DeleteProject(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	projectID, ok := idParam(c, "projectId")
	if !ok {
		return
	}
	if err := h.usecase.DeleteProject(a, productID, projectID); err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"project_id": projectID, "status": "deleted"})
}

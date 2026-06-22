package handler

import (
	"errors"
	"net/http"
	"strconv"

	"omnilogs-api/dto"
	"omnilogs-api/internal/product/usecase"
	"omnilogs-api/middleware"
	"omnilogs-api/responses"
	"omnilogs-api/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct{ usecase usecase.Usecase }

func NewHandler(usecase usecase.Usecase) *Handler { return &Handler{usecase: usecase} }

func actor(c *gin.Context) (usecase.Actor, bool) {
	id, ok := middleware.CurrentUserID(c)
	return usecase.Actor{UserID: int(id), PlatformAdmin: middleware.HasAdminPlatformRole(c)}, ok
}
func idParam(c *gin.Context, name string) (int, bool) {
	value, err := strconv.Atoi(c.Param(name))
	if err != nil || value <= 0 {
		responses.BadRequest(c, "invalid "+name)
		return 0, false
	}
	return value, true
}
func bind(c *gin.Context, value any) bool {
	if err := c.ShouldBindJSON(value); err != nil {
		responses.ValidationError(c, err)
		return false
	}
	return true
}
func fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrNotFound):
		responses.NotFound(c, "product resource not found")
	case errors.Is(err, usecase.ErrForbidden):
		responses.Forbidden(c, "product permission denied")
	case errors.Is(err, usecase.ErrConflict):
		utils.Error(c, http.StatusConflict, "CONFLICT", "product resource already exists")
	case errors.Is(err, usecase.ErrInvalid):
		responses.BadRequest(c, "invalid product relationship or payload")
	default:
		responses.Error(c, "INTERNAL_ERROR", "internal server error", err)
	}
}

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

func (h *Handler) CreateEnvironment(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	var req dto.CreateEnvironmentRequest
	if !bind(c, &req) {
		return
	}
	value, err := h.usecase.CreateEnvironment(a, productID, req)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, value)
}
func (h *Handler) ListEnvironments(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	values, err := h.usecase.ListEnvironments(a, productID)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, values)
}
func (h *Handler) UpdateEnvironment(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	id, ok := idParam(c, "environmentId")
	if !ok {
		return
	}
	var req dto.UpdateEnvironmentRequest
	if !bind(c, &req) {
		return
	}
	value, err := h.usecase.UpdateEnvironment(a, productID, id, req)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, value)
}

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
func (h *Handler) CreateScope(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	membershipID, ok := idParam(c, "membershipId")
	if !ok {
		return
	}
	var req dto.CreateMembershipScopeRequest
	if !bind(c, &req) {
		return
	}
	value, err := h.usecase.CreateScope(a, productID, membershipID, req)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, value)
}
func (h *Handler) ListScopes(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	membershipID, ok := idParam(c, "membershipId")
	if !ok {
		return
	}
	values, err := h.usecase.ListScopes(a, productID, membershipID)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, values)
}
func (h *Handler) UpdateScope(c *gin.Context) {
	a, _ := actor(c)
	productID, ok := idParam(c, "productId")
	if !ok {
		return
	}
	membershipID, ok := idParam(c, "membershipId")
	if !ok {
		return
	}
	id, ok := idParam(c, "scopeId")
	if !ok {
		return
	}
	var req dto.UpdateMembershipScopeRequest
	if !bind(c, &req) {
		return
	}
	value, err := h.usecase.UpdateScope(a, productID, membershipID, id, req)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, value)
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

package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"omnilogs-api/dto"
	customfieldparser "omnilogs-api/internal/custom_fields"
	"omnilogs-api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)



func validateFavoriteReorder(items []dto.FavoriteOrder) error {
	seenIDs := make(map[int]struct{}, len(items))
	seenOrders := make(map[int]struct{}, len(items))
	for _, item := range items {
		if _, exists := seenIDs[item.FieldDefinitionID]; exists {
			return fmt.Errorf("duplicate favorite field %d", item.FieldDefinitionID)
		}
		if _, exists := seenOrders[item.DisplayOrder]; exists {
			return fmt.Errorf("duplicate display_order %d", item.DisplayOrder)
		}
		seenIDs[item.FieldDefinitionID] = struct{}{}
		seenOrders[item.DisplayOrder] = struct{}{}
	}
	return nil
}

func (h *Handler) ParseJSON(c *gin.Context) {
	var req dto.ParseJSONRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	nodes, err := customfieldparser.ParseJSONFields(req.SampleJSON)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"nodes":       nodes,
		"field_count": countJSONNodes(nodes),
		"max_nodes":   customfieldparser.MaxSampleJSONNodes,
	})
}

func countJSONNodes(nodes []customfieldparser.JSONFieldNode) int {
	count := 0
	for _, node := range nodes {
		count++
		count += countJSONNodes(node.Children)
	}
	return count
}

func (h *Handler) ListFavoriteJSONFields(c *gin.Context) {
	productID, err := strconv.Atoi(c.Query("product_id"))
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid product_id is required"})
		return
	}
	query := scopeFieldQuery(h.db.WithContext(c.Request.Context()).Model(&models.LogFieldDefinition{}), productID, optionalID(c.Query("project_id")), optionalID(c.Query("category_id")))
	var fields []models.LogFieldDefinition
	if err := query.Where("is_favorite = TRUE AND is_active = TRUE").Order("display_order ASC NULLS LAST, field_definition_id ASC").Find(&fields).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fields)
}

func (h *Handler) FavoriteJSONField(c *gin.Context) {
	var req dto.FavoriteJSONFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rawPath := strings.TrimSpace(req.FieldPath)
	path := normalizeFavoriteFieldPath(rawPath)
	dataType := strings.ToLower(strings.TrimSpace(req.DetectedType))
	if path == "" || dataType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "field_path and detected_type are required"})
		return
	}
	db := h.db.WithContext(c.Request.Context())

	var existing models.LogFieldDefinition
	pathQuery := scopeFieldQuery(db.Model(&models.LogFieldDefinition{}), req.ProductID, req.ProjectID, req.CategoryID)
	pathCandidates := []string{path}
	if rawPath != path {
		pathCandidates = append(pathCandidates, rawPath)
	}
	if err := pathQuery.Where("field_path IN ?", pathCandidates).First(&existing).Error; err == nil {
		if existing.IsFavorite && existing.IsActive {
			c.JSON(http.StatusConflict, gin.H{"error": "JSON Path is already favorited", "field_definition_id": existing.FieldDefinitionID})
			return
		}
		elasticFieldName := "payload." + path
		if path == "$" {
			elasticFieldName = "payload"
		}
		updates := map[string]any{
			"is_active": true, "is_favorite": true, "detected_type": dataType,
			"sample_path_found": true, "field_path": path, "elastic_field_name": elasticFieldName,
		}
		if len(req.SampleValue) > 0 {
			updates["sample_value"] = req.SampleValue
		}
		if req.DisplayName != nil {
			updates["display_name"] = req.DisplayName
		}
		if err := db.Model(&existing).Clauses(clause.Returning{}).Updates(updates).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, existing)
		return
	}

	baseFieldKey := customfieldparser.FieldKeyFromPath(path)
	if path == "$" {
		baseFieldKey = "payload"
	}
	var usedFieldKeys []string
	keyQuery := scopeFieldQuery(db.Model(&models.LogFieldDefinition{}), req.ProductID, req.ProjectID, req.CategoryID)
	if err := keyQuery.Pluck("field_key", &usedFieldKeys).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fieldKey := nextAvailableFieldKey(baseFieldKey, usedFieldKeys)
	displayName := req.DisplayName
	if displayName == nil {
		label := fieldKey
		displayName = &label
	}
	productID := req.ProductID
	maxOrder := -1
	if err := scopeFieldQuery(db.Model(&models.LogFieldDefinition{}), req.ProductID, req.ProjectID, req.CategoryID).
		Select("COALESCE(MAX(display_order), -1)").Scan(&maxOrder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	order := maxOrder + 1
	field := models.LogFieldDefinition{
		ProductID: &productID, ProjectID: req.ProjectID, CategoryID: req.CategoryID,
		FieldKey: fieldKey, DisplayName: displayName, SourceSection: "payload", FieldPath: &path,
		ElasticFieldName: favoriteElasticFieldName(path), DataType: dataType, ValueSourceType: "NONE",
		IsVisible: true, IsSearchable: true, IsFilterable: true, IsSortable: true, IsAggregatable: true,
		DisplayOrder: &order, IsActive: true, IsFavorite: true, DetectedType: &dataType, SamplePathFound: true,
	}
	if len(req.SampleValue) > 0 {
		sample := json.RawMessage(append([]byte(nil), req.SampleValue...))
		field.SampleValue = &sample
	}
	if err := db.Create(&field).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, field)
}

func (h *Handler) ReorderFavoriteJSONFields(c *gin.Context) {
	var req dto.FavoriteReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validateFavoriteReorder(req.Items); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := reorderFavoriteFields(c.Request.Context(), h.db, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) RemoveFavoriteJSONField(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	result := h.db.WithContext(c.Request.Context()).Model(&models.LogFieldDefinition{}).Where("field_definition_id = ? AND is_favorite = TRUE", id).Update("is_favorite", false)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "favorite JSON field not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

func nextAvailableFieldKey(base string, usedKeys []string) string {
	used := make(map[string]struct{}, len(usedKeys))
	for _, key := range usedKeys {
		used[key] = struct{}{}
	}
	if _, exists := used[base]; !exists {
		return base
	}
	for suffix := 2; ; suffix++ {
		candidate := base + "_" + strconv.Itoa(suffix)
		if _, exists := used[candidate]; !exists {
			return candidate
		}
	}
}

func reorderFavoriteFields(ctx context.Context, db *gorm.DB, req dto.FavoriteReorderRequest) error {
	var query strings.Builder
	query.WriteString("UPDATE log_field_definitions SET display_order = CASE field_definition_id")
	args := make([]any, 0, len(req.Items)*2+2)
	fieldIDs := make([]int, 0, len(req.Items))
	for _, item := range req.Items {
		query.WriteString(" WHEN ? THEN ?")
		args = append(args, item.FieldDefinitionID, item.DisplayOrder)
		fieldIDs = append(fieldIDs, item.FieldDefinitionID)
	}
	query.WriteString(" ELSE display_order END, updated_at = NOW() WHERE product_id = ? AND project_id IS NULL AND category_id IS NULL AND is_favorite = TRUE AND is_active = TRUE AND field_definition_id IN ?")
	args = append(args, req.ProductID, fieldIDs)

	result := db.WithContext(ctx).Exec(query.String(), args...)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != int64(len(req.Items)) {
		return fmt.Errorf("one or more favorite fields were not found")
	}
	return nil
}

func normalizeFavoriteFieldPath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.TrimPrefix(path, "raw.")
	for {
		old := path
		path = strings.TrimPrefix(path, "data.")
		path = strings.TrimPrefix(path, "fields.")
		path = strings.TrimPrefix(path, "payload.")
		if path == old {
			break
		}
	}
	if path == "payload" || path == "data" || path == "fields" || path == "" {
		return "$"
	}
	return path
}

func favoriteElasticFieldName(path string) string {
	if path == "$" {
		return "payload"
	}
	return "payload." + path
}

func dynamicElasticFieldName(sourceSection, path string) string {
	path = strings.TrimSpace(path)
	path = strings.TrimPrefix(path, "raw.")
	if !strings.HasPrefix(path, sourceSection+".") && path != sourceSection {
		path = sourceSection + "." + path
	}
	return path
}

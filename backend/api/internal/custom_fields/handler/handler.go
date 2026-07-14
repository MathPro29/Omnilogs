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
	"omnilogs-api/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Handler struct{ db *gorm.DB }

type customFieldResponse struct {
	models.LogFieldDefinition
	EnumOptions []models.LogFieldEnumOption `json:"enum_options"`
}

type parseJSONRequest struct {
	SampleJSON json.RawMessage `json:"sample_json" binding:"required"`
}

type favoriteJSONFieldRequest struct {
	ProductID    int             `json:"product_id" binding:"required,gt=0"`
	ProjectID    *int            `json:"project_id,omitempty"`
	CategoryID   *int            `json:"category_id,omitempty"`
	FieldPath    string          `json:"field_path" binding:"required"`
	DisplayName  *string         `json:"display_name,omitempty"`
	SampleValue  json.RawMessage `json:"sample_value,omitempty"`
	DetectedType string          `json:"detected_type" binding:"required"`
}

type BulkDeleteCustomFieldRequest struct {
	FieldDefinitionIDs []int `json:"field_definition_ids" binding:"required,min=1"`
}

type favoriteReorderRequest struct {
	ProductID int             `json:"product_id" binding:"required,gt=0"`
	Items     []favoriteOrder `json:"items" binding:"required,min=1"`
}

type favoriteOrder struct {
	FieldDefinitionID int `json:"field_definition_id" binding:"required,gt=0"`
	DisplayOrder      int `json:"display_order" binding:"gte=0"`
}

func validateFavoriteReorder(items []favoriteOrder) error {
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
	var req parseJSONRequest
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

func scopeFieldQuery(query *gorm.DB, productID int, projectID, categoryID *int) *gorm.DB {
	query = query.Where("product_id = ?", productID)
	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	} else {
		query = query.Where("project_id IS NULL")
	}
	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	} else {
		query = query.Where("category_id IS NULL")
	}
	return query
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
	var req favoriteJSONFieldRequest
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
	var req favoriteReorderRequest
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

// handler bulk delete
func (h *Handler) BulkDeleteCustomFields(c *gin.Context) {
	var req BulkDeleteCustomFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result := h.db.WithContext(c.Request.Context()).Model(&models.LogFieldDefinition{}).Where("field_definition_id IN ?", req.FieldDefinitionIDs).Update("is_active", false)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no custom fields found or deleted"})
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"field_definition_ids": req.FieldDefinitionIDs, "status": "deleted"})
}

func optionalID(value string) *int {
	if value == "" {
		return nil
	}
	id, err := strconv.Atoi(value)
	if err != nil || id <= 0 {
		return nil
	}
	return &id
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

func reorderFavoriteFields(ctx context.Context, db *gorm.DB, req favoriteReorderRequest) error {
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
	if path == "payload" {
		return "$"
	}
	return strings.TrimPrefix(path, "payload.")
}

func favoriteElasticFieldName(path string) string {
	if path == "$" {
		return "payload"
	}
	return "payload." + path
}

func NewHandler(db *gorm.DB, _ any) *Handler { return &Handler{db: db} }

func (h *Handler) GetAllCustomFields(c *gin.Context) {
	db := h.db.WithContext(c.Request.Context())
	query := db.Model(&models.LogFieldDefinition{}).Order("display_order ASC NULLS LAST, field_definition_id ASC")
	if value := c.Query("product_id"); value != "" {
		id, err := strconv.Atoi(value)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product_id"})
			return
		}
		query = query.Where("product_id = ? OR product_id IS NULL", id)
	}
	if value := c.Query("project_id"); value != "" {
		id, err := strconv.Atoi(value)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
			return
		}
		query = query.Where("project_id = ? OR project_id IS NULL", id)
	}
	if value := c.Query("category_id"); value != "" {
		id, err := strconv.Atoi(value)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
			return
		}
		query = query.Where("category_id = ? OR category_id IS NULL", id)
	}
	if c.Query("active_only") == "true" {
		query = query.Where("is_active = TRUE")
	}
	var fields []models.LogFieldDefinition
	if err := query.Find(&fields).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fieldIDs := make([]int, 0, len(fields))
	for _, field := range fields {
		if strings.EqualFold(field.DataType, "enum") {
			fieldIDs = append(fieldIDs, field.FieldDefinitionID)
		}
	}
	optionsByField := make(map[int][]models.LogFieldEnumOption)
	if len(fieldIDs) > 0 {
		var options []models.LogFieldEnumOption
		if err := db.Where("field_definition_id IN ? AND is_active = TRUE", fieldIDs).
			Order("display_order ASC NULLS LAST, option_id ASC").Find(&options).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, option := range options {
			optionsByField[option.FieldDefinitionID] = append(optionsByField[option.FieldDefinitionID], option)
		}
	}
	response := make([]customFieldResponse, 0, len(fields))
	for _, field := range fields {
		options := optionsByField[field.FieldDefinitionID]
		if options == nil {
			options = []models.LogFieldEnumOption{}
		}
		response = append(response, customFieldResponse{LogFieldDefinition: field, EnumOptions: options})
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetCustomFieldByID(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	db := h.db.WithContext(c.Request.Context())
	var field models.LogFieldDefinition
	if err := db.First(&field, "field_definition_id = ?", id).Error; err != nil {
		status := http.StatusInternalServerError
		if err == gorm.ErrRecordNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	var options []models.LogFieldEnumOption
	if err := db.Where("field_definition_id = ?", id).Order("display_order ASC NULLS LAST, option_id ASC").Find(&options).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var sources []models.LogFieldValueSource
	if err := db.Where("field_definition_id = ?", id).Order("value_source_id ASC").Find(&sources).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"field": field, "options": options, "value_sources": sources})
}

func (h *Handler) CreateCustomField(c *gin.Context) {
	var req dto.CreateLogFieldDefinitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var configJSON *json.RawMessage
	if req.ConfigJSON != nil && string(req.ConfigJSON) != "null" {
		configJSON = &req.ConfigJSON
	}

	// Check for duplicate key/scope to prevent SQL unique constraint violations
	dupQuery := h.db.Model(&models.LogFieldDefinition{}).Where("field_key = ?", req.FieldKey)
	if req.ProductID != nil {
		dupQuery = dupQuery.Where("product_id = ?", *req.ProductID)
	} else {
		dupQuery = dupQuery.Where("product_id IS NULL")
	}
	if req.ProjectID != nil {
		dupQuery = dupQuery.Where("project_id = ?", *req.ProjectID)
	} else {
		dupQuery = dupQuery.Where("project_id IS NULL")
	}
	if req.CategoryID != nil {
		dupQuery = dupQuery.Where("category_id = ?", *req.CategoryID)
	} else {
		dupQuery = dupQuery.Where("category_id IS NULL")
	}

	var existing models.LogFieldDefinition
	if err := dupQuery.First(&existing).Error; err == nil {
		if existing.IsActive {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ฟิลด์คีย์นี้ถูกใช้งานไปแล้วในขอบเขต (Product/Project/Category) นี้"})
			return
		}

		// Reactivate the inactive record
		existing.IsActive = true
		existing.DisplayName = req.DisplayName
		existing.Description = req.Description
		existing.DataType = req.DataType
		existing.SourceSection = req.SourceSection
		existing.FieldPath = req.FieldPath
		existing.ElasticFieldName = req.ElasticFieldName
		existing.ValueSourceType = req.ValueSourceType
		existing.ValueSourceKey = req.ValueSourceKey
		existing.IsRequired = req.IsRequired
		existing.IsSensitive = req.IsSensitive
		existing.MaskBeforeIndex = req.MaskBeforeIndex
		existing.EncryptBeforeArchive = req.EncryptBeforeArchive
		if req.IsVisible != nil {
			existing.IsVisible = *req.IsVisible
		} else {
			existing.IsVisible = true
		}
		existing.IsSearchable = req.IsSearchable
		existing.IsFilterable = req.IsFilterable
		existing.IsSortable = req.IsSortable
		existing.IsAggregatable = req.IsAggregatable
		existing.DisplayOrder = req.DisplayOrder
		existing.DefaultValue = req.DefaultValue
		existing.FieldType = req.FieldType
		existing.ConfigJSON = configJSON
		existing.SchemaVersion = 1

		if err := h.db.Save(&existing).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, existing)
		return
	}
	field := models.LogFieldDefinition{
		ProductID: req.ProductID, ProjectID: req.ProjectID, CategoryID: req.CategoryID,
		FieldKey: req.FieldKey, DisplayName: req.DisplayName, Description: req.Description,
		SourceSection: req.SourceSection, FieldPath: req.FieldPath, ElasticFieldName: req.ElasticFieldName,
		DataType: req.DataType, ValueSourceType: req.ValueSourceType, ValueSourceKey: req.ValueSourceKey,
		IsRequired: req.IsRequired, IsSensitive: req.IsSensitive, MaskBeforeIndex: req.MaskBeforeIndex,
		EncryptBeforeArchive: req.EncryptBeforeArchive, IsVisible: true, IsSearchable: req.IsSearchable,
		IsFilterable: req.IsFilterable, IsSortable: req.IsSortable, IsAggregatable: req.IsAggregatable,
		DisplayOrder: req.DisplayOrder, DefaultValue: req.DefaultValue, IsActive: true,
		FieldType: req.FieldType, ConfigJSON: configJSON, SchemaVersion: 1,
	}
	if req.IsVisible != nil {
		field.IsVisible = *req.IsVisible
	}
	if err := h.db.Create(&field).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, field)
}

func (h *Handler) UpdateCustomField(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var req dto.UpdateLogFieldDefinitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]any{}
	if req.DisplayName != nil {
		updates["display_name"] = req.DisplayName
	}
	if req.Description != nil {
		updates["description"] = req.Description
	}
	if req.FieldPath != nil {
		updates["field_path"] = req.FieldPath
	}
	if req.ElasticFieldName != nil {
		updates["elastic_field_name"] = req.ElasticFieldName
	}
	if req.DataType != nil {
		updates["data_type"] = req.DataType
	}
	if req.ValueSourceType != nil {
		updates["value_source_type"] = req.ValueSourceType
	}
	if req.ValueSourceKey != nil {
		updates["value_source_key"] = req.ValueSourceKey
	}
	if req.IsRequired != nil {
		updates["is_required"] = *req.IsRequired
	}
	if req.IsSensitive != nil {
		updates["is_sensitive"] = *req.IsSensitive
	}
	if req.MaskBeforeIndex != nil {
		updates["mask_before_index"] = *req.MaskBeforeIndex
	}
	if req.EncryptBeforeArchive != nil {
		updates["encrypt_before_archive"] = *req.EncryptBeforeArchive
	}
	if req.IsVisible != nil {
		updates["is_visible"] = *req.IsVisible
	}
	if req.IsSearchable != nil {
		updates["is_searchable"] = *req.IsSearchable
	}
	if req.IsFilterable != nil {
		updates["is_filterable"] = *req.IsFilterable
	}
	if req.IsSortable != nil {
		updates["is_sortable"] = *req.IsSortable
	}
	if req.IsAggregatable != nil {
		updates["is_aggregatable"] = *req.IsAggregatable
	}
	if req.DisplayOrder != nil {
		updates["display_order"] = *req.DisplayOrder
	}
	if req.DefaultValue != nil {
		updates["default_value"] = req.DefaultValue
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.SamplePathFound != nil {
		updates["sample_path_found"] = *req.SamplePathFound
	}
	if req.FieldType != nil {
		updates["field_type"] = req.FieldType
	}
	if req.ConfigJSON != nil {
		if string(req.ConfigJSON) == "null" {
			updates["config_json"] = nil
		} else {
			updates["config_json"] = req.ConfigJSON
		}
		// Auto-increment schema version when config changes
		updates["schema_version"] = gorm.Expr("schema_version + 1")
	}
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}
	result := h.db.Model(&models.LogFieldDefinition{}).Where("field_definition_id = ?", id).Updates(updates)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "custom field not found"})
		return
	}
	h.GetCustomFieldByID(c)
}

func (h *Handler) DeleteCustomField(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	result := h.db.Model(&models.LogFieldDefinition{}).Where("field_definition_id = ?", id).Update("is_active", false)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "custom field not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) CreateOption(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var req dto.CreateLogFieldEnumOptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var field models.LogFieldDefinition
	if err := h.db.First(&field, "field_definition_id = ? AND is_active = TRUE", id).Error; err != nil {
		status := http.StatusBadRequest
		if err == gorm.ErrRecordNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": "active custom field not found"})
		return
	}
	if !strings.EqualFold(field.DataType, "enum") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "options can only be added to enum fields"})
		return
	}
	option := models.LogFieldEnumOption{
		FieldDefinitionID: id,
		OptionKey:         strings.TrimSpace(req.OptionKey), OptionLabel: strings.TrimSpace(req.OptionLabel),
		OptionValue: strings.TrimSpace(req.OptionValue), DisplayOrder: req.DisplayOrder,
		ColorCode: req.ColorCode, Description: req.Description,
		IsDefault: req.IsDefault, IsActive: true,
	}
	if option.OptionKey == "" || option.OptionLabel == "" || option.OptionValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "option_key, option_label and option_value are required"})
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if option.IsDefault {
			if err := tx.Model(&models.LogFieldEnumOption{}).Where("field_definition_id = ?", id).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Create(&option).Error
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, option)
}

func (h *Handler) ListOptions(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	query := h.db.Where("field_definition_id = ?", id)
	if c.Query("active_only") == "true" {
		query = query.Where("is_active = TRUE")
	}
	var options []models.LogFieldEnumOption
	if err := query.Order("display_order ASC NULLS LAST, option_id ASC").Find(&options).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, options)
}

func (h *Handler) UpdateOption(c *gin.Context) {
	fieldID, ok := pathID(c)
	if !ok {
		return
	}
	optionID, ok := subID(c, "optionId")
	if !ok {
		return
	}
	var req dto.UpdateLogFieldEnumOptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]any{}
	if req.OptionKey != nil {
		updates["option_key"] = strings.TrimSpace(*req.OptionKey)
	}
	if req.OptionLabel != nil {
		updates["option_label"] = strings.TrimSpace(*req.OptionLabel)
	}
	if req.OptionValue != nil {
		updates["option_value"] = strings.TrimSpace(*req.OptionValue)
	}
	if req.DisplayOrder != nil {
		updates["display_order"] = *req.DisplayOrder
	}
	if req.ColorCode != nil {
		updates["color_code"] = *req.ColorCode
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.IsDefault != nil {
		updates["is_default"] = *req.IsDefault
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if req.IsDefault != nil && *req.IsDefault {
			if err := tx.Model(&models.LogFieldEnumOption{}).Where("field_definition_id = ?", fieldID).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		result := tx.Model(&models.LogFieldEnumOption{}).Where("option_id = ? AND field_definition_id = ?", optionID, fieldID).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "option not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var option models.LogFieldEnumOption
	h.db.First(&option, optionID)
	c.JSON(http.StatusOK, option)
}

func (h *Handler) DeleteOption(c *gin.Context) {
	fieldID, ok := pathID(c)
	if !ok {
		return
	}
	optionID, ok := subID(c, "optionId")
	if !ok {
		return
	}
	result := h.db.Model(&models.LogFieldEnumOption{}).
		Where("option_id = ? AND field_definition_id = ?", optionID, fieldID).
		Updates(map[string]any{"is_active": false, "is_default": false})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "option not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) CreateValueSource(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var req dto.CreateLogFieldValueSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	source := models.LogFieldValueSource{FieldDefinitionID: id, SourceType: req.SourceType, SourceKey: req.SourceKey, SourceConfig: req.SourceConfig, LabelField: req.LabelField, ValueField: req.ValueField, IsRequired: req.IsRequired, TimeoutSeconds: req.TimeoutSeconds, RefreshMode: req.RefreshMode, IsActive: true}
	if err := h.db.Create(&source).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, source)
}

func (h *Handler) ListValueSources(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var sources []models.LogFieldValueSource
	if err := h.db.Where("field_definition_id = ?", id).Order("value_source_id ASC").Find(&sources).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sources)
}

func (h *Handler) UpdateValueSource(c *gin.Context) {
	fieldID, ok := pathID(c)
	if !ok {
		return
	}
	sourceID, ok := subID(c, "sourceId")
	if !ok {
		return
	}
	var req dto.UpdateLogFieldValueSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]any{}
	if req.SourceType != nil {
		updates["source_type"] = req.SourceType
	}
	if req.SourceKey != nil {
		updates["source_key"] = req.SourceKey
	}
	if req.SourceConfig != nil {
		updates["source_config"] = req.SourceConfig
	}
	if req.LabelField != nil {
		updates["label_field"] = req.LabelField
	}
	if req.ValueField != nil {
		updates["value_field"] = req.ValueField
	}
	if req.IsRequired != nil {
		updates["is_required"] = *req.IsRequired
	}
	if req.TimeoutSeconds != nil {
		updates["timeout_seconds"] = *req.TimeoutSeconds
	}
	if req.RefreshMode != nil {
		updates["refresh_mode"] = req.RefreshMode
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	result := h.db.Model(&models.LogFieldValueSource{}).Where("value_source_id = ? AND field_definition_id = ?", sourceID, fieldID).Updates(updates)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "value source not found"})
		return
	}
	var source models.LogFieldValueSource
	h.db.First(&source, sourceID)
	c.JSON(http.StatusOK, source)
}

func pathID(c *gin.Context) (int, bool) { return subID(c, "id") }
func subID(c *gin.Context, name string) (int, bool) {
	id, err := strconv.Atoi(c.Param(name))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return id, true
}

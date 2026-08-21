package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"omnilogs-api/dto"
	"omnilogs-api/models"
	"omnilogs-api/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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
	response := make([]dto.CustomFieldResponse, 0, len(fields))
	for _, field := range fields {
		options := optionsByField[field.FieldDefinitionID]
		if options == nil {
			options = []models.LogFieldEnumOption{}
		}
		response = append(response, dto.CustomFieldResponse{LogFieldDefinition: field, EnumOptions: options})
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

func (h *Handler) BulkDeleteCustomFields(c *gin.Context) {
	var req dto.BulkDeleteCustomFieldRequest
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

package handler

import (
	"net/http"

	"omnilogs-api/dto"
	"omnilogs-api/models"

	"github.com/gin-gonic/gin"
)

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

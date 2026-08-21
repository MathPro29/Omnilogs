package handler

import (
	"net/http"
	"strings"

	"omnilogs-api/dto"
	"omnilogs-api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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

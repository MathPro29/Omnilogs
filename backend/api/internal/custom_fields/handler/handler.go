package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct{ db *gorm.DB }

func NewHandler(db *gorm.DB) *Handler { return &Handler{db: db} }

func pathID(c *gin.Context) (int, bool) { return subID(c, "id") }

func subID(c *gin.Context, name string) (int, bool) {
	id, err := strconv.Atoi(c.Param(name))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return id, true
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

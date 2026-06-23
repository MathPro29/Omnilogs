package usecase

import (
	"omnilogs-api/models"

	"gorm.io/gorm"
)

func validScope(db *gorm.DB, productID int, level string, projectID, categoryID *int) bool {
	switch level {
	case "PRODUCT":
		return projectID == nil && categoryID == nil
	case "PROJECT":
		return projectID != nil && categoryID == nil && existsDB(db, &models.Project{}, "project_id = ? AND product_id = ?", *projectID, productID)
	case "CATEGORY":
		if projectID == nil || categoryID == nil {
			return false
		}
		return existsDB(db, &models.ProjectFeature{}, "category_id = ? AND project_id = ? AND product_id = ?", *categoryID, *projectID, productID)
	default:
		return false
	}
}

package usecase

import (
	"fmt"
	"strings"

	"omnilogs-api/models"
	"omnilogs-api/responses"

	"gorm.io/gorm"
)

func (u *usecase) getFeature(productID, projectID, id int) (*models.ProjectFeature, error) {
	var value models.ProjectFeature
	if err := u.repository.DB().Where("category_id = ? AND product_id = ? AND project_id = ?", id, productID, projectID).First(&value).Error; err != nil {
		return nil, classifyDBError(err)
	}
	return &value, nil
}

func rebuildFeaturePaths(db *gorm.DB, projectID int) error {
	var values []models.ProjectFeature
	if err := db.Where("project_id = ?", projectID).Find(&values).Error; err != nil {
		return err
	}
	byID := map[int]*models.ProjectFeature{}
	for i := range values {
		byID[values[i].CategoryID] = &values[i]
	}
	state := map[int]int{}
	var visit func(*models.ProjectFeature) error
	visit = func(value *models.ProjectFeature) error {
		if state[value.CategoryID] == 2 {
			return nil
		}
		if state[value.CategoryID] == 1 {
			return responses.ErrInvalid
		}
		state[value.CategoryID] = 1
		if value.ParentID == nil {
			path, ids := value.CategoryCode, fmt.Sprint(value.CategoryID)
			value.FullPath, value.PathIDs, value.Level = &path, &ids, 1
		} else {
			parent := byID[*value.ParentID]
			if parent == nil {
				return responses.ErrInvalid
			}
			if err := visit(parent); err != nil {
				return err
			}
			path, ids := *parent.FullPath+"/"+value.CategoryCode, *parent.PathIDs+","+fmt.Sprint(value.CategoryID)
			value.FullPath, value.PathIDs, value.Level = &path, &ids, parent.Level+1
		}
		state[value.CategoryID] = 2
		return db.Model(&models.ProjectFeature{}).Where("category_id = ?", value.CategoryID).Updates(map[string]any{"full_path": value.FullPath, "path_ids": value.PathIDs, "level": value.Level}).Error
	}
	for i := range values {
		if err := visit(&values[i]); err != nil {
			return err
		}
	}
	return nil
}

func pathContains(path *string, id int) bool {
	if path == nil {
		return false
	}
	needle := "," + fmt.Sprint(id) + ","
	return strings.Contains(","+*path+",", needle)
}

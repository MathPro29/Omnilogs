package usecase

import (
	"strings"
	"time"

	"omnilogs-api/dto"
	"omnilogs-api/models"
	"omnilogs-api/responses"

	"gorm.io/gorm"
)

func (u *usecase) CreateFeature(actor Actor, productID, projectID int, req dto.CreateProjectFeatureRequest) (*models.ProjectFeature, error) {
	target := AccessTarget{ProductID: productID, ProjectID: &projectID}
	if err := u.authorize(actor, target, "FEATURE", "CREATE"); err != nil {
		return nil, err
	}
	if (req.ProductID != 0 && req.ProductID != productID) || (req.ProjectID != 0 && req.ProjectID != projectID) || !u.exists(&models.Project{}, "project_id = ? AND product_id = ?", projectID, productID) {
		return nil, responses.ErrInvalid
	}
	value := &models.ProjectFeature{ProductID: productID, ProjectID: projectID, ParentID: req.ParentID, CategoryType: req.CategoryType, CategoryCode: normalizeCode(req.CategoryCode), CategoryName: strings.TrimSpace(req.CategoryName), IsActive: true}
	if value.CategoryCode == "" || value.CategoryName == "" {
		return nil, responses.ErrInvalid
	}
	err := u.repository.Transaction(func(tx *gorm.DB) error {
		if value.ParentID != nil && !existsDB(tx, &models.ProjectFeature{}, "category_id = ? AND product_id = ? AND project_id = ?", *value.ParentID, productID, projectID) {
			return responses.ErrInvalid
		}
		if err := tx.Create(value).Error; err != nil {
			return classifyDBError(err)
		}
		return rebuildFeaturePaths(tx, projectID)
	})
	if err != nil {
		return nil, err
	}
	return u.getFeature(productID, projectID, value.CategoryID)
}

func (u *usecase) ListFeatures(actor Actor, productID, projectID int) ([]models.ProjectFeature, error) {
	target := AccessTarget{ProductID: productID, ProjectID: &projectID}
	if err := u.authorize(actor, target, "FEATURE", "READ"); err != nil {
		return nil, err
	}
	var values []models.ProjectFeature
	q := u.repository.DB().Where("product_id = ? AND project_id = ?", productID, projectID)
	if !actor.PlatformAdmin {
		q = q.Where(`EXISTS (SELECT 1 FROM product_memberships pm JOIN product_membership_scopes pms ON pms.membership_id = pm.membership_id WHERE pm.user_id = ? AND pm.product_id = project_features.product_id AND pm.is_active = TRUE AND pms.is_active = TRUE AND (pm.expires_at IS NULL OR pm.expires_at > ?) AND (pms.scope_level = 'PRODUCT' OR (pms.scope_level = 'PROJECT' AND pms.project_id = project_features.project_id) OR (pms.scope_level = 'CATEGORY' AND pms.project_id = project_features.project_id AND (',' || project_features.path_ids || ',') LIKE ('%,' || pms.category_id::text || ',%'))))`, actor.UserID, time.Now())
	}
	return values, q.Order("level, category_id").Find(&values).Error
}

func (u *usecase) UpdateFeature(actor Actor, productID, projectID, id int, req dto.UpdateProjectFeatureRequest) (*models.ProjectFeature, error) {
	target := AccessTarget{ProductID: productID, ProjectID: &projectID, CategoryID: &id}
	if err := u.authorize(actor, target, "FEATURE", "UPDATE"); err != nil {
		return nil, err
	}
	err := u.repository.Transaction(func(tx *gorm.DB) error {
		var value models.ProjectFeature
		if err := tx.Where("category_id = ? AND product_id = ? AND project_id = ?", id, productID, projectID).First(&value).Error; err != nil {
			return classifyDBError(err)
		}
		if req.ParentID != nil {
			if *req.ParentID == id || !existsDB(tx, &models.ProjectFeature{}, "category_id = ? AND product_id = ? AND project_id = ?", *req.ParentID, productID, projectID) {
				return responses.ErrInvalid
			}
			value.ParentID = req.ParentID
		}
		if req.CategoryType != nil {
			value.CategoryType = req.CategoryType
		}
		if req.CategoryName != nil {
			value.CategoryName = strings.TrimSpace(*req.CategoryName)
			if value.CategoryName == "" {
				return responses.ErrInvalid
			}
		}
		if req.IsActive != nil {
			value.IsActive = *req.IsActive
		}
		if err := tx.Save(&value).Error; err != nil {
			return classifyDBError(err)
		}
		return rebuildFeaturePaths(tx, projectID)
	})
	if err != nil {
		return nil, err
	}
	return u.getFeature(productID, projectID, id)
}

func (u *usecase) DeleteFeature(actor Actor, productID, projectID, id int) error {
	target := AccessTarget{ProductID: productID, ProjectID: &projectID, CategoryID: &id}
	if err := u.authorize(actor, target, "FEATURE", "DELETE"); err != nil {
		return err
	}
	err := u.repository.Transaction(func(tx *gorm.DB) error {
		var value models.ProjectFeature
		if err := tx.Where("category_id = ? AND product_id = ? AND project_id = ?", id, productID, projectID).First(&value).Error; err != nil {
			return classifyDBError(err)
		}
		var childCount int64
		if err := tx.Model(&models.ProjectFeature{}).Where("parent_id = ? AND product_id = ? AND project_id = ?", id, productID, projectID).Count(&childCount).Error; err != nil {
			return err
		}
		if childCount > 0 {
			return responses.ErrInvalid
		}
		if err := tx.Delete(&models.ProjectFeature{}, "category_id = ? AND product_id = ? AND project_id = ?", id, productID, projectID).Error; err != nil {
			return classifyDBError(err)
		}
		return rebuildFeaturePaths(tx, projectID)
	})
	if err != nil {
		return err
	}
	return nil
}


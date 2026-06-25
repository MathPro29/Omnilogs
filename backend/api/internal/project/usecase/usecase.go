package usecase

import (
	"strings"
	"time"

	"omnilogs-api/dto"
	"omnilogs-api/models"
)

func (u *usecase) CreateProject(actor Actor, productID int, req dto.CreateProjectRequest) (*models.Project, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "PROJECT", "CREATE"); err != nil {
		return nil, err
	}
	if req.ProductID != 0 && req.ProductID != productID {
		return nil, ErrInvalid
	}
	value := &models.Project{ProductID: productID, ProjectCode: normalizeCode(req.ProjectCode), ProjectName: strings.TrimSpace(req.ProjectName), IsActive: true}
	if value.ProjectCode == "" || value.ProjectName == "" {
		return nil, ErrInvalid
	}
	if err := u.repository.DB().Create(value).Error; err != nil {
		return nil, classifyDBError(err)
	}
	return value, nil
}

func (u *usecase) ListProjects(actor Actor, productID int) ([]models.Project, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "PROJECT", "READ"); err != nil {
		return nil, err
	}
	var values []models.Project
	q := u.repository.DB().Where("product_id = ?", productID)
	if !actor.PlatformAdmin {
		q = q.Where(`EXISTS (SELECT 1 FROM product_memberships pm JOIN product_membership_scopes pms ON pms.membership_id = pm.membership_id WHERE pm.user_id = ? AND pm.product_id = projects.product_id AND pm.is_active = TRUE AND pms.is_active = TRUE AND (pm.expires_at IS NULL OR pm.expires_at > ?) AND (pms.scope_level = 'PRODUCT' OR (pms.scope_level IN ('PROJECT', 'CATEGORY') AND pms.project_id = projects.project_id)))`, actor.UserID, time.Now())
	}
	return values, q.Order("project_id").Find(&values).Error
}

func (u *usecase) GetProject(actor Actor, productID, id int) (*models.Project, error) {
	target := AccessTarget{ProductID: productID, ProjectID: &id}
	if err := u.authorize(actor, target, "PROJECT", "READ"); err != nil {
		return nil, err
	}
	var value models.Project
	if err := u.repository.DB().Where("project_id = ? AND product_id = ?", id, productID).First(&value).Error; err != nil {
		return nil, classifyDBError(err)
	}
	return &value, nil
}

func (u *usecase) UpdateProject(actor Actor, productID, id int, req dto.UpdateProjectRequest) (*models.Project, error) {
	target := AccessTarget{ProductID: productID, ProjectID: &id}
	if err := u.authorize(actor, target, "PROJECT", "UPDATE"); err != nil {
		return nil, err
	}
	updates := map[string]any{}
	if req.ProjectName != nil {
		name := strings.TrimSpace(*req.ProjectName)
		if name == "" {
			return nil, ErrInvalid
		}
		updates["project_name"] = name
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	res := u.repository.DB().Model(&models.Project{}).Where("project_id = ? AND product_id = ?", id, productID).Updates(updates)
	if res.Error != nil {
		return nil, classifyDBError(res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return u.GetProject(actor, productID, id)
}

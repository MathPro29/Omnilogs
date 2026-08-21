package dto

type CreateProjectRequest struct {
	ProductID   int    `json:"product_id,omitempty" binding:"omitempty,gt=0"`
	ProjectCode string `json:"project_code,omitempty" binding:"omitempty"`
	ProjectName string `json:"project_name" binding:"required"`
}
type UpdateProjectRequest struct {
	ProjectName *string `json:"project_name,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}
type ProjectResponse struct {
	ProjectID   int    `json:"project_id"`
	ProductID   int    `json:"product_id"`
	ProjectCode string `json:"project_code"`
	ProjectName string `json:"project_name"`
	IsActive    bool   `json:"is_active"`
	TimestampResponse
}

type CreateProjectFeatureRequest struct {
	ProductID    int     `json:"product_id,omitempty" binding:"omitempty,gt=0"`
	ProjectID    int     `json:"project_id,omitempty" binding:"omitempty,gt=0"`
	ParentID     *int    `json:"parent_id,omitempty" binding:"omitempty,gt=0"`
	CategoryType *string `json:"category_type,omitempty"`
	CategoryCode string  `json:"category_code,omitempty" binding:"omitempty"`
	CategoryName string  `json:"category_name" binding:"required"`
}
type UpdateProjectFeatureRequest struct {
	CategoryCode *string `json:"category_code,omitempty"`
	ParentID     *int    `json:"parent_id,omitempty"`
	CategoryType *string `json:"category_type,omitempty"`
	CategoryName *string `json:"category_name,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
}

type DeleteProjectFeatureRequest struct {
	ProductID  int `json:"product_id" binding:"required,gt=0"`
	ProjectID  int `json:"project_id" binding:"required,gt=0"`
	CategoryID int `json:"category_id" binding:"required,gt=0"`
}

type ProjectFeatureResponse struct {
	CategoryID   int     `json:"category_id"`
	ProductID    int     `json:"product_id"`
	ProjectID    int     `json:"project_id"`
	ParentID     *int    `json:"parent_id,omitempty"`
	CategoryType *string `json:"category_type,omitempty"`
	CategoryCode string  `json:"category_code"`
	CategoryName string  `json:"category_name"`
	FullPath     *string `json:"full_path,omitempty"`
	PathIDs      *string `json:"path_ids,omitempty"`
	Level        int     `json:"level"`
	IsActive     bool    `json:"is_active"`
	TimestampResponse
}

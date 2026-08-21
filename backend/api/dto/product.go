package dto

type CreateProductRequest struct {
	ProductName  string                     `json:"product_name" binding:"required"`
	ProductCode  string                     `json:"product_code,omitempty" binding:"omitempty"`
	Description  string                     `json:"description,omitempty"`
	Environments []CreateEnvironmentRequest `json:"environments,omitempty" binding:"omitempty"`
	OwnerID      *int                       `json:"owner_id,omitempty" binding:"omitempty"`
}

type UpdateProductRequest struct {
	ProductName *string `json:"product_name,omitempty"`
	ProductCode *string `json:"product_code,omitempty"`
	Description *string `json:"description,omitempty"`
	SetupStatus *string `json:"setup_status,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

type DeleteProductRequest struct {
	ProductID int `json:"product_id" binding:"required"`
}

// Bulk delete Product
type BulkDeleteProductRequest struct {
	ProductIDs []int `json:"product_ids" binding:"required"`
}

type ProductResponse struct {
	ProductID           int                   `json:"product_id"`
	ProductName         string                `json:"product_name"`
	ProductCode         string                `json:"product_code"`
	Description         string                `json:"description,omitempty"`
	SetupStatus         string                `json:"setup_status,omitempty"`
	ProductEnvironments []EnvironmentResponse `json:"product_environments,omitempty"`
	IsActive            bool                  `json:"is_active"`
	TimestampResponse
}

type SetupStatusResponse struct {
	ProductID        int      `json:"product_id"`
	ProductName      string   `json:"product_name"`
	ProductCode      string   `json:"product_code"`
	Description      string   `json:"description,omitempty"`
	SetupStatus      string   `json:"setup_status"`
	CurrentStep      int      `json:"current_step"`
	ProgressPercent  int      `json:"progress_percent"`
	Environments     []any    `json:"environments"`
	ProjectsCount    int      `json:"projects_count"`
	FeaturesCount    int      `json:"features_count"`
	SubFeaturesCount int      `json:"sub_features_count"`
	TotalNodesCount  int      `json:"total_nodes_count"`
	ApiKeysCount     int      `json:"api_keys_count"`
	TestLogReceived  bool     `json:"test_log_received"`
	ValidationErrors []string `json:"validation_errors,omitempty"`
}

type ReorderProjectsRequest struct {
	ProjectIDs []int `json:"project_ids" binding:"required"`
}

type ReorderFeaturesRequest struct {
	CategoryIDs []int `json:"category_ids" binding:"required"`
}

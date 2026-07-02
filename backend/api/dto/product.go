package dto

type CreateProductRequest struct {
	ProductName  string                     `json:"product_name" binding:"required"`
	ProductCode  string                     `json:"product_code,omitempty" binding:"omitempty"`
	Environments []CreateEnvironmentRequest `json:"environments,omitempty" binding:"omitempty"`
}
type UpdateProductRequest struct {
	ProductName *string `json:"product_name,omitempty"`
	ProductCode *string `json:"product_code,omitempty"`
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
	ProductEnvironments []EnvironmentResponse `json:"product_environments,omitempty"`
	IsActive            bool                  `json:"is_active"`
	TimestampResponse
}
package dto

import "time"

type ProductAccessScopeInput struct {
	ProjectID  *int   `json:"project_id,omitempty" binding:"omitempty,gt=0"`
	CategoryID *int   `json:"category_id,omitempty" binding:"omitempty,gt=0"`
	ScopeLevel string `json:"scope_level" binding:"required,oneof=PRODUCT PROJECT CATEGORY"`
}

type UpsertProductAccessRequest struct {
	RoleID    int                       `json:"role_id" binding:"required,gt=0"`
	Scopes    []ProductAccessScopeInput `json:"scopes" binding:"required,min=1,dive"`
	ExpiresAt *time.Time                `json:"expires_at,omitempty"`
	IsActive  *bool                     `json:"is_active,omitempty"`
}

type EffectivePermissionResponse struct {
	ResourceType string `json:"resource_type"`
	Action       string `json:"action"`
	Source       string `json:"source"`
}

type ProductAccessRoleResponse struct {
	RoleID      int                        `json:"role_id"`
	RoleCode    string                     `json:"role_code"`
	RoleName    string                     `json:"role_name"`
	Permissions []RolePermissionAssignment `json:"permissions"`
	AccessLevel string                     `json:"access_level"`
	MemberCount int                        `json:"member_count"`
	IsActive    bool                       `json:"is_active"`
}

type ProductAccessMemberResponse struct {
	MembershipID         int                           `json:"membership_id"`
	UserID               int                           `json:"user_id"`
	Username             *string                       `json:"username,omitempty"`
	FullName             string                        `json:"full_name"`
	Email                string                        `json:"email"`
	RoleID               int                           `json:"role_id"`
	RoleCode             string                        `json:"role_code"`
	RoleName             string                        `json:"role_name"`
	AccessLevel          string                        `json:"access_level"`
	EffectivePermissions []EffectivePermissionResponse `json:"effective_permissions"`
	Scopes               []MembershipScopeResponse     `json:"scopes"`
	ExpiresAt            *time.Time                    `json:"expires_at,omitempty"`
	IsActive             bool                          `json:"is_active"`
}

type ProductAccessOverviewResponse struct {
	ProductID int                           `json:"product_id"`
	Roles     []ProductAccessRoleResponse   `json:"roles"`
	Members   []ProductAccessMemberResponse `json:"members"`
	UpdatedAt time.Time                     `json:"updated_at"`
}

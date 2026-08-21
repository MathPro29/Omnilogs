package dto

import (
	"encoding/json"
	"time"
)

type RolePermissionAssignment struct {
	ResourceType string `json:"resource_type" binding:"required"`
	Action       string `json:"action" binding:"required"`
}

type CreateRoleTemplateRequest struct {
	RoleCode         string  `json:"role_code" binding:"required"`
	RoleName         string  `json:"role_name" binding:"required"`
	Description      *string `json:"description,omitempty"`
	IsSystemTemplate bool    `json:"is_system_template"`
}
type UpdateRoleRequest struct {
	RoleName    *string                    `json:"role_name,omitempty"`
	Description *string                    `json:"description,omitempty"`
	Permissions []RolePermissionAssignment `json:"permissions,omitempty"`
	IsActive    *bool                      `json:"is_active,omitempty"`
}

type DeleteRoleRequest struct {
	ProductID int `json:"product_id" binding:"required,gt=0"`
	RoleID    int `json:"role_id" binding:"required,gt=0"`
}

type RoleResponse struct {
	RoleID      int                        `json:"role_id"`
	RoleCode    string                     `json:"role_code"`
	RoleName    string                     `json:"role_name"`
	Description *string                    `json:"description,omitempty"`
	Permissions []RolePermissionAssignment `json:"permissions,omitempty"`
	IsActive    bool                       `json:"is_active"`
	TimestampResponse
}

type UpdateRoleTemplateRequest struct {
	RoleName    *string         `json:"role_name,omitempty"`
	Description *string         `json:"description,omitempty"`
	Permissions json.RawMessage `json:"permissions,omitempty"`
}

type RoleTemplateRequest struct {
	RoleName    *string         `json:"role_name,omitempty"`
	Description *string         `json:"description,omitempty"`
	Permissions json.RawMessage `json:"permissions,omitempty"`
}

type CreatePlatformRoleRequest struct {
	RoleCode     string          `json:"role_code" binding:"required"`
	RoleName     string          `json:"role_name" binding:"required"`
	Description  *string         `json:"description,omitempty"`
	Permissions  json.RawMessage `json:"permissions" binding:"required"`
	IsSystemRole bool            `json:"is_system_role"`
}
type UpdatePlatformRoleRequest struct {
	RoleName     *string         `json:"role_name,omitempty"`
	Description  *string         `json:"description,omitempty"`
	Permissions  json.RawMessage `json:"permissions,omitempty"`
	IsSystemRole *bool           `json:"is_system_role,omitempty"`
}

type PlatformRoleResponse struct {
	PlatformRoleID int             `json:"platform_role_id"`
	RoleCode       string          `json:"role_code"`
	RoleName       string          `json:"role_name"`
	Description    *string         `json:"description,omitempty"`
	Permissions    json.RawMessage `json:"permissions,omitempty"`
	IsSystemRole   bool            `json:"is_system_role"`
	TimestampResponse
}

type CreateProductRoleRequest struct {
	ProductID   int                        `json:"product_id,omitempty" binding:"omitempty,gt=0"`
	TemplateID  *int                       `json:"template_id,omitempty"`
	RoleCode    string                     `json:"role_code,omitempty"`
	RoleName    string                     `json:"role_name" binding:"required"`
	Permissions []RolePermissionAssignment `json:"permissions,omitempty" binding:"omitempty,dive"`
}

type CreatePlatformMembershipRequest struct {
	UserID         int        `json:"user_id" binding:"required,gt=0"`
	PlatformRoleID int        `json:"platform_role_id" binding:"required,gt=0"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

type UpdatePlatformMembershipRequest struct {
	PlatformRoleID int        `json:"platform_role_id" binding:"required,gt=0"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	IsActive       *bool      `json:"is_active,omitempty"`
}

type PlatformMembershipResponse struct {
	PlatformMembershipID int        `json:"platform_membership_id"`
	UserID               int        `json:"user_id"`
	PlatformRoleID       int        `json:"platform_role_id"`
	IsActive             bool       `json:"is_active"`
	ExpiresAt            *time.Time `json:"expires_at,omitempty"`
	TimestampResponse
}
type UpdateMembershipScopeRequest struct {
	ProjectID  *int   `json:"project_id,omitempty" binding:"omitempty,gt=0"`
	CategoryID *int   `json:"category_id,omitempty" binding:"omitempty,gt=0"`
	ScopeLevel string `json:"scope_level" binding:"required,oneof=PRODUCT PROJECT CATEGORY"`
}
type CreateProductMembershipRequest struct {
	UserID    int        `json:"user_id" binding:"required,gt=0"`
	ProductID int        `json:"product_id,omitempty" binding:"omitempty,gt=0"`
	RoleID    int        `json:"role_id,omitempty" binding:"omitempty,gt=0"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}
type CreateBulkProductMembershipRequest struct {
	Memberships []CreateProductMembershipRequest `json:"memberships" binding:"required,min=1,dive"`
}
type UpdateProductMembershipRequest struct {
	RoleID    *int       `json:"role_id,omitempty" binding:"omitempty,gt=0"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	IsActive  *bool      `json:"is_active,omitempty"`
}
type ProductMembershipResponse struct {
	MembershipID int        `json:"membership_id"`
	UserID       int        `json:"user_id"`
	ProductID    int        `json:"product_id"`
	RoleID       int        `json:"role_id"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	IsActive     bool       `json:"is_active"`
	TimestampResponse
}
type CreateMembershipScopeRequest struct {
	ProjectID  *int   `json:"project_id,omitempty" binding:"omitempty,gt=0"`
	CategoryID *int   `json:"category_id,omitempty" binding:"omitempty,gt=0"`
	ScopeLevel string `json:"scope_level" binding:"required,oneof=PRODUCT PROJECT CATEGORY"`
}
type MembershipScopeResponse struct {
	ScopeID      int    `json:"scope_id"`
	MembershipID int    `json:"membership_id"`
	ProductID    int    `json:"product_id"`
	ProjectID    *int   `json:"project_id,omitempty"`
	CategoryID   *int   `json:"category_id,omitempty"`
	ScopeLevel   string `json:"scope_level"`
	IsActive     bool   `json:"is_active"`
	TimestampResponse
}

type CreatePermissionRuleRequest struct {
	UserID       int        `json:"user_id" binding:"required,gt=0"`
	ProductID    *int       `json:"product_id,omitempty" binding:"omitempty,gt=0"`
	RoleID       *int       `json:"role_id,omitempty" binding:"omitempty,gt=0"`
	ProjectID    *int       `json:"project_id,omitempty" binding:"omitempty,gt=0"`
	CategoryID   *int       `json:"category_id,omitempty" binding:"omitempty,gt=0"`
	ResourceType string     `json:"resource_type" binding:"required,oneof=PRODUCT PROJECT FEATURE CATEGORY ROLE ACCESS USER API_KEY ENVIRONMENT LOG"`
	Action       string     `json:"action" binding:"required,oneof=CREATE READ UPDATE DELETE GRANT REVOKE EXPORT VIEW_SENSITIVE"`
	Effect       string     `json:"effect" binding:"required,oneof=ALLOW DENY"`
	ScopeLevel   string     `json:"scope_level" binding:"required,oneof=GLOBAL PRODUCT PROJECT CATEGORY"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}
type UpdatePermissionRuleRequest struct {
	Effect    *string    `json:"effect,omitempty" binding:"omitempty,oneof=ALLOW DENY"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	IsActive  *bool      `json:"is_active,omitempty"`
}
type PermissionRuleResponse struct {
	PermissionRuleID int        `json:"permission_rule_id"`
	UserID           int        `json:"user_id"`
	ProductID        *int       `json:"product_id,omitempty"`
	RoleID           *int       `json:"role_id,omitempty"`
	ProjectID        *int       `json:"project_id,omitempty"`
	CategoryID       *int       `json:"category_id,omitempty"`
	ResourceType     string     `json:"resource_type"`
	Action           string     `json:"action"`
	Effect           string     `json:"effect"`
	ScopeLevel       string     `json:"scope_level"`
	GrantedBy        *int       `json:"granted_by,omitempty"`
	IsActive         bool       `json:"is_active"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	TimestampResponse
}
type PermissionCheckRequest struct {
	ResourceType  string `json:"resource_type" binding:"required"`
	Action        string `json:"action" binding:"required"`
	ProductID     *int   `json:"product_id,omitempty"`
	EnvironmentID *int   `json:"environment_id,omitempty"`
	ProjectID     *int   `json:"project_id,omitempty"`
	CategoryID    *int   `json:"category_id,omitempty"`
}
type PermissionCheckResponse struct {
	Allowed       bool   `json:"allowed"`
	MatchedRuleID *int   `json:"matched_rule_id,omitempty"`
	Reason        string `json:"reason,omitempty"`
}

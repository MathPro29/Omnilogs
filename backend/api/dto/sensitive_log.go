package dto

import "time"

type CreateSensitiveLogAccessRequest struct {
	ProductID         int     `json:"product_id" binding:"required,gt=0"`
	LogID             *string `json:"log_id,omitempty" binding:"omitempty,uuid"`
	FieldDefinitionID *int    `json:"field_definition_id,omitempty" binding:"omitempty,gt=0"`
	SecretID          *string `json:"secret_id,omitempty" binding:"omitempty,uuid"`
	FieldPath         *string `json:"field_path,omitempty"`
	Reason            string  `json:"reason" binding:"required"`
}

type ReviewSensitiveLogAccessRequest struct {
	ApprovalStatus string     `json:"approval_status" binding:"required,oneof=APPROVED REJECTED"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

type RevealSensitiveLogRequest struct {
	RequestID string `json:"request_id" binding:"required,uuid"`
	Password  string `json:"password" binding:"required"`
}

type SensitiveLogAccessRequestResponse struct {
	RequestID         string     `json:"request_id"`
	UserID            int        `json:"user_id"`
	ProductID         int        `json:"product_id"`
	LogID             *string    `json:"log_id,omitempty"`
	FieldDefinitionID *int       `json:"field_definition_id,omitempty"`
	SecretID          *string    `json:"secret_id,omitempty"`
	FieldPath         *string    `json:"field_path,omitempty"`
	Reason            *string    `json:"reason,omitempty"`
	ApprovalStatus    string     `json:"approval_status"`
	ApprovedBy        *int       `json:"approved_by,omitempty"`
	ApprovedAt        *time.Time `json:"approved_at,omitempty"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	TimestampResponse
}

// SensitiveLogSecretResponse intentionally excludes encrypted_value, value_hash,
// and encryption key metadata. Plaintext is returned only by the reveal endpoint.
type SensitiveLogSecretResponse struct {
	SecretID          string     `json:"secret_id"`
	LogID             string     `json:"log_id"`
	ProductID         int        `json:"product_id"`
	ProjectID         *int       `json:"project_id,omitempty"`
	CategoryID        *int       `json:"category_id,omitempty"`
	EnvironmentID     *int       `json:"environment_id,omitempty"`
	FieldDefinitionID *int       `json:"field_definition_id,omitempty"`
	FieldKey          string     `json:"field_key"`
	FieldPath         string     `json:"field_path"`
	SourceSection     string     `json:"source_section"`
	RequiresApproval  bool       `json:"requires_approval"`
	RetentionUntil    *time.Time `json:"retention_until,omitempty"`
	PurgedAt          *time.Time `json:"purged_at,omitempty"`
	CreatedAt         *time.Time `json:"created_at,omitempty"`
}

type RevealedSensitiveLogResponse struct {
	SecretID  string     `json:"secret_id"`
	FieldKey  string     `json:"field_key"`
	FieldPath string     `json:"field_path"`
	Value     any        `json:"value"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type SensitiveLogAccessHistoryResponse struct {
	AccessID     string     `json:"access_id"`
	RequestID    *string    `json:"request_id,omitempty"`
	ProductID    *int       `json:"product_id,omitempty"`
	UserID       int        `json:"user_id"`
	ProjectID    *int       `json:"project_id,omitempty"`
	CategoryID   *int       `json:"category_id,omitempty"`
	LogID        *string    `json:"log_id,omitempty"`
	FieldPath    *string    `json:"field_path,omitempty"`
	AccessReason *string    `json:"access_reason,omitempty"`
	AuthMethod   *string    `json:"auth_method,omitempty"`
	AccessStatus *string    `json:"access_status,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
}

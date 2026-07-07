package dto

import "time"

type CreateAuditSecretAccessRequest struct {
	SecretID string `json:"secret_id" binding:"required,uuid"`
	Reason   string `json:"reason" binding:"required"`
}

type ReviewAuditSecretAccessRequest struct {
	ApprovalStatus string     `json:"approval_status" binding:"required,oneof=APPROVED REJECTED"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

type RevealAuditSecretRequest struct {
	RequestID string `json:"request_id" binding:"required,uuid"`
	Password  string `json:"password" binding:"required"`
}

type RevealedAuditSecretResponse struct {
	SecretID      string     `json:"secret_id"`
	AuditID       string     `json:"audit_id"`
	FieldKey      string     `json:"field_key"`
	FieldPath     string     `json:"field_path"`
	SourceSection string     `json:"source_section"`
	Value         any        `json:"value"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
}

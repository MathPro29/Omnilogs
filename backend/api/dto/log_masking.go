package dto

type CreateLogMaskingRuleRequest struct {
	ProductID            *int    `json:"product_id,omitempty" binding:"omitempty,gt=0"`
	FieldKey             *string `json:"field_key,omitempty"`
	FieldPath            *string `json:"field_path,omitempty"`
	MatchType            *string `json:"match_type,omitempty"`
	MaskType             *string `json:"mask_type,omitempty"`
	MaskValue            *string `json:"mask_value,omitempty"`
	SecretHandlingMode   string  `json:"secret_handling_mode" binding:"required"`
	EncryptionRequired   bool    `json:"encryption_required"`
	ViewRequiresPassword bool    `json:"view_requires_password"`
}

type UpdateLogMaskingRuleRequest struct {
	FieldKey             *string `json:"field_key,omitempty"`
	FieldPath            *string `json:"field_path,omitempty"`
	MatchType            *string `json:"match_type,omitempty"`
	MaskType             *string `json:"mask_type,omitempty"`
	MaskValue            *string `json:"mask_value,omitempty"`
	SecretHandlingMode   *string `json:"secret_handling_mode,omitempty"`
	EncryptionRequired   *bool   `json:"encryption_required,omitempty"`
	ViewRequiresPassword *bool   `json:"view_requires_password,omitempty"`
	IsActive             *bool   `json:"is_active,omitempty"`
}

type LogMaskingRuleResponse struct {
	MaskingRuleID        int     `json:"masking_rule_id"`
	ProductID            *int    `json:"product_id,omitempty"`
	FieldKey             *string `json:"field_key,omitempty"`
	FieldPath            *string `json:"field_path,omitempty"`
	MatchType            *string `json:"match_type,omitempty"`
	MaskType             *string `json:"mask_type,omitempty"`
	MaskValue            *string `json:"mask_value,omitempty"`
	SecretHandlingMode   string  `json:"secret_handling_mode"`
	EncryptionRequired   bool    `json:"encryption_required"`
	ViewRequiresPassword bool    `json:"view_requires_password"`
	IsActive             bool    `json:"is_active"`
	TimestampResponse
}

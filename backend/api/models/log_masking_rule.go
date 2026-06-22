package models

type LogMaskingRule struct {
	MaskingRuleID        int     `gorm:"primaryKey;autoIncrement" json:"masking_rule_id"`
	ProductID            *int    `gorm:"index:idx_mask_product_active,priority:1;index:idx_mask_product_key,priority:1;index:idx_mask_product_path,priority:1;index:idx_mask_product_secret_mode,priority:1" json:"product_id,omitempty"`
	FieldKey             *string `gorm:"index:idx_mask_product_key,priority:2" json:"field_key,omitempty"`
	FieldPath            *string `gorm:"index:idx_mask_product_path,priority:2" json:"field_path,omitempty"`
	MatchType            *string `json:"match_type,omitempty"`
	MaskType             *string `json:"mask_type,omitempty"`
	MaskValue            *string `json:"mask_value,omitempty"`
	SecretHandlingMode   string  `gorm:"not null;default:MASK_ONLY;index:idx_mask_product_secret_mode,priority:2" json:"secret_handling_mode"`
	EncryptionRequired   bool    `gorm:"not null;default:false" json:"encryption_required"`
	ViewRequiresPassword bool    `gorm:"not null;default:false" json:"view_requires_password"`
	IsActive             bool    `gorm:"not null;default:true;index:idx_mask_product_active,priority:2" json:"is_active"`
	Timestamps
}

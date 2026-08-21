package models

type FieldConfigJSON struct {
	StringConfig *StringFieldConfig     `json:"string_config,omitempty"`
	NumberConfig *NumberFieldConfig     `json:"number_config,omitempty"`
	DateConfig   *DateFieldConfig       `json:"date_config,omitempty"`
	EnumConfig   *EnumFieldConfig       `json:"enum_config,omitempty"`
	Validation   *FieldValidationConfig `json:"validation,omitempty"`
	Transform    []TransformationStep   `json:"transform,omitempty"`
	Display      *FieldDisplayConfig    `json:"display,omitempty"`
	Security     *FieldSecurityConfig   `json:"security,omitempty"`
	Computed     *FieldComputedConfig   `json:"computed,omitempty"`
}

type StringFieldConfig struct {
	MinLength *int    `json:"min_length,omitempty"`
	MaxLength *int    `json:"max_length,omitempty"`
	Regex     *string `json:"regex,omitempty"`
	Trim      bool    `json:"trim,omitempty"`
	Uppercase bool    `json:"uppercase,omitempty"`
	Lowercase bool    `json:"lowercase,omitempty"`
}

type NumberFieldConfig struct {
	Min       *float64 `json:"min,omitempty"`
	Max       *float64 `json:"max,omitempty"`
	Decimal   bool     `json:"decimal,omitempty"`
	Precision *int     `json:"precision,omitempty"`
}

type DateFieldConfig struct {
	Format   *string `json:"format,omitempty"`
	Timezone *string `json:"timezone,omitempty"`
}

type EnumFieldConfig struct {
	Options  []EnumOptionConfig `json:"options,omitempty"`
	Source   string             `json:"source,omitempty"`
	CacheTTL *int               `json:"cache_ttl,omitempty"`
}

type EnumOptionConfig struct {
	Label       string  `json:"label"`
	Value       string  `json:"value"`
	Color       *string `json:"color,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	Description *string `json:"description,omitempty"`
}

type FieldValidationConfig struct {
	Required    bool    `json:"required,omitempty"`
	Nullable    bool    `json:"nullable,omitempty"`
	JSONSchema  *string `json:"json_schema,omitempty"`
	EmailDomain *string `json:"email_domain,omitempty"`
}

type TransformationStep struct {
	Type   string         `json:"type"`
	Params map[string]any `json:"params,omitempty"`
}

type FieldDisplayConfig struct {
	Type       string             `json:"type,omitempty"`
	Conditions []DisplayCondition `json:"conditions,omitempty"`
}

type DisplayCondition struct {
	Operator string `json:"operator"`
	Value    any    `json:"value"`
	Action   string `json:"action"`
}

type FieldSecurityConfig struct {
	MaskType          *string `json:"mask_type,omitempty"`
	EncryptRequired   bool    `json:"encrypt_required,omitempty"`
	RequirePermission bool    `json:"require_permission,omitempty"`
}

type FieldComputedConfig struct {
	IsComputed bool   `json:"is_computed,omitempty"`
	Expression string `json:"expression,omitempty"`
}

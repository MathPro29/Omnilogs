package dto

import (
	"encoding/json"
	"time"
)

type CreateLogSourceRequest struct {
	ProductID      *int    `json:"product_id,omitempty" binding:"omitempty,gt=0"`
	SourceCode     string  `json:"source_code" binding:"required"`
	SourceName     string  `json:"source_name" binding:"required"`
	SourceType     string  `json:"source_type" binding:"required"`
	SourcePlatform string  `json:"source_platform" binding:"required"`
	RuntimeName    *string `json:"runtime_name,omitempty"`
	SDKName        *string `json:"sdk_name,omitempty"`
	SDKVersion     *string `json:"sdk_version,omitempty"`
	Description    *string `json:"description,omitempty"`
}

type UpdateLogSourceRequest struct {
	SourceName      *string `json:"source_name,omitempty"`
	SourceType      *string `json:"source_type,omitempty"`
	SourcePlatform  *string `json:"source_platform,omitempty"`
	RuntimeName     *string `json:"runtime_name,omitempty"`
	SDKName         *string `json:"sdk_name,omitempty"`
	SDKVersion      *string `json:"sdk_version,omitempty"`
	DiscoveryStatus *string `json:"discovery_status,omitempty"`
	Description     *string `json:"description,omitempty"`
	IsActive        *bool   `json:"is_active,omitempty"`
}

type LogSourceResponse struct {
	SourceID        int        `json:"source_id"`
	ProductID       *int       `json:"product_id,omitempty"`
	SourceCode      string     `json:"source_code"`
	SourceName      string     `json:"source_name"`
	SourceType      string     `json:"source_type"`
	SourcePlatform  string     `json:"source_platform"`
	RuntimeName     *string    `json:"runtime_name,omitempty"`
	SDKName         *string    `json:"sdk_name,omitempty"`
	SDKVersion      *string    `json:"sdk_version,omitempty"`
	DiscoveryStatus string     `json:"discovery_status"`
	CreatedBySource string     `json:"created_by_source"`
	FirstSeenAt     *time.Time `json:"first_seen_at,omitempty"`
	LastSeenAt      *time.Time `json:"last_seen_at,omitempty"`
	Description     *string    `json:"description,omitempty"`
	IsActive        bool       `json:"is_active"`
	TimestampResponse
}

type UpsertLogIngestionPolicyRequest struct {
	ProductID               int     `json:"product_id" binding:"required,gt=0"`
	EnvironmentID           int     `json:"environment_id" binding:"required,gt=0"`
	MaxLogsPerMinute        *int    `json:"max_logs_per_minute,omitempty" binding:"omitempty,gt=0"`
	MaxBatchSize            *int    `json:"max_batch_size,omitempty" binding:"omitempty,gt=0"`
	FlushIntervalSeconds    *int    `json:"flush_interval_seconds,omitempty" binding:"omitempty,gt=0"`
	MaxRetries              *int    `json:"max_retries,omitempty" binding:"omitempty,gte=0"`
	RetryIntervalSeconds    *int    `json:"retry_interval_seconds,omitempty" binding:"omitempty,gte=0"`
	QueueEnabled            *bool   `json:"queue_enabled,omitempty"`
	QueueWorkerEnabled      *bool   `json:"queue_worker_enabled,omitempty"`
	QueueLockTimeoutSeconds *int    `json:"queue_lock_timeout_seconds,omitempty" binding:"omitempty,gt=0"`
	StrictOrderingEnabled   *bool   `json:"strict_ordering_enabled,omitempty"`
	RetentionDays           *int    `json:"retention_days,omitempty" binding:"omitempty,gte=0"`
	ArchiveEnabled          *bool   `json:"archive_enabled,omitempty"`
	ArchiveFormat           *string `json:"archive_format,omitempty"`
	ArchiveStoragePath      *string `json:"archive_storage_path,omitempty"`
}

type LogIngestionPolicyResponse struct {
	PolicyID                int     `json:"policy_id"`
	ProductID               int     `json:"product_id"`
	EnvironmentID           int     `json:"environment_id"`
	MaxLogsPerMinute        *int    `json:"max_logs_per_minute,omitempty"`
	MaxBatchSize            *int    `json:"max_batch_size,omitempty"`
	FlushIntervalSeconds    *int    `json:"flush_interval_seconds,omitempty"`
	MaxRetries              *int    `json:"max_retries,omitempty"`
	RetryIntervalSeconds    *int    `json:"retry_interval_seconds,omitempty"`
	QueueEnabled            bool    `json:"queue_enabled"`
	QueueWorkerEnabled      bool    `json:"queue_worker_enabled"`
	QueueLockTimeoutSeconds int     `json:"queue_lock_timeout_seconds"`
	StrictOrderingEnabled   bool    `json:"strict_ordering_enabled"`
	RetentionDays           *int    `json:"retention_days,omitempty"`
	ArchiveEnabled          bool    `json:"archive_enabled"`
	ArchiveFormat           *string `json:"archive_format,omitempty"`
	ArchiveStoragePath      *string `json:"archive_storage_path,omitempty"`
	TimestampResponse
}

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

type CreateLogFieldDefinitionRequest struct {
	ProductID            *int    `json:"product_id,omitempty" binding:"omitempty,gt=0"`
	ProjectID            *int    `json:"project_id,omitempty" binding:"omitempty,gt=0"`
	CategoryID           *int    `json:"category_id,omitempty" binding:"omitempty,gt=0"`
	FieldKey             string  `json:"field_key" binding:"required"`
	DisplayName          *string `json:"display_name,omitempty"`
	Description          *string `json:"description,omitempty"`
	SourceSection        string  `json:"source_section" binding:"required"`
	FieldPath            *string `json:"field_path,omitempty"`
	ElasticFieldName     string  `json:"elastic_field_name" binding:"required"`
	DataType             string  `json:"data_type" binding:"required"`
	ValueSourceType      string  `json:"value_source_type" binding:"required"`
	ValueSourceKey       *string `json:"value_source_key,omitempty"`
	IsRequired           bool    `json:"is_required"`
	IsSensitive          bool    `json:"is_sensitive"`
	MaskBeforeIndex      bool    `json:"mask_before_index"`
	EncryptBeforeArchive bool    `json:"encrypt_before_archive"`
	IsVisible            *bool   `json:"is_visible,omitempty"`
	IsSearchable         bool    `json:"is_searchable"`
	IsFilterable         bool    `json:"is_filterable"`
	IsSortable           bool    `json:"is_sortable"`
	IsAggregatable       bool    `json:"is_aggregatable"`
	DisplayOrder         *int    `json:"display_order,omitempty" binding:"omitempty,gte=0"`
	DefaultValue         *string `json:"default_value,omitempty"`
}

type UpdateLogFieldDefinitionRequest struct {
	DisplayName          *string `json:"display_name,omitempty"`
	Description          *string `json:"description,omitempty"`
	FieldPath            *string `json:"field_path,omitempty"`
	ElasticFieldName     *string `json:"elastic_field_name,omitempty"`
	DataType             *string `json:"data_type,omitempty"`
	ValueSourceType      *string `json:"value_source_type,omitempty"`
	ValueSourceKey       *string `json:"value_source_key,omitempty"`
	IsRequired           *bool   `json:"is_required,omitempty"`
	IsSensitive          *bool   `json:"is_sensitive,omitempty"`
	MaskBeforeIndex      *bool   `json:"mask_before_index,omitempty"`
	EncryptBeforeArchive *bool   `json:"encrypt_before_archive,omitempty"`
	IsVisible            *bool   `json:"is_visible,omitempty"`
	IsSearchable         *bool   `json:"is_searchable,omitempty"`
	IsFilterable         *bool   `json:"is_filterable,omitempty"`
	IsSortable           *bool   `json:"is_sortable,omitempty"`
	IsAggregatable       *bool   `json:"is_aggregatable,omitempty"`
	DisplayOrder         *int    `json:"display_order,omitempty" binding:"omitempty,gte=0"`
	DefaultValue         *string `json:"default_value,omitempty"`
	IsActive             *bool   `json:"is_active,omitempty"`
}

type LogFieldDefinitionResponse struct {
	FieldDefinitionID    int     `json:"field_definition_id"`
	ProductID            *int    `json:"product_id,omitempty"`
	ProjectID            *int    `json:"project_id,omitempty"`
	CategoryID           *int    `json:"category_id,omitempty"`
	FieldKey             string  `json:"field_key"`
	DisplayName          *string `json:"display_name,omitempty"`
	Description          *string `json:"description,omitempty"`
	SourceSection        string  `json:"source_section"`
	FieldPath            *string `json:"field_path,omitempty"`
	ElasticFieldName     string  `json:"elastic_field_name"`
	DataType             string  `json:"data_type"`
	ValueSourceType      string  `json:"value_source_type"`
	ValueSourceKey       *string `json:"value_source_key,omitempty"`
	IsRequired           bool    `json:"is_required"`
	IsSensitive          bool    `json:"is_sensitive"`
	MaskBeforeIndex      bool    `json:"mask_before_index"`
	EncryptBeforeArchive bool    `json:"encrypt_before_archive"`
	IsVisible            bool    `json:"is_visible"`
	IsSearchable         bool    `json:"is_searchable"`
	IsFilterable         bool    `json:"is_filterable"`
	IsSortable           bool    `json:"is_sortable"`
	IsAggregatable       bool    `json:"is_aggregatable"`
	DisplayOrder         *int    `json:"display_order,omitempty"`
	DefaultValue         *string `json:"default_value,omitempty"`
	IsActive             bool    `json:"is_active"`
	TimestampResponse
}

type CreateLogFieldEnumOptionRequest struct {
	FieldDefinitionID int    `json:"field_definition_id" binding:"required,gt=0"`
	OptionKey         string `json:"option_key" binding:"required"`
	OptionLabel       string `json:"option_label" binding:"required"`
	OptionValue       string `json:"option_value" binding:"required"`
	DisplayOrder      *int   `json:"display_order,omitempty" binding:"omitempty,gte=0"`
	IsDefault         bool   `json:"is_default"`
}

type UpdateLogFieldEnumOptionRequest struct {
	OptionLabel  *string `json:"option_label,omitempty"`
	OptionValue  *string `json:"option_value,omitempty"`
	DisplayOrder *int    `json:"display_order,omitempty" binding:"omitempty,gte=0"`
	IsDefault    *bool   `json:"is_default,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
}

type LogFieldEnumOptionResponse struct {
	OptionID int `json:"option_id"`
	CreateLogFieldEnumOptionRequest
	IsActive bool `json:"is_active"`
	TimestampResponse
}

type CreateLogFieldValueSourceRequest struct {
	FieldDefinitionID int             `json:"field_definition_id" binding:"required,gt=0"`
	SourceType        string          `json:"source_type" binding:"required"`
	SourceKey         string          `json:"source_key" binding:"required"`
	SourceConfig      json.RawMessage `json:"source_config,omitempty"`
	LabelField        *string         `json:"label_field,omitempty"`
	ValueField        *string         `json:"value_field,omitempty"`
	IsRequired        bool            `json:"is_required"`
	TimeoutSeconds    *int            `json:"timeout_seconds,omitempty" binding:"omitempty,gt=0"`
	RefreshMode       string          `json:"refresh_mode" binding:"required"`
}

type UpdateLogFieldValueSourceRequest struct {
	SourceType     *string         `json:"source_type,omitempty"`
	SourceKey      *string         `json:"source_key,omitempty"`
	SourceConfig   json.RawMessage `json:"source_config,omitempty"`
	LabelField     *string         `json:"label_field,omitempty"`
	ValueField     *string         `json:"value_field,omitempty"`
	IsRequired     *bool           `json:"is_required,omitempty"`
	TimeoutSeconds *int            `json:"timeout_seconds,omitempty" binding:"omitempty,gt=0"`
	RefreshMode    *string         `json:"refresh_mode,omitempty"`
	IsActive       *bool           `json:"is_active,omitempty"`
}

type LogFieldValueSourceResponse struct {
	ValueSourceID int `json:"value_source_id"`
	CreateLogFieldValueSourceRequest
	IsActive bool `json:"is_active"`
	TimestampResponse
}

type CreateElasticIndexPolicyRequest struct {
	ProductID        int     `json:"product_id" binding:"required,gt=0"`
	EnvironmentID    *int    `json:"environment_id,omitempty" binding:"omitempty,gt=0"`
	IndexPrefix      string  `json:"index_prefix" binding:"required"`
	IndexPattern     *string `json:"index_pattern,omitempty"`
	WriteAlias       *string `json:"write_alias,omitempty"`
	RolloverType     string  `json:"rollover_type" binding:"required"`
	NumberOfShards   *int    `json:"number_of_shards,omitempty" binding:"omitempty,gt=0"`
	NumberOfReplicas *int    `json:"number_of_replicas,omitempty" binding:"omitempty,gte=0"`
	RetentionDays    *int    `json:"retention_days,omitempty" binding:"omitempty,gte=0"`
	SchemaVersion    *string `json:"schema_version,omitempty"`
}

type UpdateElasticIndexPolicyRequest struct {
	IndexPattern     *string `json:"index_pattern,omitempty"`
	WriteAlias       *string `json:"write_alias,omitempty"`
	RolloverType     *string `json:"rollover_type,omitempty"`
	NumberOfShards   *int    `json:"number_of_shards,omitempty" binding:"omitempty,gt=0"`
	NumberOfReplicas *int    `json:"number_of_replicas,omitempty" binding:"omitempty,gte=0"`
	RetentionDays    *int    `json:"retention_days,omitempty" binding:"omitempty,gte=0"`
	SchemaVersion    *string `json:"schema_version,omitempty"`
	IsActive         *bool   `json:"is_active,omitempty"`
}

type ElasticIndexPolicyResponse struct {
	ElasticPolicyID  int     `json:"elastic_policy_id"`
	ProductID        int     `json:"product_id"`
	EnvironmentID    *int    `json:"environment_id,omitempty"`
	IndexPrefix      string  `json:"index_prefix"`
	IndexPattern     *string `json:"index_pattern,omitempty"`
	WriteAlias       *string `json:"write_alias,omitempty"`
	RolloverType     string  `json:"rollover_type"`
	NumberOfShards   int     `json:"number_of_shards"`
	NumberOfReplicas int     `json:"number_of_replicas"`
	RetentionDays    *int    `json:"retention_days,omitempty"`
	SchemaVersion    string  `json:"schema_version"`
	IsActive         bool    `json:"is_active"`
	TimestampResponse
}

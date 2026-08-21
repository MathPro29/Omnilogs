package dto

import "time"

type CreateRetentionPolicyRequest struct {
	ProductID        int     `json:"product_id"`
	EnvironmentID    *int    `json:"environment_id,omitempty"`
	ProjectID        *int    `json:"project_id,omitempty"`
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	RetentionMode    string  `json:"retention_mode"` // legacy: DAILY, WEEKLY, MONTHLY, CUSTOM
	RetentionUnit    string  `json:"retention_unit"` // legacy: DAYS, WEEKS, MONTHS, YEARS
	RetentionValue   int     `json:"retention_value,omitempty"`
	FolderStructure  string  `json:"folder_structure"`  // 7_DAILY_FOLDERS, 4_WEEKLY_SUBFOLDERS, MONTHLY_FOLDERS, CUSTOM
	AutoPurgeAction  string  `json:"auto_purge_action"` // DELETE, ARCHIVE_COLD, COMPRESS_GZIP
	StorageProvider  string  `json:"storage_provider"`  // S3, GCS, MINIO, LOCAL
	BucketName       *string `json:"bucket_name,omitempty"`
	CronSchedule     string  `json:"cron_schedule"`
	ScheduleTimezone string  `json:"schedule_timezone"`
	IsActive         *bool   `json:"is_active,omitempty"`

	ActiveRetentionValue     int    `json:"active_retention_value,omitempty"`
	ActiveRetentionUnit      string `json:"active_retention_unit,omitempty"`
	ArchiveEnabled           bool   `json:"archive_enabled"`
	ArchiveAfterValue        *int   `json:"archive_after_value,omitempty"`
	ArchiveAfterUnit         string `json:"archive_after_unit,omitempty"`
	DeleteActiveAfterArchive bool   `json:"delete_active_after_archive"`
	ArchiveRetentionValue    *int   `json:"archive_retention_value,omitempty"`
	ArchiveRetentionUnit     string `json:"archive_retention_unit,omitempty"`
	ArchiveNeverDelete       bool   `json:"archive_never_delete"`
	ApplyToExistingLogs      bool   `json:"apply_to_existing_logs"`
}

type UpdateRetentionPolicyRequest struct {
	PolicyID         int     `json:"policy_id"`
	ProductID        int     `json:"product_id"`
	Name             *string `json:"name,omitempty"`
	Description      *string `json:"description,omitempty"`
	RetentionMode    *string `json:"retention_mode,omitempty"`
	RetentionUnit    *string `json:"retention_unit,omitempty"`
	RetentionValue   *int    `json:"retention_value,omitempty"`
	FolderStructure  *string `json:"folder_structure,omitempty"`
	AutoPurgeAction  *string `json:"auto_purge_action,omitempty"`
	StorageProvider  *string `json:"storage_provider,omitempty"`
	BucketName       *string `json:"bucket_name,omitempty"`
	CronSchedule     *string `json:"cron_schedule,omitempty"`
	ScheduleTimezone *string `json:"schedule_timezone,omitempty"`
	IsActive         *bool   `json:"is_active,omitempty"`

	EnvironmentID            *int    `json:"environment_id,omitempty"`
	ActiveRetentionValue     *int    `json:"active_retention_value,omitempty"`
	ActiveRetentionUnit      *string `json:"active_retention_unit,omitempty"`
	ArchiveEnabled           *bool   `json:"archive_enabled,omitempty"`
	ArchiveAfterValue        *int    `json:"archive_after_value,omitempty"`
	ArchiveAfterUnit         *string `json:"archive_after_unit,omitempty"`
	DeleteActiveAfterArchive *bool   `json:"delete_active_after_archive,omitempty"`
	ArchiveRetentionValue    *int    `json:"archive_retention_value,omitempty"`
	ArchiveRetentionUnit     *string `json:"archive_retention_unit,omitempty"`
	ArchiveNeverDelete       *bool   `json:"archive_never_delete,omitempty"`
	ApplyToExistingLogs      *bool   `json:"apply_to_existing_logs,omitempty"`
}

type RetentionPolicyResponse struct {
	PolicyID         int        `json:"policy_id"`
	ProductID        int        `json:"product_id"`
	EnvironmentID    *int       `json:"environment_id,omitempty"`
	ProjectID        *int       `json:"project_id,omitempty"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	RetentionMode    string     `json:"retention_mode"`
	RetentionUnit    string     `json:"retention_unit"`
	RetentionValue   int        `json:"retention_value"`
	TotalDays        int        `json:"total_days"`
	FolderStructure  string     `json:"folder_structure"`
	AutoPurgeAction  string     `json:"auto_purge_action"`
	StorageProvider  string     `json:"storage_provider"`
	BucketName       *string    `json:"bucket_name,omitempty"`
	CronSchedule     string     `json:"cron_schedule"`
	ScheduleTimezone string     `json:"schedule_timezone"`
	IsActive         bool       `json:"is_active"`
	LastPurgeAt      *time.Time `json:"last_purge_at,omitempty"`
	NextPurgeAt      *time.Time `json:"next_purge_at,omitempty"`
	CreatedAt        *time.Time `json:"created_at,omitempty"`
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`

	ActiveRetentionValue     int        `json:"active_retention_value"`
	ActiveRetentionUnit      string     `json:"active_retention_unit"`
	ArchiveEnabled           bool       `json:"archive_enabled"`
	ArchiveAfterValue        *int       `json:"archive_after_value,omitempty"`
	ArchiveAfterUnit         string     `json:"archive_after_unit,omitempty"`
	DeleteActiveAfterArchive bool       `json:"delete_active_after_archive"`
	ArchiveRetentionValue    *int       `json:"archive_retention_value,omitempty"`
	ArchiveRetentionUnit     string     `json:"archive_retention_unit,omitempty"`
	ArchiveNeverDelete       bool       `json:"archive_never_delete"`
	ApplyToExistingLogs      bool       `json:"apply_to_existing_logs"`
	EffectiveFrom            *time.Time `json:"effective_from,omitempty"`
}

type SimulateRetentionRequest struct {
	RetentionMode   string `json:"retention_mode"`   // DAILY, WEEKLY, MONTHLY, CUSTOM
	RetentionUnit   string `json:"retention_unit"`   // DAYS, WEEKS, MONTHS
	RetentionValue  int    `json:"retention_value"`  // e.g. 1
	FolderStructure string `json:"folder_structure"` // 7_DAILY_FOLDERS, 4_WEEKLY_SUBFOLDERS, etc.
}

type FolderNodeDTO struct {
	Key         string          `json:"key"`
	Title       string          `json:"title"`
	Path        string          `json:"path"`
	Type        string          `json:"type"` // root, month, week, day, file
	FolderCount int             `json:"folder_count"`
	FileCount   int             `json:"file_count"`
	EstimatedMB float64         `json:"estimated_mb"`
	Status      string          `json:"status"` // ACTIVE, ARCHIVED, TO_BE_PURGED
	ExpiryDate  string          `json:"expiry_date,omitempty"`
	Children    []FolderNodeDTO `json:"children,omitempty"`
}

type SimulationResultDTO struct {
	RetentionMode        string        `json:"retention_mode"`
	TotalRetentionDays   int           `json:"total_retention_days"`
	CalculatedCutoffDate string        `json:"calculated_cutoff_date"`
	TotalFoldersCount    int           `json:"total_folders_count"`
	ActiveFoldersCount   int           `json:"active_folders_count"`
	PurgeFoldersCount    int           `json:"purge_folders_count"`
	EstimatedSavingsMB   float64       `json:"estimated_savings_mb"`
	TreeData             FolderNodeDTO `json:"tree_data"`
}

type PolicyStatsDTO struct {
	TotalPolicies           int     `json:"total_policies"`
	ActivePolicies          int     `json:"active_policies"`
	TotalStorageBytes       int64   `json:"total_storage_bytes"`
	ActiveStorageBytes      int64   `json:"active_storage_bytes"`
	ArchiveGZIPStorageBytes int64   `json:"archive_gzip_storage_bytes"`
	ReadyZIPStorageBytes    int64   `json:"ready_zip_storage_bytes"`
	TotalStorageGB          float64 `json:"total_storage_gb"`
	ScheduledPurges24h      int     `json:"scheduled_purges_24h"`
	TotalFoldersCount       int     `json:"total_folders_count"`
}

type RetentionHistoricalRequest struct {
	ProductID     int        `json:"product_id" form:"product_id"`
	EnvironmentID int        `json:"environment_id" form:"environment_id" binding:"required,gt=0"`
	DateFrom      *time.Time `json:"date_from,omitempty" form:"date_from"`
	DateTo        *time.Time `json:"date_to,omitempty" form:"date_to"`
	GroupBy       string     `json:"group_by,omitempty" form:"group_by"`
	CategoryID    *int       `json:"category_id,omitempty" form:"category_id"`
	FeatureID     *int       `json:"feature_id,omitempty" form:"feature_id"`
	SubFeatureID  *int       `json:"sub_feature_id,omitempty" form:"sub_feature_id"`
}

type RetentionPreviewRequest struct {
	RetentionHistoricalRequest
	PolicyID *int `json:"policy_id,omitempty" form:"policy_id"`
}

type HistoricalLogBucket struct {
	Period         string     `json:"period"`
	DateFrom       time.Time  `json:"date_from"`
	DateTo         time.Time  `json:"date_to"`
	OldestLog      *time.Time `json:"oldest_log,omitempty"`
	NewestLog      *time.Time `json:"newest_log,omitempty"`
	DocumentCount  int64      `json:"document_count"`
	EstimatedBytes int64      `json:"estimated_size_bytes"`
	ArchiveStatus  string     `json:"archive_status"`
}

type RetentionHistoricalResponse struct {
	ProductID     int                   `json:"product_id"`
	EnvironmentID int                   `json:"environment_id"`
	OldestLog     *time.Time            `json:"oldest_log,omitempty"`
	NewestLog     *time.Time            `json:"newest_log,omitempty"`
	DocumentCount int64                 `json:"document_count"`
	EstimatedSize int64                 `json:"estimated_size_bytes"`
	ArchiveStatus string                `json:"archive_status"`
	Buckets       []HistoricalLogBucket `json:"buckets"`
}

type RetentionPreviewResponse struct {
	Historical             RetentionHistoricalResponse `json:"historical"`
	ArchiveCandidateCount  int64                       `json:"archive_candidate_count"`
	DeleteCandidateCount   int64                       `json:"delete_candidate_count"`
	DeleteRequiresVerified bool                        `json:"delete_requires_verified_archive"`
	ActiveCutoff           *time.Time                  `json:"active_cutoff,omitempty"`
	ArchiveCutoff          *time.Time                  `json:"archive_cutoff,omitempty"`
}

package models

import "time"

// LogRetentionPolicy represents retention policy for log storage and purging
type LogRetentionPolicy struct {
	PolicyID         int        `gorm:"primaryKey;autoIncrement" json:"policy_id"`
	ProductID        int        `gorm:"not null;index:idx_retention_product" json:"product_id"`
	EnvironmentID    *int       `gorm:"index:idx_retention_env" json:"environment_id,omitempty"`
	ProjectID        *int       `gorm:"index:idx_retention_project" json:"project_id,omitempty"`
	Name             string     `gorm:"type:varchar(100);not null" json:"name"`
	Description      string     `gorm:"type:text" json:"description"`
	RetentionMode    string     `gorm:"type:varchar(20);not null;default:'WEEKLY'" json:"retention_mode"` // DAILY, WEEKLY, MONTHLY, CUSTOM
	RetentionUnit    string     `gorm:"type:varchar(20);not null;default:'WEEKS'" json:"retention_unit"`  // DAYS, WEEKS, MONTHS, YEARS
	RetentionValue   int        `gorm:"not null;default:1" json:"retention_value"`
	TotalDays        int        `gorm:"not null;default:7" json:"total_days"`
	FolderStructure  string     `gorm:"type:varchar(50);not null;default:'7_DAILY_FOLDERS'" json:"folder_structure"` // 7_DAILY_FOLDERS, 4_WEEKLY_SUBFOLDERS, MONTHLY_FOLDERS, CUSTOM
	AutoPurgeAction  string     `gorm:"type:varchar(30);not null;default:'DELETE'" json:"auto_purge_action"`         // DELETE, ARCHIVE_COLD, COMPRESS_GZIP
	StorageProvider  string     `gorm:"type:varchar(30);not null;default:'LOCAL'" json:"storage_provider"`           // S3, GCS, MINIO, LOCAL
	BucketName       *string    `gorm:"type:varchar(150)" json:"bucket_name,omitempty"`
	CronSchedule     string     `gorm:"type:varchar(50);default:'0 1 * * *'" json:"cron_schedule"`
	ScheduleTimezone string     `gorm:"type:varchar(64);not null;default:'Asia/Bangkok'" json:"schedule_timezone"`
	IsActive         bool       `gorm:"not null;default:true;index" json:"is_active"`
	LastPurgeAt      *time.Time `gorm:"type:timestamptz" json:"last_purge_at,omitempty"`
	NextPurgeAt      *time.Time `gorm:"type:timestamptz" json:"next_purge_at,omitempty"`

	// Calendar-based retention fields. The legacy fields above remain for
	// backwards compatibility with the existing policy screen and worker.
	ActiveRetentionValue     int    `gorm:"not null;default:1" json:"active_retention_value"`
	ActiveRetentionUnit      string `gorm:"type:varchar(10);not null;default:'DAY'" json:"active_retention_unit"`
	ArchiveEnabled           bool   `gorm:"not null;default:false" json:"archive_enabled"`
	ArchiveAfterValue        *int   `json:"archive_after_value,omitempty"`
	ArchiveAfterUnit         string `gorm:"type:varchar(10)" json:"archive_after_unit,omitempty"`
	DeleteActiveAfterArchive bool   `gorm:"not null;default:false" json:"delete_active_after_archive"`
	ArchiveRetentionValue    *int   `json:"archive_retention_value,omitempty"`
	ArchiveRetentionUnit     string `gorm:"type:varchar(10)" json:"archive_retention_unit,omitempty"`
	ArchiveNeverDelete       bool   `gorm:"not null;default:false" json:"archive_never_delete"`
	ApplyToExistingLogs      bool   `gorm:"not null;default:false" json:"apply_to_existing_logs"`
	// EffectiveFrom freezes the lower timestamp boundary for policies that
	// must only process logs ingested after the policy was created. It is nil
	// when ApplyToExistingLogs is enabled.
	EffectiveFrom *time.Time `gorm:"type:timestamptz;index" json:"effective_from,omitempty"`
	Timestamps
}

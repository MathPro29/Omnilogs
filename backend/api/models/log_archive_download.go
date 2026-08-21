package models

import "time"

// LogArchiveDownload is a durable ZIP package created before expired archive
// files are removed. Keeping the package metadata lets the UI show a
// "Ready to load" list even after the original daily files are gone.
type LogArchiveDownload struct {
	DownloadID    string     `gorm:"type:uuid;primaryKey" json:"download_id"`
	ProductID     int        `gorm:"not null;index" json:"product_id"`
	EnvironmentID int        `gorm:"not null;index" json:"environment_id"`
	PolicyID      *int       `gorm:"index" json:"policy_id,omitempty"`
	DateFrom      *time.Time `gorm:"type:timestamptz" json:"date_from,omitempty"`
	DateTo        *time.Time `gorm:"type:timestamptz" json:"date_to,omitempty"`
	FileName      string     `gorm:"type:varchar(255);not null" json:"file_name"`
	FilePath      string     `gorm:"type:text;not null" json:"-"`
	SizeBytes     int64      `gorm:"not null;default:0" json:"size_bytes"`
	ArchiveCount  int        `gorm:"not null;default:0" json:"archive_count"`
	Status        string     `gorm:"type:varchar(30);not null;index" json:"status"`
	DownloadedAt  *time.Time `gorm:"type:timestamptz" json:"downloaded_at,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
}

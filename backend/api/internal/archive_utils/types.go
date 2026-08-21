package archive_utils

import "omnilogs-api/models"

type ArchiveData struct {
	Logs      []map[string]any        `json:"logs"`
	AuditLogs []models.SystemAuditLog `json:"audit_logs"`
}

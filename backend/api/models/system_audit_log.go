package models

import (
	"encoding/json"
	"time"
)

const (
	AuditResourceTypeMainLog = "MAIN_LOG"

	AuditActionViewLogDetail      = "VIEW_LOG_DETAIL"
	AuditActionSearchLogs         = "SEARCH_LOGS"
	AuditActionExportLogs         = "EXPORT_LOGS"
	AuditActionViewSensitiveField = "VIEW_SENSITIVE_FIELD"

	AuditResultSuccess = "SUCCESS"
	AuditResultFailed  = "FAILED"
	AuditResultDenied  = "DENIED"
)

type SystemAuditLog struct {
	AuditID      string          `gorm:"type:uuid;primaryKey" json:"audit_id"`
	ActorUserID  *int64          `gorm:"column:actor_user_id;index:idx_system_audit_logs_actor_created_at,priority:1" json:"actor_user_id,omitempty"`
	ProductID    *int64          `gorm:"column:product_id;index:idx_system_audit_logs_product_created_at,priority:1" json:"product_id,omitempty"`
	Action       string          `gorm:"not null;size:100;index:idx_system_audit_logs_action" json:"action"`
	ResourceType string          `gorm:"not null;size:50;index:idx_system_audit_logs_resource,priority:1" json:"resource_type"`
	ResourceID   *string         `gorm:"size:100;index:idx_system_audit_logs_resource,priority:2" json:"resource_id,omitempty"`
	RequestID    *string         `gorm:"size:100" json:"request_id,omitempty"`
	TraceID      *string         `gorm:"size:100" json:"trace_id,omitempty"`
	Method       *string         `gorm:"size:10" json:"method,omitempty"`
	Path         *string         `gorm:"type:text" json:"path,omitempty"`
	Result       string          `gorm:"not null;size:20;default:SUCCESS;index:idx_system_audit_logs_result" json:"result"`
	Metadata     json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"metadata"`
	IPAddress    *string         `gorm:"type:text" json:"ip_address,omitempty"`
	UserAgent    *string         `gorm:"type:text" json:"user_agent,omitempty"`
	CreatedAt    *time.Time      `gorm:"type:timestamptz;index:idx_system_audit_logs_created_at,sort:desc;index:idx_system_audit_logs_actor_created_at,priority:2,sort:desc;index:idx_system_audit_logs_product_created_at,priority:2,sort:desc" json:"created_at,omitempty"`
}

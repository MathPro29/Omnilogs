package dto

// LogQuery represents the incoming query parameters for log monitoring
type LogQuery struct {
	ProductID     int    `form:"product_id" binding:"required"`
	EnvironmentID int    `form:"environment_id"`
	LogLevel      string `form:"log_level"`
	Search        string `form:"search"`       // ข้อความที่ใช้ค้นหาใน payload
	StartTime     string `form:"start_time"`   // RFC3339 format
	EndTime       string `form:"end_time"`     // RFC3339 format
	ProjectIDs    string `form:"project_ids"`  // comma-separated, e.g. "1,2,3"
	CategoryIDs   string `form:"category_ids"` // comma-separated, e.g. "4,5,6"
	Limit         int    `form:"limit,default=50"`
	Offset        int    `form:"offset,default=0"`
}

// AuditLogQuery represents the incoming query parameters for audit logs
type AuditLogQuery struct {
	ProductID         *int  `form:"product_id"`
	AllowedProductIDs []int `form:"-" json:"-"`
	ProjectID         *int  `form:"project_id"`
	FeatureID         *int  `form:"feature_id"`
	Limit             int   `form:"limit,default=50"`
	Offset            int   `form:"offset,default=0"`
}

package audit

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var auditSpecialActionSuffixes = map[string]struct{}{
	"push-to-archives": {},
	"restore":          {},
	"review":           {},
	"reveal":           {},
}

var auditUserActivityPaths = map[string]struct{}{
	"/api/v1/auth/register":        {},
	"/api/v1/auth/login":           {},
	"/api/v1/auth/refresh-token":   {},
	"/api/v1/auth/logout":          {},
	"/api/v1/auth/forgot-password": {},
	"/api/v1/auth/reset-password":  {},
}

var auditExcludedPaths = map[string]struct{}{
	"/api/v1/queues/consume": {},
}

func shouldRecordAuditEvent(c *gin.Context) bool {
	fullPath := c.FullPath()
	if fullPath == "" {
		fullPath = c.Request.URL.Path
	}

	fullPath = strings.TrimSpace(fullPath)
	if fullPath == "" {
		return false
	}

	if strings.HasPrefix(fullPath, "/api/v1/audit-logs") || strings.HasPrefix(fullPath, "/api/v1/dashboard") || strings.HasPrefix(fullPath, "/api/v1/queues") {
		return false
	}

	// ตัด endpoint ภายในระบบที่เป็นงานประมวลผลออกจาก audit log เพื่อลด noise
	if _, ok := auditExcludedPaths[fullPath]; ok {
		return false
	}

	if _, ok := auditUserActivityPaths[fullPath]; ok {
		return true
	}

	switch strings.ToUpper(strings.TrimSpace(c.Request.Method)) {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	case http.MethodGet:
		parts := parseAuditPathParts(c)
		for _, part := range parts {
			if _, ok := auditSpecialActionSuffixes[strings.ToLower(strings.TrimSpace(part))]; ok {
				return true
			}
		}
	}

	return false
}

func parseAuditPathParts(c *gin.Context) []string {
	fullPath := c.FullPath()
	if fullPath == "" {
		fullPath = c.Request.URL.Path
	}

	trimmed := strings.Trim(fullPath, "/")
	if trimmed == "" {
		return nil
	}

	parts := strings.Split(trimmed, "/")
	if len(parts) >= 3 && parts[0] == "api" && parts[1] == "v1" {
		return parts[2:]
	}
	return parts
}

func isDynamicParam(part string) bool {
	part = strings.TrimSpace(part)
	return part == "" || strings.HasPrefix(part, "{") || strings.HasPrefix(part, ":")
}

func formatResource(part string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(part), "-", "_"))
}

func auditAction(c *gin.Context) string {
	parts := parseAuditPathParts(c)
	if len(parts) == 0 {
		return strings.ToUpper(c.Request.Method)
	}

	resource := "SYSTEM"
	for _, part := range parts {
		if isDynamicParam(part) {
			continue
		}
		resource = formatResource(part)
		break
	}

	target := "UNKNOWN"
	for i := len(parts) - 1; i >= 0; i-- {
		if isDynamicParam(parts[i]) {
			continue
		}
		target = formatResource(parts[i])
		break
	}

	action := methodAction(c.Request.Method)
	if target == resource {
		return resource + "." + action
	}
	return resource + "." + target + "." + action
}

func auditResource(c *gin.Context) (string, *string) {
	parts := parseAuditPathParts(c)
	if len(parts) == 0 {
		return "SYSTEM", nil
	}

	resourceType := "SYSTEM"
	for _, part := range parts {
		if isDynamicParam(part) {
			continue
		}
		resourceType = formatResource(part)
		break
	}

	for i := len(parts) - 1; i >= 0; i-- {
		part := strings.TrimSpace(parts[i])
		if isDynamicParam(part) {
			continue
		}
		if formatResource(part) != resourceType {
			return resourceType, &part
		}
	}

	return resourceType, nil
}

func methodAction(method string) string {
	switch action := strings.ToUpper(strings.TrimSpace(method)); action {
	case http.MethodGet:
		return "READ"
	case http.MethodPost:
		return "CREATE"
	case http.MethodPut, http.MethodPatch:
		return "UPDATE"
	case http.MethodDelete:
		return "DELETE"
	default:
		return action
	}
}

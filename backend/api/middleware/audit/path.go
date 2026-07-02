package audit

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

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

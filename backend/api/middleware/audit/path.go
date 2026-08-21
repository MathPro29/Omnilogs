package audit

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type auditRouteRule struct {
	method string
	path   string
}

// auditAllowedRoutes is intentionally opt-in. A new API route is not audited
// until its method and Gin route template are explicitly listed here.
var auditAllowedRoutes = []auditRouteRule{
	{http.MethodPost, "/api/v1/auth/register"},
	{http.MethodPost, "/api/v1/auth/login"},
	{http.MethodPost, "/api/v1/auth/refresh-token"},
	{http.MethodPost, "/api/v1/auth/logout"},
	{http.MethodPost, "/api/v1/auth/forgot-password"},
	{http.MethodPost, "/api/v1/auth/reset-password"},

	{http.MethodPut, "/api/v1/admin/give-admin"},
	{http.MethodPut, "/api/v1/admin/edit-user"},
	{http.MethodDelete, "/api/v1/admin/user/:id"},
	{http.MethodPost, "/api/v1/users"},
	{http.MethodPut, "/api/v1/users/:id"},
	{http.MethodDelete, "/api/v1/users/:id"},

	{http.MethodPost, "/api/v1/products"},
	{http.MethodPost, "/api/v1/products/setup"},
	{http.MethodPatch, "/api/v1/products/:productId"},
	{http.MethodDelete, "/api/v1/products/:productId"},
	{http.MethodPost, "/api/v1/products/:productId/restore"},
	{http.MethodDelete, "/api/v1/products/product/bulk-delete"},
	{http.MethodPatch, "/api/v1/products/:productId/setup/product"},
	{http.MethodPost, "/api/v1/products/:productId/setup/complete"},

	{http.MethodPost, "/api/v1/products/:productId/memberships"},
	{http.MethodPost, "/api/v1/products/:productId/memberships/bulk"},
	{http.MethodPatch, "/api/v1/products/:productId/memberships/:membershipId"},
	{http.MethodDelete, "/api/v1/products/:productId/memberships/:membershipId"},
	{http.MethodPut, "/api/v1/products/:productId/access/members/:userId"},
	{http.MethodPost, "/api/v1/products/:productId/memberships/:membershipId/scopes"},
	{http.MethodPatch, "/api/v1/products/:productId/memberships/:membershipId/scopes/:scopeId"},
	{http.MethodDelete, "/api/v1/products/:productId/memberships/:membershipId/scopes/:scopeId"},

	{http.MethodPost, "/api/v1/products/:productId/roles"},
	{http.MethodPatch, "/api/v1/products/:productId/roles/:roleId"},
	{http.MethodDelete, "/api/v1/products/:productId/roles/:roleId"},
	{http.MethodPost, "/api/v1/products/:productId/permission-rules"},
	{http.MethodPatch, "/api/v1/products/:productId/permission-rules/:ruleId"},

	{http.MethodPost, "/api/v1/products/:productId/projects"},
	{http.MethodPatch, "/api/v1/products/:productId/projects/:projectId"},
	{http.MethodDelete, "/api/v1/products/:productId/projects/:projectId"},
	{http.MethodPost, "/api/v1/products/:productId/projects/:projectId/features"},
	{http.MethodPatch, "/api/v1/products/:productId/projects/:projectId/features/:featureId"},
	{http.MethodDelete, "/api/v1/products/:productId/projects/:projectId/features/:featureId"},
	{http.MethodPost, "/api/v1/products/:productId/environments"},
	{http.MethodPatch, "/api/v1/products/:productId/environments/:environmentId"},
	{http.MethodDelete, "/api/v1/products/:productId/environments/:environmentId"},

	{http.MethodPost, "/api/v1/products/:productId/api-keys"},
	{http.MethodPatch, "/api/v1/products/:productId/api-keys/:keyId"},
	{http.MethodDelete, "/api/v1/products/:productId/api-keys/:keyId"},
	{http.MethodPost, "/api/v1/products/:productId/log-routes"},
	{http.MethodDelete, "/api/v1/products/:productId/log-routes/:routeId"},
	{http.MethodPost, "/api/v1/products/:productId/log-routing-rules"},
	{http.MethodDelete, "/api/v1/products/:productId/log-routing-rules/:ruleId"},

	{http.MethodPost, "/api/v1/products/:productId/elastic-index-policies"},
	{http.MethodPatch, "/api/v1/products/:productId/elastic-index-policies/:elasticPolicyId"},
	{http.MethodDelete, "/api/v1/products/:productId/elastic-index-policies/:elasticPolicyId"},
	{http.MethodPost, "/api/v1/products/:productId/elastic-index-policies/:elasticPolicyId/push-to-archives"},
	{http.MethodDelete, "/api/v1/products/:productId/elastic-index-policies/clear-logs"},
	{http.MethodPost, "/api/v1/products/:productId/log-archives/:archiveId/restore"},
	{http.MethodPost, "/api/v1/products/:productId/log-archives"},
	{http.MethodPost, "/api/v1/products/:productId/log-archives/:archiveId/verify"},
	{http.MethodGet, "/api/v1/products/:productId/log-archives/:archiveId"},
	{http.MethodPost, "/api/v1/products/:productId/retention-policies"},
	{http.MethodPut, "/api/v1/products/:productId/retention-policies/:policyId"},
	{http.MethodDelete, "/api/v1/products/:productId/retention-policies/:policyId"},
	{http.MethodPatch, "/api/v1/products/:productId/retention-policies/:policyId/toggle"},
	{http.MethodPost, "/api/v1/products/:productId/retention-policies/:policyId/run-now"},

	{http.MethodPost, "/api/v1/products/:productId/sensitive-logs/requests"},
	{http.MethodPost, "/api/v1/products/:productId/sensitive-logs/requests/:requestId/review"},
	{http.MethodPost, "/api/v1/products/:productId/sensitive-logs/reveal"},
	{http.MethodGet, "/api/v1/products/:productId/sensitive-logs/main-logs/:logId/raw"},

	{http.MethodPost, "/api/v1/audit-logs/:auditId/secrets/requests"},
	{http.MethodPost, "/api/v1/audit-logs/:auditId/secrets/requests/:requestId/review"},
	{http.MethodPost, "/api/v1/audit-logs/:auditId/secrets/reveal"},
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

	method := strings.ToUpper(strings.TrimSpace(c.Request.Method))
	for _, rule := range auditAllowedRoutes {
		if method == rule.method && auditPathMatches(rule.path, fullPath) {
			return true
		}
	}
	return false
}

func auditPathMatches(pattern, path string) bool {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(patternParts) != len(pathParts) {
		return false
	}
	for i := range patternParts {
		if strings.HasPrefix(patternParts[i], ":") {
			if strings.TrimSpace(pathParts[i]) == "" {
				return false
			}
			continue
		}
		if patternParts[i] != pathParts[i] {
			return false
		}
	}
	return true
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
	if value := strings.TrimSpace(c.GetString("audit_action")); value != "" {
		return value
	}

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
	if resourceType := strings.TrimSpace(c.GetString("audit_resource_type")); resourceType != "" {
		return resourceType, stringPointer(c.GetString("audit_resource_id"))
	}

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

package audit

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
)

var auditRedactedKeys = map[string]struct{}{
	"authorization":   {},
	"password":        {},
	"confirmpassword": {},
	"newpassword":     {},
	"refreshtoken":    {},
	"accesstoken":     {},
	"token":           {},
	"apikey":          {},
	"secret":          {},
	"clientsecret":    {},
	"privatekey":      {},
	"passphrase":      {},
	"credential":      {},
	"credentials":     {},
	"cookie":          {},
}

func normalizeKey(key string) string {
	k := strings.ToLower(strings.TrimSpace(key))
	k = strings.ReplaceAll(k, "_", "")
	k = strings.ReplaceAll(k, "-", "")
	return k
}

func isAuditRedactedKey(key string) bool {
	normKey := normalizeKey(key)
	_, ok := auditRedactedKeys[normKey]
	return ok
}

func buildAuditMetadata(c *gin.Context, requestBody []byte, statusCode int) json.RawMessage {
	metadata := map[string]any{
		"status_code": statusCode,
	}

	if reason := c.GetString("audit_reason"); strings.TrimSpace(reason) != "" {
		metadata["reason"] = reason
	}

	if query := c.Request.URL.Query(); len(query) > 0 {
		queryMap := make(map[string]any, len(query))
		for key, values := range query {
			if len(values) == 1 {
				queryMap[key] = values[0]
			} else {
				queryMap[key] = values
			}
		}
		var value any = queryMap
		redactAuditMetadata(&value)
		metadata["query"] = value
	}

	if len(requestBody) > 0 {
		var bodyValue any
		if json.Unmarshal(requestBody, &bodyValue) == nil {
			redactAuditMetadata(&bodyValue)
			metadata["request"] = bodyValue
		}
	}

	if val, exists := c.Get("audit_payload"); exists {
		if raw, err := json.Marshal(val); err == nil {
			var payload any
			if json.Unmarshal(raw, &payload) == nil {
				redactAuditMetadata(&payload)
				metadata["payload"] = payload
			}
		}
	}

	data, err := json.Marshal(metadata)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return data
}

func redactAuditMetadata(value *any) {
	switch typed := (*value).(type) {
	case map[string]any:
		for key, child := range typed {
			if isAuditRedactedKey(key) {
				typed[key] = "[REDACTED]"
				continue
			}
			redactAuditMetadata(&child)
			typed[key] = child
		}
	case []any:
		for i, child := range typed {
			redactAuditMetadata(&child)
			typed[i] = child
		}
	}
}

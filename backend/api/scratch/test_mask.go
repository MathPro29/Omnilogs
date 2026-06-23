package scratch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

var auditRedactedKeys = map[string]struct{}{
	"authorization":    {},
	"confirmPassword":  {},
	"confirm_password": {},
	"newPassword":      {},
	"new_password":     {},
	"password":         {},
	"refreshToken":     {},
	"refresh_token":    {},
	"token":            {},
}

func TestMask() {
	raw := []byte(`{"username":"testuser","password":"mysecretpassword","nested":{"token":"supertoken","normal":"val"}}`)
	sanitized := sanitizeJSONBytes(raw)
	fmt.Println("Sanitized JSON:", string(sanitized))
}

func sanitizeJSONBytes(raw []byte) []byte {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return nil
	}

	var payload any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil
	}

	sanitizeJSONValue(&payload)
	sanitized, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	return sanitized
}

func sanitizeJSONValue(value *any) {
	switch typed := (*value).(type) {
	case map[string]any:
		for key, child := range typed {
			if _, ok := auditRedactedKeys[key]; ok || isAuditRedactedKey(key) {
				typed[key] = "[REDACTED]"
				continue
			}
			sanitizeJSONValue(&child)
			typed[key] = child
		}
	case []any:
		for i := range typed {
			child := typed[i]
			sanitizeJSONValue(&child)
			typed[i] = child
		}
	}
}

func isAuditRedactedKey(key string) bool {
	_, ok := auditRedactedKeys[strings.TrimSpace(key)]
	if ok {
		return true
	}
	_, ok = auditRedactedKeys[strings.ToLower(strings.TrimSpace(key))]
	return ok
}

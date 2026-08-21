package audit

import (
	"encoding/json"
	"fmt"
	"strings"

	"omnilogs-api/models"
	"omnilogs-api/utils"

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
	"resettoken":      {},
}

type auditMetadataResult struct {
	Metadata json.RawMessage
	Secrets  []models.AuditSecret
}

type auditSecretRef struct {
	SecretID      string `json:"secret_id"`
	FieldKey      string `json:"field_key"`
	FieldPath     string `json:"field_path"`
	SourceSection string `json:"source_section"`
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

func buildAuditMetadata(c *gin.Context, requestBody []byte, statusCode int, auditID string, encryptionKey string) auditMetadataResult {
	metadata := map[string]any{
		"status_code": statusCode,
	}

	var secrets []models.AuditSecret
	var refs []auditSecretRef

	if reason := c.GetString("audit_reason"); strings.TrimSpace(reason) != "" {
		metadata["reason"] = reason
	}

	if headerMap := buildSensitiveHeaderMap(c); len(headerMap) > 0 {
		value, headerSecrets, headerRefs := redactAuditValue(headerMap, "headers", "headers", auditID, auditUserID(c), encryptionKey)
		metadata["headers"] = value
		secrets = append(secrets, headerSecrets...)
		refs = append(refs, headerRefs...)
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
		value, querySecrets, queryRefs := redactAuditValue(queryMap, "query", "query", auditID, auditUserID(c), encryptionKey)
		metadata["query"] = value
		secrets = append(secrets, querySecrets...)
		refs = append(refs, queryRefs...)
	}

	if len(requestBody) > 0 {
		var bodyValue any
		if json.Unmarshal(requestBody, &bodyValue) == nil {
			value, requestSecrets, requestRefs := redactAuditValue(bodyValue, "request", "request", auditID, auditUserID(c), encryptionKey)
			metadata["request"] = value
			secrets = append(secrets, requestSecrets...)
			refs = append(refs, requestRefs...)
		}
	}

	if val, exists := c.Get("audit_payload"); exists {
		if raw, err := json.Marshal(val); err == nil {
			var payload any
			if json.Unmarshal(raw, &payload) == nil {
				value, payloadSecrets, payloadRefs := redactAuditValue(payload, "payload", "payload", auditID, auditUserID(c), encryptionKey)
				metadata["payload"] = value
				secrets = append(secrets, payloadSecrets...)
				refs = append(refs, payloadRefs...)
			}
		}
	}

	if len(refs) > 0 {
		metadata["secret_refs"] = refs
	}

	data, err := json.Marshal(metadata)
	if err != nil {
		return auditMetadataResult{Metadata: json.RawMessage(`{}`), Secrets: secrets}
	}
	return auditMetadataResult{Metadata: data, Secrets: secrets}
}

func buildSensitiveHeaderMap(c *gin.Context) map[string]any {
	headers := map[string]any{}
	for _, key := range []string{"Authorization", "Cookie", "X-API-Key"} {
		if value := strings.TrimSpace(c.GetHeader(key)); value != "" {
			headers[key] = value
		}
	}
	return headers
}

func redactAuditValue(value any, fieldKey string, fieldPath string, auditID string, actorUserID *int64, encryptionKey string) (any, []models.AuditSecret, []auditSecretRef) {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		var secrets []models.AuditSecret
		var refs []auditSecretRef
		for key, child := range typed {
			childPath := joinAuditPath(fieldPath, key)
			if isAuditRedactedKey(key) {
				secret, ref, masked := newAuditSecretRecord(auditID, actorUserID, key, childPath, fieldKey, child, encryptionKey)
				result[key] = masked
				secrets = append(secrets, secret)
				refs = append(refs, ref)
				continue
			}
			redactedChild, childSecrets, childRefs := redactAuditValue(child, key, childPath, auditID, actorUserID, encryptionKey)
			result[key] = redactedChild
			secrets = append(secrets, childSecrets...)
			refs = append(refs, childRefs...)
		}
		return result, secrets, refs
	case []any:
		result := make([]any, len(typed))
		var secrets []models.AuditSecret
		var refs []auditSecretRef
		for i, child := range typed {
			childPath := fmt.Sprintf("%s[%d]", fieldPath, i)
			redactedChild, childSecrets, childRefs := redactAuditValue(child, fieldKey, childPath, auditID, actorUserID, encryptionKey)
			result[i] = redactedChild
			secrets = append(secrets, childSecrets...)
			refs = append(refs, childRefs...)
		}
		return result, secrets, refs
	default:
		return value, nil, nil
	}
}

func joinAuditPath(parent string, key string) string {
	if parent == "" {
		return key
	}
	return parent + "." + key
}

func newAuditSecretRecord(auditID string, actorUserID *int64, fieldKey string, fieldPath string, sourceSection string, value any, encryptionKey string) (models.AuditSecret, auditSecretRef, string) {
	plaintext := stringifyAuditSecretValue(value)
	encryptedValue, err := utils.EncryptAESGCM(plaintext, []byte(encryptionKey))
	if err != nil {
		encryptedValue = ""
	}
	algorithm := "AES-GCM"
	secretID := newAuditUUID()
	secret := models.AuditSecret{
		SecretID:            secretID,
		AuditID:             auditID,
		ActorUserID:         actorUserID,
		FieldKey:            fieldKey,
		FieldPath:           fieldPath,
		SourceSection:       sourceSection,
		EncryptedValue:      encryptedValue,
		EncryptionAlgorithm: &algorithm,
		RequiresApproval:    true,
	}
	ref := auditSecretRef{
		SecretID:      secretID,
		FieldKey:      fieldKey,
		FieldPath:     fieldPath,
		SourceSection: sourceSection,
	}
	return secret, ref, "[REDACTED]"
}

func stringifyAuditSecretValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		raw, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprintf("%v", value)
		}
		return string(raw)
	}
}

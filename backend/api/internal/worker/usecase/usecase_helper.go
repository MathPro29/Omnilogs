package usecase

import (
	"fmt"
	"strings"

	"omnilogs-api/models"
	"omnilogs-api/utils"
)

func (u *usecase) maskPayloadAndExtractSecrets(
	data map[string]any,
	parentPath string,
	productID int,
	logID string,
	fields []models.LogFieldDefinition,
	rules []models.LogMaskingRule,
	matchers *sensitiveMatchers,
) ([]models.LogSensitiveFieldSecret, error) {
	var secrets []models.LogSensitiveFieldSecret

	for key, value := range data {
		currentPath := key
		if parentPath != "" {
			currentPath = parentPath + "." + key
		}

		switch valTyped := value.(type) {
		case map[string]any:
			subSecrets, err := u.maskPayloadAndExtractSecrets(valTyped, currentPath, productID, logID, fields, rules, matchers)
			if err != nil {
				return nil, err
			}
			secrets = append(secrets, subSecrets...)
		case []any:
			for i, arrayVal := range valTyped {
				if arrayMap, ok := arrayVal.(map[string]any); ok {
					subSecrets, err := u.maskPayloadAndExtractSecrets(arrayMap, fmt.Sprintf("%s[%d]", currentPath, i), productID, logID, fields, rules, matchers)
					if err != nil {
						return nil, err
					}
					secrets = append(secrets, subSecrets...)
				}
			}
		default:
			isSensitive := false
			var fieldDef *models.LogFieldDefinition

			if matchers != nil {
				if matchedField := matchers.matchField(key, currentPath); matchedField != nil {
					isSensitive = true
					fieldDef = matchedField
				}
			} else {
				for i := range fields {
					if fields[i].FieldKey == key || (fields[i].FieldPath != nil && *fields[i].FieldPath == currentPath) {
						isSensitive = true
						fieldDef = &fields[i]
						break
					}
				}
			}

			var matchedRule *models.LogMaskingRule
			if matchers != nil {
				if rule := matchers.matchRule(key, currentPath); rule != nil {
					isSensitive = true
					matchedRule = rule
				}
			} else {
				for i := range rules {
					if (rules[i].FieldKey != nil && *rules[i].FieldKey == key) || (rules[i].FieldPath != nil && *rules[i].FieldPath == currentPath) {
						isSensitive = true
						matchedRule = &rules[i]
						break
					}
				}
			}

			if isSensitive {
				strVal := fmt.Sprintf("%v", value)
				encVal, err := utils.EncryptAESGCM(strVal, []byte(u.encryptionKey))
				if err != nil {
					return nil, err
				}

				secretID := newUUID()
				requiresApproval := true

				secretRecord := models.LogSensitiveFieldSecret{
					SecretID:         secretID,
					LogID:            logID,
					ProductID:        productID,
					FieldKey:         key,
					FieldPath:        currentPath,
					SourceSection:    "payload",
					EncryptedValue:   encVal,
					RequiresApproval: requiresApproval,
				}
				if fieldDef != nil {
					secretRecord.FieldDefinitionID = &fieldDef.FieldDefinitionID
				}
				secrets = append(secrets, secretRecord)

				maskValue := "[REDACTED]"
				if matchedRule != nil && matchedRule.MaskValue != nil {
					maskValue = *matchedRule.MaskValue
				}
				data[key] = maskValue
			}
		}
	}

	return secrets, nil
}

func buildSensitiveMatchers(fields []models.LogFieldDefinition, rules []models.LogMaskingRule) *sensitiveMatchers {
	matchers := &sensitiveMatchers{
		fieldByKey:  make(map[string]*models.LogFieldDefinition, len(fields)),
		fieldByPath: make(map[string]*models.LogFieldDefinition, len(fields)),
		ruleByKey:   make(map[string]*models.LogMaskingRule, len(rules)),
		ruleByPath:  make(map[string]*models.LogMaskingRule, len(rules)),
	}

	for i := range fields {
		field := &fields[i]
		if field.FieldKey != "" {
			matchers.fieldByKey[field.FieldKey] = field
		}
		if field.FieldPath != nil && *field.FieldPath != "" {
			matchers.fieldByPath[*field.FieldPath] = field
		}
	}

	for i := range rules {
		rule := &rules[i]
		if rule.FieldKey != nil && *rule.FieldKey != "" {
			matchers.ruleByKey[*rule.FieldKey] = rule
		}
		if rule.FieldPath != nil && *rule.FieldPath != "" {
			matchers.ruleByPath[*rule.FieldPath] = rule
		}
	}

	return matchers
}

func (m *sensitiveMatchers) matchField(key string, path string) *models.LogFieldDefinition {
	if m == nil {
		return nil
	}
	if field := m.fieldByPath[path]; field != nil {
		return field
	}
	if field := m.fieldByPath[normalizeArrayPath(path)]; field != nil {
		return field
	}
	return m.fieldByKey[key]
}

func (m *sensitiveMatchers) matchRule(key string, path string) *models.LogMaskingRule {
	if m == nil {
		return nil
	}
	if rule := m.ruleByPath[path]; rule != nil {
		return rule
	}
	if rule := m.ruleByPath[normalizeArrayPath(path)]; rule != nil {
		return rule
	}
	return m.ruleByKey[key]
}

func normalizeArrayPath(path string) string {
	var builder strings.Builder
	for i := 0; i < len(path); i++ {
		if path[i] != '[' {
			builder.WriteByte(path[i])
			continue
		}
		end := strings.IndexByte(path[i:], ']')
		if end <= 1 {
			builder.WriteByte(path[i])
			continue
		}
		index := path[i+1 : i+end]
		allDigits := true
		for _, char := range index {
			if char < '0' || char > '9' {
				allDigits = false
				break
			}
		}
		if allDigits {
			builder.WriteString("[]")
			i += end
		} else {
			builder.WriteString(path[i : i+end+1])
			i += end
		}
	}
	return builder.String()
}

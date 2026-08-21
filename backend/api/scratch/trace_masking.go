//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"omnilogs-api/configs"
	"omnilogs-api/models"
)

func main() {
	os.Setenv("ENV_PATH", "../.env")
	env := configs.LoadEnv()
	db, err := configs.ConnectDB(env)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	db = db.Debug()

	productID := 1

	// Load fields and rules
	var fields []models.LogFieldDefinition
	if err := db.Where("product_id = ? AND is_sensitive = TRUE AND is_active = TRUE", productID).Find(&fields).Error; err != nil {
		log.Fatalf("failed to find fields: %v", err)
	}

	var rules []models.LogMaskingRule
	if err := db.Where("product_id = ? AND is_active = TRUE", productID).Find(&rules).Error; err != nil {
		log.Fatalf("failed to find rules: %v", err)
	}

	fmt.Printf("Loaded %d fields and %d rules for product %d\n", len(fields), len(rules), productID)

	payloadJSON := `{
		"log_level": "INFO",
		"event_type": "sensitive-data-storage-test",
		"message": "Raw customer and contact data should be masked in Main Logs",
		"customer_name": "Somchai Jaidee",
		"customer_address": "99 Sukhumvit Road, Bangkok 10110",
		"customer_email": "customer.test@example.com",
		"callback_url": "https://example.com/orders/OMNI-1001"
	}`

	var payload map[string]any
	if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
		log.Fatalf("failed to parse payload: %v", err)
	}

	matchers := buildSensitiveMatchers(fields, rules)

	secrets, err := maskPayloadAndExtractSecrets(payload, "", productID, "test-log-id", fields, rules, matchers)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	fmt.Printf("Secrets extracted: %d\n", len(secrets))
	for _, s := range secrets {
		fmt.Printf("Secret - Key: %s, Path: %s\n", s.FieldKey, s.FieldPath)
	}

	maskedJSON, _ := json.MarshalIndent(payload, "", "  ")
	fmt.Println("=== Masked Payload ===")
	fmt.Println(string(maskedJSON))
}

type sensitiveMatchers struct {
	fieldByKey  map[string]*models.LogFieldDefinition
	fieldByPath map[string]*models.LogFieldDefinition
	ruleByKey   map[string]*models.LogMaskingRule
	ruleByPath  map[string]*models.LogMaskingRule
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
	return m.fieldByKey[key]
}

func (m *sensitiveMatchers) matchRule(key string, path string) *models.LogMaskingRule {
	if m == nil {
		return nil
	}
	if rule := m.ruleByPath[path]; rule != nil {
		return rule
	}
	return m.ruleByKey[key]
}

func maskPayloadAndExtractSecrets(
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
			subSecrets, err := maskPayloadAndExtractSecrets(valTyped, currentPath, productID, logID, fields, rules, matchers)
			if err != nil {
				return nil, err
			}
			secrets = append(secrets, subSecrets...)
		default:
			isSensitive := false
			var fieldDef *models.LogFieldDefinition

			if matchers != nil {
				if matchedField := matchers.matchField(key, currentPath); matchedField != nil {
					isSensitive = true
					fieldDef = matchedField
				}
			}

			var matchedRule *models.LogMaskingRule
			if matchers != nil {
				if rule := matchers.matchRule(key, currentPath); rule != nil {
					isSensitive = true
					matchedRule = rule
				}
			}

			if isSensitive {
				secretRecord := models.LogSensitiveFieldSecret{
					LogID:            logID,
					ProductID:        productID,
					FieldKey:         key,
					FieldPath:        currentPath,
					SourceSection:    "payload",
					EncryptedValue:   "encrypted",
					RequiresApproval: true,
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

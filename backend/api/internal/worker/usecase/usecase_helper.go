package usecase

import (
	"fmt"

	"omnilogs-api/models"
	"omnilogs-api/utils"
)

// maskPayloadAndExtractSecrets ทำการสแกนหาข้อมูลส่วนบุคคล (Sensitive Data) หรือข้อมูลลับใน JSON Payload
// โดยจะทำงานแบบ Recursive (เรียกตัวเองซ้ำ) เพื่อเจาะลึกลงไปใน Object หรือ Array ที่ซ้อนทับกัน
// หากพบฟิลด์ที่ตรงกับเงื่อนไข จะทำการ 1) เข้ารหัส (Encrypt) เก็บไว้ใน DB และ 2) แทนที่ค่าเดิมด้วยตัวอักษรทับ (Masking)
func (u *usecase) maskPayloadAndExtractSecrets(
	data map[string]any,
	parentPath string,
	productID int,
	logID string,
	fields []models.LogFieldDefinition,
	rules []models.LogMaskingRule,
) ([]models.LogSensitiveFieldSecret, error) {
	var secrets []models.LogSensitiveFieldSecret

	for key, value := range data {
		// สร้าง Path แบบ Dot-notation เช่น "user.email" หรือ "user.address.zipcode"
		// เพื่อใช้เปรียบเทียบกับเงื่อนไขการตั้งค่า
		currentPath := key
		if parentPath != "" {
			currentPath = parentPath + "." + key
		}

		switch valTyped := value.(type) {
		case map[string]any:
			// ถ้าเจอ Object (map) ซ้อนอยู่ข้างใน ให้เรียกฟังก์ชันนี้ซ้ำเพื่อเจาะลึกลงไปอีก
			subSecrets, err := u.maskPayloadAndExtractSecrets(valTyped, currentPath, productID, logID, fields, rules)
			if err != nil {
				return nil, err
			}
			secrets = append(secrets, subSecrets...)
		case []any:
			// ถ้าเจอ Array ให้วนลูปหา Object ที่อยู่ใน Array
			// ตัวอย่าง Path จะได้เป็น "users[0]", "users[1]"
			for i, arrayVal := range valTyped {
				if arrayMap, ok := arrayVal.(map[string]any); ok {
					subSecrets, err := u.maskPayloadAndExtractSecrets(arrayMap, fmt.Sprintf("%s[%d]", currentPath, i), productID, logID, fields, rules)
					if err != nil {
						return nil, err
					}
					secrets = append(secrets, subSecrets...)
				}
			}
		default:
			// สำหรับชนิดข้อมูลพื้นฐาน (String, Number, Boolean)
			isSensitive := false
			var fieldDef *models.LogFieldDefinition
			
			// 1. ตรวจสอบว่าตรงกับรายชื่อฟิลด์สำคัญ (LogFieldDefinition) หรือไม่
			for i := range fields {
				if fields[i].FieldKey == key || (fields[i].FieldPath != nil && *fields[i].FieldPath == currentPath) {
					isSensitive = true
					fieldDef = &fields[i]
					break
				}
			}

			// 2. ตรวจสอบว่าตรงกับกฎการเซ็นเซอร์ (LogMaskingRule) หรือไม่
			var matchedRule *models.LogMaskingRule
			for i := range rules {
				if (rules[i].FieldKey != nil && *rules[i].FieldKey == key) || (rules[i].FieldPath != nil && *rules[i].FieldPath == currentPath) {
					isSensitive = true
					matchedRule = &rules[i]
					break
				}
			}

			if isSensitive {
				// แปลงค่าดั้งเดิมให้เป็น String เพื่อเตรียมเข้ารหัส
				strVal := fmt.Sprintf("%v", value)

				// นำค่าจริงไปเข้ารหัส (Encrypt) ด้วยกุญแจเข้ารหัส AES-GCM ของระบบ
				encVal, err := utils.EncryptAESGCM(strVal, []byte(u.encryptionKey))
				if err != nil {
					return nil, err
				}

				var secretID = newUUID()
				var requiresApproval = true // ต้องขออนุมัติก่อนดูค่าความลับเสมอ

				// สร้าง Record เตรียมนำไปบันทึกลงตาราง LogSensitiveFieldSecret
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

				// ดำเนินการ Masking (แทนที่) ค่าเดิมใน Payload เพื่อส่งต่อให้ Elasticsearch
				// ทำให้ ES ได้รับข้อมูลที่ถูกเบลอแล้วแทนที่ข้อมูลจริง
				maskValue := "[REDACTED]"
				if matchedRule != nil && matchedRule.MaskValue != nil {
					maskValue = *matchedRule.MaskValue // ใช้ข้อความเบลอแบบ Custom ถ้ามีการตั้งไว้
				}
				data[key] = maskValue
			}
		}
	}

	return secrets, nil
}

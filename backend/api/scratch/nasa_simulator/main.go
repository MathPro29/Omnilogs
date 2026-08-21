package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"omnilogs-api/configs"
	"omnilogs-api/models"
	"omnilogs-api/utils"
)

var statuses = []string{"INFO", "WARN", "ERROR", "PANIC", "DEBUG"}

func main() {
	fmt.Println("🚀 เริ่มรันตัวจำลอง NASA Simulator...")

	// โหลดคอนฟิกเพื่อดึงฐานข้อมูลมาค้นหา ID ล่าสุด
	env := configs.LoadEnv()
	db, err := configs.ConnectDB(env)
	if err != nil {
		log.Fatalf("❌ เชื่อมต่อฐานข้อมูลล้มเหลว: %v", err)
	}

	// 1. ดึง Product
	var product models.Product
	if err := db.Where("product_name = ?", "NASA").First(&product).Error; err != nil {
		fmt.Println("⚠️ ไม่พบ Product 'NASA', กำลังใช้สิทธิ์เลือก Product แรกในฐานข้อมูล...")
		if err := db.First(&product).Error; err != nil {
			log.Fatalf("❌ ไม่พบ Product ใด ๆ ในฐานข้อมูล: %v", err)
		}
	}

	// 2. ดึง Environment
	var envItem models.ProductEnvironment
	if err := db.Where("product_id = ?", product.ProductID).First(&envItem).Error; err != nil {
		log.Fatalf("❌ ไม่พบ Environment ของ Product ID %d: %v", product.ProductID, err)
	}

	// 3. ดึง Project
	var project models.Project
	if err := db.Where("product_id = ? AND project_name = ?", product.ProductID, "Apollo11").First(&project).Error; err != nil {
		fmt.Println("⚠️ ไม่พบ Project 'Apollo11', กำลังใช้สิทธิ์เลือก Project แรกใน Product นี้...")
		if err := db.Where("product_id = ?", product.ProductID).First(&project).Error; err != nil {
			log.Fatalf("❌ ไม่พบ Project ใด ๆ ใน Product นี้: %v", err)
		}
	}

	// 4. ดึง Features
	var features []models.ProjectFeature
	if err := db.Where("project_id = ?", project.ProjectID).Find(&features).Error; err != nil {
		log.Printf("⚠️ ดึง Features (Categories) ล้มเหลว: %v", err)
	}

	// โหลด Custom Fields ที่ใช้งานสำหรับ Product นี้
	var activeFields []models.LogFieldDefinition
	if err := db.Where("(product_id = ? OR product_id IS NULL) AND is_active = TRUE", product.ProductID).Find(&activeFields).Error; err != nil {
		log.Printf("⚠️ ไม่สามารถโหลด Custom Fields: %v", err)
	}

	// ดึง Enum Options เตรียมไว้
	fieldEnums := make(map[int][]string)
	for _, field := range activeFields {
		if field.DataType == "enum" {
			var opts []models.LogFieldEnumOption
			if err := db.Where("field_definition_id = ? AND is_active = TRUE", field.FieldDefinitionID).Find(&opts).Error; err == nil {
				for _, opt := range opts {
					fieldEnums[field.FieldDefinitionID] = append(fieldEnums[field.FieldDefinitionID], opt.OptionValue)
				}
			}
		}
	}

	fmt.Printf("🛰️ ข้อมูลเป้าหมาย => Product: %s (ID %d), Env: %s (ID %d), Project: %s (ID %d), Features: %d รายการ, Custom Fields: %d รายการ\n",
		product.ProductName, product.ProductID, envItem.EnvironmentName, envItem.EnvironmentID, project.ProjectName, project.ProjectID, len(features), len(activeFields))

	// 5. ดึง User คนแรกเพื่อไป Gen Token
	var user models.User
	if err := db.First(&user).Error; err != nil {
		log.Fatalf("❌ ไม่พบ User: %v", err)
	}

	// สร้าง JWT Token ดึงจาก utils ของโปรเจกต์
	token, err := utils.GenerateToken(uint(user.UserID), user.Email, "ADMIN", 1, "access", env.JWTSecret, 24*time.Hour)
	if err != nil {
		log.Fatalf("❌ สร้าง Token ล้มเหลว: %v", err)
	}

	var wg sync.WaitGroup
	totalBatches := 100 // ยิงพร้อมกัน 100 Logs

	for i := 1; i <= totalBatches; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// สุ่มเลือก Sub-category / Feature
			var categoryID *int
			featureName := "Unknown"
			if len(features) > 0 {
				feat := features[rand.Intn(len(features))]
				categoryID = &feat.CategoryID
				featureName = feat.CategoryName
			}

			// สุ่มสถานะและข้อความให้ตรงกับ NASA Theme หรือทั่วไป
			level := statuses[rand.Intn(len(statuses))]
			msg := generateNASAStatusMessage(featureName, level)

			// จำลองข้อมูลสำหรับ Custom Fields
			customFields := make(map[string]interface{})
			for _, field := range activeFields {
				var val interface{}
				switch field.DataType {
				case "enum":
					enums := fieldEnums[field.FieldDefinitionID]
					if len(enums) > 0 {
						val = enums[rand.Intn(len(enums))]
					} else {
						val = "default_enum_val"
					}
				case "string":
					val = fmt.Sprintf("simulated-%s-%d", field.FieldKey, rand.Intn(100))
				case "integer", "int":
					val = rand.Intn(1000)
				case "number", "float":
					val = rand.Float64() * 100.0
				case "boolean", "bool":
					val = rand.Intn(2) == 1
				case "datetime", "date":
					val = time.Now().Add(time.Duration(-rand.Intn(24)) * time.Hour).Format(time.RFC3339)
				case "json":
					val = map[string]interface{}{"status": "ok", "value": rand.Intn(100)}
				case "object":
					val = map[string]interface{}{"nested_key": "nested_val"}
				default:
					val = "simulated-value"
				}
				customFields[field.FieldKey] = val
			}

			// 1. สร้าง JSON ก้อนเล็ก (input_payload)
			inputPayload := map[string]interface{}{
				"log_level":     level,
				"message":       fmt.Sprintf("[Log %03d] %s", workerID, msg),
				"timestamp":     time.Now().Format(time.RFC3339Nano),
				"project_id":    project.ProjectID,
				"custom_fields": customFields,
			}
			if categoryID != nil {
				inputPayload["category_id"] = *categoryID
			}
			payloadBytes, _ := json.Marshal(inputPayload)

			// 2. สร้าง JSON ก้อนใหญ่
			batchReq := map[string]interface{}{
				"product_id":      product.ProductID,
				"environment_id":  envItem.EnvironmentID,
				"queue_key":       fmt.Sprintf("%s-key", project.ProjectCode),
				"source_type":     "application",
				"source_platform": "SIMULATOR",
				"priority":        1,
				"logs": []map[string]interface{}{
					{
						"sequence_no":     1,
						"source_type":     "application",
						"source_platform": "SIMULATOR",
						"input_payload":   json.RawMessage(payloadBytes),
					},
				},
			}
			reqBytes, _ := json.Marshal(batchReq)

			// 3. ยิง Request เข้าหา /api/v1/queues พร้อมแนบ Token
			req, _ := http.NewRequest("POST", "http://localhost:2910/api/v1/queues", bytes.NewBuffer(reqBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("❌ ยิง Log %d ล้มเหลว: %v\n", workerID, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == 201 {
				fmt.Printf("✅ ยิง Log สำเร็จ (Worker %03d) -> หมวดหมู่: %-15s สถานะ: %s\n", workerID, featureName, level)
			} else {
				fmt.Printf("⚠️ ยิง Log (Worker %03d) ถูกปฏิเสธ - HTTP %d\n", workerID, resp.StatusCode)
			}
		}(i)

		// หน่วงเวลาเล็กน้อยเพื่อให้จำลองเหมือนเหตุการณ์จริง
		time.Sleep(10 * time.Millisecond)
	}

	wg.Wait()
	fmt.Println("🎉 การทดสอบระบบจำลองส่ง Logs เสร็จสิ้น!")
}

// สร้างข้อความจำลองเหตุการณ์
func generateNASAStatusMessage(featureName, level string) string {
	if level == "PANIC" || level == "ERROR" {
		switch featureName {
		case "Head_logs":
			return "ความดันส่วนหัวลดลงกะทันหัน!"
		case "Wings_logs":
			return "ปีกซ้ายได้รับความเสียหายจากสะเก็ดดาว"
		case "Passenger_logs":
			return "ออกซิเจนในห้องโดยสารต่ำกว่าเกณฑ์"
		case "Engine_logs":
			return "เครื่องยนต์หลักดับกะทันหัน"
		case "Radar_logs":
			return "เรดาร์สูญเสียสัญญาณจากศูนย์ฮิวสตัน"
		case "Temp_logs":
			return "อุณหภูมิไอพ่นร้อนจัด (Overheat)"
		case "Compass_logs":
			return "ระบบนำทางเข็มทิศรวน"
		default:
			return "ระบบเกิดข้อผิดพลาดร้ายแรง!"
		}
	} else if level == "WARN" {
		return "พบความผิดปกติเล็กน้อย แนะนำให้ตรวจสอบ"
	}

	// INFO / DEBUG
	switch featureName {
	case "Head_logs":
		return "เซ็นเซอร์ส่วนหัวรายงานสถานะปกติ"
	case "Wings_logs":
		return "ปีกซ้ายขวากางออกทำมุม 100% เรียบร้อย"
	case "Passenger_logs":
		return "ลูกเรือรัดเข็มขัดนิรภัยพร้อมเข้าสู่วงโคจร"
	case "Engine_logs":
		return "ระดับเชื้อเพลิงไฮโดรเจนเหลวคงที่"
	case "Radar_logs":
		return "เส้นทางวงโคจรโล่ง ไร้วัตถุรบกวน"
	case "Temp_logs":
		return "ระบบควบคุมอุณหภูมิรักษาความเย็นได้ดี"
	case "Compass_logs":
		return "ล็อกเป้าดวงจันทร์สำเร็จ"
	default:
		return "ทุกระบบทำงานปกติ"
	}
}

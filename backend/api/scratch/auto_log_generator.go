//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"omnilogs-api/configs"
	"omnilogs-api/models"
	"omnilogs-api/utils"
)

func main() {
	fmt.Println("🚀 Starting Omnilogs Auto Log Generator...")

	env := configs.LoadEnv()
	db, err := configs.ConnectDB(env)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// 1. Find or create Product
	var product models.Product
	if err := db.First(&product).Error; err != nil {
		product = models.Product{
			ProductName: "Omnilogs Platform",
			ProductCode: "OMNI",
			IsActive:    true,
		}
		if err := db.Create(&product).Error; err != nil {
			log.Fatalf("❌ Failed to create product: %v", err)
		}
		fmt.Printf("✅ Created default Product: %s (ID %d)\n", product.ProductName, product.ProductID)
	}

	// 2. Find or create Environment
	var envItem models.ProductEnvironment
	if err := db.Where("product_id = ?", product.ProductID).First(&envItem).Error; err != nil {
		envItem = models.ProductEnvironment{
			ProductID:       product.ProductID,
			EnvironmentName: "Production",
			EnvironmentCode: "prod",
		}
		if err := db.Create(&envItem).Error; err != nil {
			log.Fatalf("❌ Failed to create environment: %v", err)
		}
		fmt.Printf("✅ Created default Environment: %s (ID %d)\n", envItem.EnvironmentName, envItem.EnvironmentID)
	}

	// 3. Find or create Project
	var project models.Project
	if err := db.Where("product_id = ?", product.ProductID).First(&project).Error; err != nil {
		project = models.Project{
			ProductID:   product.ProductID,
			ProjectName: "Core API Service",
			ProjectCode: "core-api",
			IsActive:    true,
		}
		if err := db.Create(&project).Error; err != nil {
			log.Fatalf("❌ Failed to create project: %v", err)
		}
		fmt.Printf("✅ Created default Project: %s (ID %d)\n", project.ProjectName, project.ProjectID)
	}

	// 4. Find user & generate token
	var user models.User
	if err := db.First(&user).Error; err != nil {
		log.Fatalf("❌ No user found in database: %v", err)
	}

	token, err := utils.GenerateToken(uint(user.UserID), user.Email, "ADMIN", 1, "access", env.JWTSecret, 24*time.Hour)
	if err != nil {
		log.Fatalf("❌ Failed to generate token: %v", err)
	}

	levels := []string{"INFO", "INFO", "INFO", "WARN", "ERROR", "DEBUG", "PANIC"}
	messages := []string{
		"User authentication succeeded",
		"Database connection pool initialized",
		"Cache miss for key user_session:1002",
		"HTTP 500 Internal Server Error in /api/v1/checkout",
		"High memory usage detected on worker node #3",
		"Payment gateway response timeout after 5000ms",
		"Elasticsearch index document indexing completed",
		"Panic recovered: nil pointer dereference in user_service.go:42",
		"Scheduled cron job log_retention_purge finished in 1.2s",
		"API rate limit exceeded for client IP 192.168.1.105",
	}

	fmt.Println("📡 Ingesting 50 sample logs into backend queue...")

	successCount := 0
	for i := 1; i <= 50; i++ {
		level := levels[rand.Intn(len(levels))]
		msg := messages[rand.Intn(len(messages))]
		ts := time.Now().Add(time.Duration(-rand.Intn(30)) * time.Minute).Format(time.RFC3339Nano)

		inputPayload := map[string]interface{}{
			"log_level":   level,
			"message":     fmt.Sprintf("[AutoLog #%02d] %s", i, msg),
			"timestamp":   ts,
			"project_id":  project.ProjectID,
			"trace_id":    fmt.Sprintf("trace-%d-%d", time.Now().Unix(), i),
			"request_id":  fmt.Sprintf("req-%d-%d", time.Now().Unix(), i),
			"service":     "api-gateway",
			"environment": envItem.EnvironmentCode,
		}

		payloadBytes, _ := json.Marshal(inputPayload)

		batchReq := map[string]interface{}{
			"product_id":      product.ProductID,
			"environment_id":  envItem.EnvironmentID,
			"queue_key":       fmt.Sprintf("%s-key", project.ProjectCode),
			"source_type":     "application",
			"source_platform": "AUTO_GENERATOR",
			"priority":        1,
			"logs": []map[string]interface{}{
				{
					"sequence_no":     1,
					"source_type":     "application",
					"source_platform": "AUTO_GENERATOR",
					"input_payload":   json.RawMessage(payloadBytes),
				},
			},
		}

		reqBytes, _ := json.Marshal(batchReq)
		req, _ := http.NewRequest("POST", "http://localhost:2910/api/v1/queues", bytes.NewBuffer(reqBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err == nil {
			if resp.StatusCode == 201 || resp.StatusCode == 200 {
				successCount++
			}
			resp.Body.Close()
		}
		time.Sleep(20 * time.Millisecond)
	}

	fmt.Printf("🎉 Successfully generated and sent %d / 50 logs to Omnilogs Queue!\n", successCount)
}

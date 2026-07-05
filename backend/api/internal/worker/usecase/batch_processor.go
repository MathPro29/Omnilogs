package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"omnilogs-api/models"
	"omnilogs-api/utils"
)

// ตัวรับ Batch Process Task
func (u *usecase) processClaimedBatch(ctx context.Context, batch *models.LogQueueBatch) error {

	slog.Info("claimed batch",
		slog.String("worker_id", u.workerID),
		slog.String("batch_id", batch.BatchID),
		slog.Int("total_logs", batch.TotalLogs),
	)

	// โหลดรายการ Log ย่อย (Items) ทั้งหมดที่อยู่ใน Batch นี้เพื่อนำมาประมวลผล
	items, err := u.repo.LoadPendingItems(ctx, batch.BatchID)
	if err != nil {
		return u.failBatch(ctx, batch, err.Error())
	}
	if len(items) == 0 {
		// ถ้าไม่มี Item ให้ประมวลผลเลย ก็จบ Batch ทันที
		return u.repo.FinishBatch(ctx, batch.BatchID, "COMPLETED", nil)
	}

	var processedCount int
	var retryCount int
	var failedCount int

	// วนลูปประมวลผลทีละ Item
	for i := range items {
		if err := u.processItem(ctx, batch, &items[i]); err != nil {
			// ตรวจสอบว่า Error เป็นการบอกให้รอ Retry หรือเป็นการเฟลถาวร
			if strings.Contains(err.Error(), "retry scheduled") {
				retryCount++
			} else {
				failedCount++
			}
			continue // ไปทำ Item ถัดไป
		}
		processedCount++ // นับจำนวนที่ทำสำเร็จ
	}

	// สรุปสถานะของ Batch หลังจากทำครบทุก Item
	status := "COMPLETED"
	switch {
	case processedCount == 0 && retryCount > 0 && failedCount == 0:
		status = "RETRY_PENDING" // ต้องรอประมวลผลใหม่ทั้งหมด
	case processedCount == 0 && failedCount > 0:
		status = "FAILED" // เฟลถาวรทั้งหมด
	case retryCount > 0 || failedCount > 0:
		status = "PARTIAL" // สำเร็จบางส่วน
	}

	// บันทึกสถานะกลับลง Database
	message := fmt.Sprintf("processed=%d retry_pending=%d failed=%d", processedCount, retryCount, failedCount)
	return u.repo.FinishBatch(ctx, batch.BatchID, status, &message)
}

// processItem คือหัวใจหลักของ Worker ที่ทำหน้าที่ประมวลผล Log 1 รายการ
// ขั้นตอน: ล้างข้อมูลเก่า -> ตรวจสอบข้อมูล -> Validate โครงสร้าง -> ทำ Data Masking -> บันทึกเข้า Elasticsearch -> สร้าง Index Reference ลง DB
func (u *usecase) processItem(ctx context.Context, batch *models.LogQueueBatch, item *models.LogQueueItem) error {
	// ส่วนที่ 1: ทำเครื่องหมายว่า item นี้กำลังถูก worker ตัวนี้ประมวลผล
	now := time.Now()
	if err := u.repo.MarkItemProcessing(ctx, item.QueueItemID, u.workerID, now); err != nil {
		return err
	}
	item.WorkerID = stringPtr(u.workerID)
	item.ProcessingAttempts++
	item.ProcessingStartedAt = &now

	if batch.ProductID == nil || batch.EnvironmentID == nil {
		// Worker ส่ง Log เข้า Elasticsearch ไม่ได้จนกว่าจะรู้ Product และ Environment
		return u.handleItemFailure(ctx, batch, item, "VALIDATION", "BATCH_METADATA_MISSING", "product_id and environment_id are required before indexing")
	}

	// ส่วนที่ 2: โหลดกฎ Sensitive Field และ Masking ของ Product นี้
	fields, _ := u.repo.GetSensitiveFieldDefinitions(ctx, *batch.ProductID)
	rules, _ := u.repo.GetLogMaskingRules(ctx, *batch.ProductID)

	// ส่วนที่ 3: แปลง Raw JSON เป็น Map เพื่ออ่าน Hierarchy และทำ Masking
	var payload map[string]any
	if err := json.Unmarshal(item.InputPayload, &payload); err != nil {
		return u.handleItemFailure(ctx, batch, item, "TRANSFORM", "INVALID_PAYLOAD", err.Error())
	}

	// === จำลองการทำงาน==
	if msg, ok := payload["message"].(string); ok {
		switch msg {
		case "FORCE_RETRY":
			// จำลองการ Fail เพื่อรอ Retry เสมอ
			return u.handleItemFailure(ctx, batch, item, "TEST", "SIMULATED_RETRY_ERROR", "simulated temporary failure for retry")

		case "FORCE_RETRY_THEN_SUCCESS":
			// ถ้าเป็นครั้งแรก (RetryCount = 0) ให้เฟลเพื่อส่งไป Retry
			// แต่ถ้ามีการดึงขึ้นมาประมวลผลใหม่รอบสอง (RetryCount > 0) ให้ปล่อยผ่านเพื่อสำเร็จ (Success)
			if item.RetryCount == 0 {
				return u.handleItemFailure(ctx, batch, item, "TEST", "TEMPORARY_FAILURE", "simulated temporary failure (will succeed on retry)")
			}

		case "FORCE_FAIL":
			// จำลองการเฟลถาวร (ถ้าระบบดึงไป retry ครบกำหนดแล้ว ก็จะย้ายลงตาราง Failed log และถูกลบตามที่เราเขียนในส่วนแรก)
			return u.handleItemFailure(ctx, batch, item, "TEST", "SIMULATED_PERMANENT_ERROR", "simulated permanent failure")
		}
	}
	// ===============================================

	docID := newUUID()
	originalPayloadBytes := append([]byte(nil), item.InputPayload...)

	// ส่วนที่ 4: ตรวจ Hierarchy ก่อนเกิด Side Effect ใด ๆ
	// ถ้าความสัมพันธ์ผิด จะบันทึก Log Failure และหยุดก่อนเก็บ Secret หรือส่ง Elasticsearch
	_, meta, err := buildElasticDocument(batch, item)
	if err != nil {
		return u.handleItemFailure(ctx, batch, item, "VALIDATION", "INVALID_PAYLOAD", err.Error())
	}
	if err := u.validateLogHierarchy(ctx, *batch.ProductID, *batch.EnvironmentID, meta); err != nil {
		return u.handleItemFailure(ctx, batch, item, "VALIDATION", "INVALID_LOG_HIERARCHY", err.Error())
	}

	// ส่วนที่ 5: Mask ข้อมูลสำคัญแบบ Recursive และรวบรวมค่าจริงเป็น Secret
	secrets, err := u.maskPayloadAndExtractSecrets(payload, "", *batch.ProductID, docID, fields, rules)
	if err != nil {
		return u.handleItemFailure(ctx, batch, item, "TRANSFORM", "MASKING_FAILED", err.Error())
	}

	// ส่วนที่ 6: บันทึก Secret ลง PostgreSQL เพื่อให้ Reveal ได้เฉพาะผู้มีสิทธิ์
	for _, secret := range secrets {
		if err := u.repo.CreateSensitiveFieldSecret(ctx, &secret); err != nil {
			return u.handleItemFailure(ctx, batch, item, "DATABASE", "SENSITIVE_SECRET_CREATE_FAILED", err.Error())
		}
	}

	// ส่วนที่ 7: สร้าง Document จาก Payload ที่ Mask แล้ว
	// Elasticsearch จึงไม่ได้รับข้อมูลลับต้นฉบับ
	maskedPayloadBytes, err := json.Marshal(payload)
	if err != nil {
		return u.handleItemFailure(ctx, batch, item, "TRANSFORM", "PAYLOAD_MARSHAL_FAILED", err.Error())
	}
	item.InputPayload = maskedPayloadBytes

	document, meta, err := buildElasticDocument(batch, item)
	if err != nil {
		return u.handleItemFailure(ctx, batch, item, "TRANSFORM", "INVALID_PAYLOAD", err.Error())
	}

	// ส่วนที่ 8: เลือก Index ตาม Product, Environment และเวลา แล้วส่งเข้า Elasticsearch
	indexName := u.resolveIndexName(ctx, *batch.ProductID, batch.EnvironmentID, meta.Timestamp)
	indexedAt := time.Now()

	statusCode, syncErr := u.indexDocument(ctx, indexName, docID, document)
	if syncErr != nil {
		return u.handleItemFailure(ctx, batch, item, "ELASTICSEARCH", "INDEX_REQUEST_FAILED", syncErr.Error())
	}

	// ส่วนที่ 9: สร้าง Reference เชื่อม Record ใน PostgreSQL กับ Document ใน Elasticsearch
	indexRef := models.LogIndexRef{
		LogID:              docID,
		ProductID:          *batch.ProductID,
		ProjectID:          meta.ProjectID,
		CategoryID:         meta.CategoryID,
		EnvironmentID:      *batch.EnvironmentID,
		SourceID:           batch.SourceID,
		BatchID:            &batch.BatchID,
		QueueItemID:        &item.QueueItemID,
		ResponseStatusCode: &statusCode,
		DurationMs:         meta.DurationMs,
		LogLevel:           meta.LogLevel,
		EventType:          meta.EventType,
		SourceRequestID:    meta.SourceRequestID,
		CorrelationID:      meta.CorrelationID,
		TraceID:            meta.TraceID,
		SpanID:             meta.SpanID,
		ParentSpanID:       meta.ParentSpanID,
		RequestMethod:      meta.RequestMethod,
		RequestPath:        meta.RequestPath,
		RoutePattern:       meta.RoutePattern,
		FeatureFullPath:    meta.FeatureFullPath,
		FeaturePathIDs:     meta.FeaturePathIDs,
		Timestamp:          meta.Timestamp,
		IngestedAt:         &indexedAt,
		ElasticIndex:       indexName,
		ElasticDocumentID:  docID,
		IndexStatus:        "INDEXED",
		IndexedAt:          &indexedAt,
		LastSyncAt:         &indexedAt,
	}
	if err := u.repo.UpsertIndexRef(ctx, &indexRef); err != nil {
		return u.handleItemFailure(ctx, batch, item, "DATABASE", "INDEX_REF_CREATE_FAILED", err.Error())
	}
	if err := u.storeOriginalPayload(ctx, docID, *batch.ProductID, batch.EnvironmentID, originalPayloadBytes, indexedAt); err != nil {
		return u.handleItemFailure(ctx, batch, item, "DATABASE", "PAYLOAD_ARCHIVE_CREATE_FAILED", err.Error())
	}

	// เมื่อบันทึก Reference สำเร็จ จึงลบ Queue Item ออกจากตารางคิว
return u.repo.DeleteQueueItem(ctx, item.QueueItemID)
}

func (u *usecase) storeOriginalPayload(ctx context.Context, logID string, productID int, environmentID *int, payload []byte, indexedAt time.Time) error {
	encryptedPayload, err := utils.EncryptAESGCM(string(payload), []byte(u.encryptionKey))
	if err != nil {
		return err
	}

	// Retention settings
	var retentionUntil time.Time 
	policy, err:= u.repo.FindIndexPolicy(ctx, productID, environmentID)
	if err == nil && policy != nil && policy.RetentionDays != nil {
		retentionUntil = time.Now().AddDate(0, 0, *policy.RetentionDays)
	} else {
		retentionUntil = time.Now().AddDate(1, 0, 0)
	}

	sizeBytes := int64(len(payload))
	checksum := utils.SHA256Hex(string(payload))
	fileFormat := "JSON"
	keyRef := "DATA_ENCRYPTION_KEY"
	algorithm := "AES-256-GCM"
	objectPath := fmt.Sprintf("postgres://log-payloads/%s", logID)

	ref := &models.LogObjectStorageRef{
		ObjectRefID:         newUUID(),
		LogID:               logID,
		ProductID:           productID,
		EnvironmentID:       environmentID,
		StorageProvider:     "POSTGRES",
		ObjectPath:          objectPath,
		ObjectType:          "INPUT_PAYLOAD",
		FileFormat:          &fileFormat,
		SizeBytes:           &sizeBytes,
		Checksum:            &checksum,
		EncryptedPayload:    &encryptedPayload,
		EncryptionKeyRef:    &keyRef,
		EncryptionAlgorithm: &algorithm,
		IsEncrypted:         true,
		RetentionUntil:      &retentionUntil,
	}

	return u.repo.UpsertObjectStorageRef(ctx, ref)
}

package usecase

import (
	"encoding/json"
	"testing"

	"omnilogs-api/internal/queue"
	"omnilogs-api/models"
	"omnilogs-api/utils"
)

func TestRawPayloadArchiveKeepsOriginalValueBeforeMasking(t *testing.T) {
	const encryptionKey = "12345678901234567890123456789012"
	productID := 1
	environmentID := 2
	original := []byte(`{"customer_name":"Somchai Jaidee","customer_email":"customer.test@example.com"}`)

	var payload map[string]any
	if err := json.Unmarshal(original, &payload); err != nil {
		t.Fatal(err)
	}

	u := &usecase{encryptionKey: encryptionKey}
	fields := []models.LogFieldDefinition{{FieldKey: "customer_name"}}
	matchers := buildSensitiveMatchers(fields, nil)
	if _, err := u.maskPayloadAndExtractSecrets(payload, "", productID, "log-1", fields, nil, matchers); err != nil {
		t.Fatal(err)
	}
	if payload["customer_name"] != "[REDACTED]" {
		t.Fatalf("indexed payload customer_name = %#v", payload["customer_name"])
	}

	ref, err := u.buildObjectStorageRef(queue.LogMessage{
		LogID:         "log-1",
		ProductID:     &productID,
		EnvironmentID: &environmentID,
	}, original)
	if err != nil {
		t.Fatal(err)
	}
	if ref.EncryptedPayload == nil {
		t.Fatal("encrypted raw payload was not created")
	}

	decrypted, err := utils.DecryptAESGCM(*ref.EncryptedPayload, []byte(encryptionKey))
	if err != nil {
		t.Fatal(err)
	}
	if decrypted != string(original) {
		t.Fatalf("raw payload = %s, want %s", decrypted, original)
	}
}

package usecase

import "testing"

func TestIsDefaultSensitiveKey(t *testing.T) {
	for _, key := range []string{"Authorization", "x-api-key", "password", "refresh_token", "Private.Key"} {
		if !isDefaultSensitiveKey(key) {
			t.Fatalf("%q must be sensitive", key)
		}
	}
	if isDefaultSensitiveKey("request_id") {
		t.Fatal("request_id must not be treated as sensitive")
	}
}

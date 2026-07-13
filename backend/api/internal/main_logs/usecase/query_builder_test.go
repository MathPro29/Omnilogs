package usecase

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildSearchQueryNormalizesCustomFieldPath(t *testing.T) {
	path := "payload.custom_fields.customer_tier"
	value := "PREMIUM"
	query := buildSearchQuery(SearchInput{
		ProductID:        1,
		CustomFieldPath:  &path,
		CustomFieldValue: &value,
		Page:             1,
		PerPage:          20,
	})

	encoded, err := json.Marshal(query)
	if err != nil {
		t.Fatalf("marshal query: %v", err)
	}
	text := string(encoded)
	if !strings.Contains(text, `payload.custom_fields.customer_tier.keyword`) {
		t.Fatalf("expected normalized custom field path, got %s", text)
	}
	if strings.Contains(text, `payload.payload.`) {
		t.Fatalf("custom field path was prefixed twice: %s", text)
	}
}

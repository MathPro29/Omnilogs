package handler

import (
	"strings"
	"testing"
)

func TestValidateFavoriteReorderRejectsDuplicates(t *testing.T) {
	if err := validateFavoriteReorder([]favoriteOrder{
		{FieldDefinitionID: 1, DisplayOrder: 0},
		{FieldDefinitionID: 1, DisplayOrder: 1},
	}); err == nil || !strings.Contains(err.Error(), "duplicate favorite field") {
		t.Fatalf("expected duplicate field error, got %v", err)
	}
	if err := validateFavoriteReorder([]favoriteOrder{
		{FieldDefinitionID: 1, DisplayOrder: 0},
		{FieldDefinitionID: 2, DisplayOrder: 0},
	}); err == nil || !strings.Contains(err.Error(), "duplicate display_order") {
		t.Fatalf("expected duplicate order error, got %v", err)
	}
}

func TestValidateFavoriteReorderAcceptsOrderedItems(t *testing.T) {
	if err := validateFavoriteReorder([]favoriteOrder{
		{FieldDefinitionID: 2, DisplayOrder: 0},
		{FieldDefinitionID: 1, DisplayOrder: 1},
	}); err != nil {
		t.Fatalf("unexpected reorder validation error: %v", err)
	}
}

func TestNextAvailableFieldKeyUsesFirstGap(t *testing.T) {
	key := nextAvailableFieldKey("customer_tier", []string{
		"customer_tier", "customer_tier_2", "customer_tier_4", "other_field",
	})
	if key != "customer_tier_3" {
		t.Fatalf("expected first available suffix, got %s", key)
	}
}

func TestNextAvailableFieldKeyKeepsUnusedBase(t *testing.T) {
	if key := nextAvailableFieldKey("customer_tier", []string{"other_field"}); key != "customer_tier" {
		t.Fatalf("expected unused base key, got %s", key)
	}
}

func TestNormalizeFavoriteFieldPathRemovesPayloadEnvelope(t *testing.T) {
	if got := normalizeFavoriteFieldPath("raw.payload.request.user.email"); got != "request.user.email" {
		t.Fatalf("unexpected canonical path: %s", got)
	}
	if got := normalizeFavoriteFieldPath("payload"); got != "$" {
		t.Fatalf("unexpected payload root path: %s", got)
	}
}

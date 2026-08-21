package provider

import (
	"testing"
	"time"
)

func TestIndexMatchesRangeIncludesLegacyRollbackIndex(t *testing.T) {
	from := time.Date(2026, time.August, 14, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, 1)
	legacy := "omnilogs-product-77-env-137-search-v3-rollback-f66cdb54c5b2463a96adf97b0bec6b1d"
	if !indexMatchesRange(legacy, &from, &to) {
		t.Fatal("legacy rollback index must remain visible to re-archive and safe delete queries")
	}
	if !indexMatchesRange("omnilogs-product-77-env-137-search-v3-2026.08.14", &from, &to) {
		t.Fatal("daily index in range was excluded")
	}
	if indexMatchesRange("omnilogs-product-77-env-137-search-v3-2026.08.10", &from, &to) {
		t.Fatal("daily index outside range was included")
	}
	if indexMatchesRange("omnilogs-product-77-env-137-search-v3-2026.08.15", &from, &to) {
		t.Fatal("exclusive date_to index was included")
	}
}

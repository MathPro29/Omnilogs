package usecase

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildSearchQueryKeepsProductIsolationAndPagination(t *testing.T) {
	projectIDs := []int64{10, 20}
	query := buildSearchQuery(SearchInput{ProductID: 7, ProjectIDs: projectIDs, Page: 3, PerPage: 25})
	if query["from"] != 50 || query["size"] != 25 {
		t.Fatalf("unexpected pagination: %#v", query)
	}

	queryNode := query["query"].(map[string]any)
	boolNode := queryNode["bool"].(map[string]any)
	filters := boolNode["filter"].([]map[string]any)
	productTerm := filters[0]["term"].(map[string]any)
	if productTerm["product_id"] != int64(7) {
		t.Fatalf("product isolation filter is missing: %#v", filters)
	}
	if len(filters) != 2 {
		t.Fatalf("expected product and multi-project filters, got %#v", filters)
	}
}

func TestBuildSearchQueryPrefersSingleProjectFilter(t *testing.T) {
	projectID := int64(10)
	query := buildSearchQuery(SearchInput{ProductID: 7, ProjectID: &projectID, ProjectIDs: []int64{20}, Page: 1, PerPage: 20})
	filters := query["query"].(map[string]any)["bool"].(map[string]any)["filter"].([]map[string]any)
	if len(filters) != 2 {
		t.Fatalf("multi-project filter must not be added with project_id: %#v", filters)
	}
}

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

func TestNormalizeCustomFieldPathSupportsArrayNotation(t *testing.T) {
	path, ok := normalizeCustomFieldPath("payload.orders[].items[].sku")
	if !ok {
		t.Fatal("expected array path to be valid")
	}
	if path != "payload.orders.items.sku" {
		t.Fatalf("unexpected normalized path: %s", path)
	}
}

func TestBuildSearchQueryRejectsUnsafeCustomFieldPath(t *testing.T) {
	path := `custom_fields.name);DELETE`
	value := "unsafe"
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
	if strings.Contains(string(encoded), "DELETE") {
		t.Fatalf("unsafe field path leaked into Elasticsearch query: %s", encoded)
	}
}

func TestValidateSearchInputRejectsUnsafeOrIncompleteCustomFilter(t *testing.T) {
	unsafePath := `custom_fields.name);DELETE`
	value := "unsafe"
	if err := validateSearchInput(SearchInput{CustomFieldPath: &unsafePath, CustomFieldValue: &value}); err == nil {
		t.Fatal("expected unsafe custom field path to fail validation")
	}

	validPath := "custom_fields.customer_tier"
	if err := validateSearchInput(SearchInput{CustomFieldPath: &validPath}); err == nil {
		t.Fatal("expected incomplete custom field filter to fail validation")
	}
}

func TestNormalizeSearchInputBoundsAndDeduplicates(t *testing.T) {
	input := normalizeSearchInput(SearchInput{
		Page:       -1,
		PerPage:    1000,
		ProjectIDs: []int64{2, 2, -1, 3},
	})
	if input.Page != 1 || input.PerPage != 100 {
		t.Fatalf("unexpected pagination: page=%d per_page=%d", input.Page, input.PerPage)
	}
	if len(input.ProjectIDs) != 2 || input.ProjectIDs[0] != 2 || input.ProjectIDs[1] != 3 {
		t.Fatalf("unexpected normalized ids: %#v", input.ProjectIDs)
	}
}

func BenchmarkBuildSearchQuery(b *testing.B) {
	keyword := "checkout timeout"
	path := "payload.custom_fields.customer_tier"
	value := "PREMIUM"
	input := SearchInput{
		ProductID:        1,
		ProjectIDs:       []int64{1, 2, 3},
		CategoryIDs:      []int64{10, 20, 30},
		CustomFieldPath:  &path,
		CustomFieldValue: &value,
		Keyword:          &keyword,
		Page:             1,
		PerPage:          20,
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = buildSearchQuery(input)
	}
}

func TestDynamicDataPathAndSort(t *testing.T) {
	path := "data.order.total"
	normalized, ok := normalizeCustomFieldPath(path)
	if !ok || normalized != path {
		t.Fatalf("normalized dynamic path = %q, ok=%v", normalized, ok)
	}
	order := "desc"
	query := buildSearchQuery(SearchInput{
		ProductID: 1, Page: 1, PerPage: 20,
		SortField: &path, SortOrder: order,
	})
	sortItems := query["sort"].([]any)
	first := sortItems[0].(map[string]any)
	if _, exists := first[path]; !exists {
		t.Fatalf("dynamic sort does not contain %s: %#v", path, first)
	}
}

func TestDynamicSortRejectsUnsafePath(t *testing.T) {
	path := "data.user.*"
	if err := validateSearchInput(SearchInput{SortField: &path}); err == nil {
		t.Fatal("expected unsafe dynamic sort path to be rejected")
	}
}

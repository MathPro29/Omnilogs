package usecase

import "testing"

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

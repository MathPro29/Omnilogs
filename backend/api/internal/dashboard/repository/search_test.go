package repository

import (
	"strconv"
	"strings"
	"testing"

	"omnilogs-api/dto"
)

func TestBuildElasticsearchQueryUsesFilterContext(t *testing.T) {
	query := buildElasticsearchQuery(dto.LogQuery{
		ProductID:     7,
		EnvironmentID: 2,
		Search:        "timeout",
		Limit:         20,
	})

	queryNode := query["query"].(map[string]any)
	boolNode := queryNode["bool"].(map[string]any)
	filters := boolNode["filter"].([]map[string]any)
	must := boolNode["must"].([]map[string]any)
	if len(filters) != 2 {
		t.Fatalf("expected product and environment filters, got %d", len(filters))
	}
	if len(must) != 1 {
		t.Fatalf("expected keyword query in must context, got %d", len(must))
	}
}

func TestParseCSVInt64DeduplicatesAndCapsInput(t *testing.T) {
	parts := make([]string, 0, 120)
	for i := 1; i <= 120; i++ {
		parts = append(parts, strconv.Itoa(i))
	}
	parts = append(parts, "1", "invalid", "-2")

	values := parseCSVInt64(strings.Join(parts, ","))
	if len(values) != maxFilterIDs {
		t.Fatalf("expected %d ids, got %d", maxFilterIDs, len(values))
	}
	if values[0] != 1 || values[len(values)-1] != 100 {
		t.Fatalf("unexpected bounded values: first=%d last=%d", values[0], values[len(values)-1])
	}
}

func BenchmarkBuildElasticsearchQuery(b *testing.B) {
	query := dto.LogQuery{
		ProductID:   7,
		ProjectIDs:  "1,2,3",
		CategoryIDs: "10,20,30",
		Search:      "checkout timeout",
		Limit:       50,
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = buildElasticsearchQuery(query)
	}
}

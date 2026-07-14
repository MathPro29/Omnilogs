package migrate

import (
	"strings"
	"testing"
)

func TestQueryPerformanceIndexesAreConcurrentAndPartial(t *testing.T) {
	if len(queryPerformanceIndexStatements) != 2 {
		t.Fatalf("expected two targeted indexes, got %d", len(queryPerformanceIndexStatements))
	}
	for _, statement := range queryPerformanceIndexStatements {
		upper := strings.ToUpper(statement)
		if !strings.Contains(upper, "CONCURRENTLY") {
			t.Fatalf("index must be created concurrently: %s", statement)
		}
		if !strings.Contains(upper, "WHERE") {
			t.Fatalf("write-heavy queue index must be partial: %s", statement)
		}
	}
}

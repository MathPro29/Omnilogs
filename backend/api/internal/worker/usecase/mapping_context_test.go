package usecase

import "testing"

func TestMappingStatusAndSource(t *testing.T) {
	if got := mappingSource("CLIENT_EXPLICIT"); got != "explicit" {
		t.Fatalf("explicit source = %q", got)
	}
	if got := mappingSource("ROUTING_RULE"); got != "common_rule" {
		t.Fatalf("common source = %q", got)
	}
	if got := mappingStatus("CLASSIFIED", "ROUTING_RULE"); got != "mapped" {
		t.Fatalf("mapped status = %q", got)
	}
	if got := mappingStatus("UNCLASSIFIED", "ROUTING_CONFLICT"); got != "ambiguous" {
		t.Fatalf("ambiguous status = %q", got)
	}
}

package utils

import "testing"

func TestNormalizeName(t *testing.T) {
	for _, input := range []string{"Home", "home", " HOME", "Home "} {
		if got := NormalizeName(input); got != "home" {
			t.Fatalf("NormalizeName(%q) = %q, want home", input, got)
		}
	}
}

func TestNormalizeCode(t *testing.T) {
	if got := NormalizeCode(" Order.Create "); got != "order.create" {
		t.Fatalf("NormalizeCode() = %q", got)
	}
}

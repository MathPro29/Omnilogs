package utils

import "testing"

func TestCategoryScopeAllowsParent(t *testing.T) {
	path := "1,4,9"
	tests := []struct {
		name       string
		requested  int
		parentRead bool
		want       bool
	}{
		{name: "root parent", requested: 1, parentRead: true, want: true},
		{name: "direct parent", requested: 4, parentRead: true, want: true},
		{name: "selected category", requested: 9, parentRead: true, want: true},
		{name: "sibling stays denied", requested: 8, parentRead: true, want: false},
		{name: "write does not inherit upward", requested: 4, parentRead: false, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := CategoryScopeAllowsParent(&path, test.requested, test.parentRead); got != test.want {
				t.Fatalf("CategoryScopeAllowsParent() = %v, want %v", got, test.want)
			}
		})
	}
}

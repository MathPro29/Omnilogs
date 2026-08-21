package usecase

import "testing"

func TestRoleRequiresProductScope(t *testing.T) {
	tests := []struct {
		role string
		want bool
	}{
		{role: "owner", want: true},
		{role: "OWNER", want: true},
		{role: "admin", want: false},
		{role: "viewer", want: false},
	}
	for _, test := range tests {
		if got := roleRequiresProductScope(test.role); got != test.want {
			t.Fatalf("roleRequiresProductScope(%q) = %v, want %v", test.role, got, test.want)
		}
	}
}

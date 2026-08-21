package usecase

import (
	"testing"

	"omnilogs-api/models"
)

func TestPermissionListAllows(t *testing.T) {
	tests := []struct {
		name string
		raw  []models.ProductRolePermission
		want bool
	}{
		{name: "matching permission", raw: []models.ProductRolePermission{{ResourceType: "PROJECT", Action: "UPDATE"}}, want: true},
		{name: "case insensitive", raw: []models.ProductRolePermission{{ResourceType: "project", Action: "update"}}, want: true},
		{name: "denied action", raw: []models.ProductRolePermission{{ResourceType: "PROJECT", Action: "READ"}}, want: false},
		{name: "empty", raw: nil, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := permissionListAllows(test.raw, "PROJECT", "UPDATE"); got != test.want {
				t.Fatalf("permissionListAllows() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestPathContains(t *testing.T) {
	path := "1,12,25"
	if !pathContains(&path, 12) {
		t.Fatal("expected category 12 in path")
	}
	if pathContains(&path, 2) {
		t.Fatal("category matching must not use partial IDs")
	}
	if pathContains(nil, 1) {
		t.Fatal("nil path must not match")
	}
}

func TestMembershipReadAllowed(t *testing.T) {
	if !membershipReadAllowed("feature") {
		t.Fatal("feature metadata should be readable by a scoped member")
	}
	if membershipReadAllowed("API_KEY") || membershipReadAllowed("ACCESS") {
		t.Fatal("sensitive administration resources require role permissions")
	}
}

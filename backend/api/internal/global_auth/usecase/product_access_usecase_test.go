package usecase

import (
	"omnilogs-api/models"
	"testing"
)

func TestRoleAccessLevel(t *testing.T) {
	tests := []struct {
		name, code  string
		permissions []models.ProductRolePermission
		want        string
	}{
		{name: "owner", code: "owner", want: "FULL_ACCESS"},
		{name: "reader", code: "viewer", permissions: []models.ProductRolePermission{{Action: "READ"}}, want: "READ_ONLY"},
		{name: "editor", code: "editor", permissions: []models.ProductRolePermission{{Action: "UPDATE"}}, want: "EDITOR"},
		{name: "admin", code: "admin", permissions: []models.ProductRolePermission{{Action: "GRANT"}}, want: "ADMIN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := roleAccessLevel(tt.code, tt.permissions); got != tt.want {
				t.Fatalf("got %s want %s", got, tt.want)
			}
		})
	}
}

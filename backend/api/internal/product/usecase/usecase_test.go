package usecase

import (
	"encoding/json"
	"testing"
)

func TestPermissionJSONAllows(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{name: "all", raw: `{"all":true}`, want: true},
		{name: "action list", raw: `{"PROJECT":["READ","UPDATE"]}`, want: true},
		{name: "action map", raw: `{"project":{"update":true}}`, want: true},
		{name: "flat permission", raw: `{"PROJECT.UPDATE":true}`, want: true},
		{name: "denied action", raw: `{"PROJECT":["READ"]}`, want: false},
		{name: "malformed", raw: `{`, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := permissionJSONAllows(json.RawMessage(test.raw), "PROJECT", "UPDATE"); got != test.want {
				t.Fatalf("permissionJSONAllows() = %v, want %v", got, test.want)
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

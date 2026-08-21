package objectstore

import "testing"

func TestR2KeyUsesPortableForwardSlashesAndPrefix(t *testing.T) {
	store := &R2{prefix: "omnilogs/backups"}
	got := store.Key(`archives\product-7\2026\08\16.ndjson.gz`)
	want := "omnilogs/backups/archives/product-7/2026/08/16.ndjson.gz"
	if got != want {
		t.Fatalf("key = %q, want %q", got, want)
	}
}

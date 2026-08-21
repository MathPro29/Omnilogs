package repository

import (
	"os"
	"path/filepath"
	"testing"
)

func TestActualFileUsageCountsPhysicalFilesAndDeduplicatesPaths(t *testing.T) {
	directory := t.TempDir()
	first := filepath.Join(directory, "first.ndjson.gz")
	second := filepath.Join(directory, "ready.zip")
	if err := os.WriteFile(first, []byte("12345"), 0600); err != nil {
		t.Fatalf("write first file: %v", err)
	}
	if err := os.WriteFile(second, []byte("1234567"), 0600); err != nil {
		t.Fatalf("write second file: %v", err)
	}
	bytes, files := actualFileUsage([]string{first, first, second, filepath.Join(directory, "missing")})
	if bytes != 12 || files != 2 {
		t.Fatalf("unexpected actual usage: bytes=%d files=%d", bytes, files)
	}
}

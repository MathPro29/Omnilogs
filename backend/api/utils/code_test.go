package utils

import "testing"

func TestGenerateCodePreservesThaiCombiningMarks(t *testing.T) {
	got := GenerateCode("ประวัติการจอง")
	if got != "ประวัติการจอง" {
		t.Fatalf("GenerateCode removed Thai vowels or tone marks: %q", got)
	}
}

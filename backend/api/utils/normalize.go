package utils

import "strings"

// NormalizeName returns the canonical value used for case-insensitive,
// whitespace-insensitive uniqueness checks.
func NormalizeName(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

// NormalizeCode returns the canonical representation stored for identifiers.
func NormalizeCode(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

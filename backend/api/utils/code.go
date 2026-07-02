package utils

import (
	"math/rand"
	"strings"
	"time"
	"unicode"
)

// GenerateCode converts a display name into a standardized code.
// E.g., "My Product Name" -> "my_product_name"
func GenerateCode(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return generateRandomCode()
	}

	// Convert spaces and symbols to underscores
	var result strings.Builder
	lastWasUnderscore := false

	for _, char := range name {
		if unicode.IsSpace(char) || char == '-' || char == '_' || char == '/' || char == '\\' || char == '.' || char == ',' {
			if !lastWasUnderscore {
				result.WriteRune('_')
				lastWasUnderscore = true
			}
		} else if unicode.IsLetter(char) || unicode.IsDigit(char) {
			result.WriteRune(unicode.ToLower(char))
			lastWasUnderscore = false
		}
	}

	code := result.String()
	// Trim leading/trailing underscores
	code = strings.Trim(code, "_")

	if code == "" {
		return generateRandomCode()
	}

	return code
}

func generateRandomCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return "auto_" + string(b)
}

package usecase

import (
	"fmt"
	"sort"
	"strings"
)

// buildSearchText creates a searchable projection of every scalar in the raw
// payload. The original payload remains in _source for display.
func buildSearchText(payload map[string]any) string {
	values := make([]string, 0, 32)
	collectSearchValues(&values, payload)
	return strings.Join(values, " ")
}

func collectSearchValues(values *[]string, value any) {
	switch typed := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			collectSearchValues(values, typed[key])
		}
	case []any:
		for _, item := range typed {
			collectSearchValues(values, item)
		}
	case string:
		if trimmed := strings.TrimSpace(typed); trimmed != "" {
			*values = append(*values, trimmed)
		}
	case bool, float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		*values = append(*values, fmt.Sprint(typed))
	}
}

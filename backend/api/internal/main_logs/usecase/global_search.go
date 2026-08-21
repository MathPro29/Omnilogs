package usecase

// globalSearchFields lists every text-oriented field that a free-text search
// should examine.  Fields are tried with multi_match (best_fields) so the
// highest-scoring match wins.  `search_text` keeps the highest boost so
// indices that populate it still get the fastest path.
var globalSearchFields = []string{
	"search_text^3",
	"payload.message^2",
	"payload.error_message^2",
	"payload.request_path",
	"payload.url",
	"payload.trace_id",
	"payload.source_request_id",
	"payload.error_code",
	"payload.stack_trace",
	"data.message^2",
	"data.error_message^2",
	"data.request_path",
	"data.url",
	"data.trace_id",
	"data.source_request_id",
	"data.error_code",
	"data.stack_trace",
}

func buildGlobalSearchClause(keyword string) map[string]any {
	return map[string]any{
		"bool": map[string]any{
			"should": []map[string]any{
				// 1. multi_match across all known text fields (best_fields)
				{
					"multi_match": map[string]any{
						"query":    keyword,
						"fields":   globalSearchFields,
						"type":     "best_fields",
						"operator": "and",
						"lenient":  true,
					},
				},
				// 2. phrase-prefix for partial / type-ahead matching
				{
					"multi_match": map[string]any{
						"query":   keyword,
						"fields":  globalSearchFields,
						"type":    "phrase_prefix",
						"lenient": true,
					},
				},
				// 3. query_string wildcard fallback for substring matches
				{
					"query_string": map[string]any{
						"query":            "*" + escapeQueryString(keyword) + "*",
						"fields":           globalSearchFields,
						"default_operator": "AND",
						"lenient":          true,
					},
				},
			},
			"minimum_should_match": 1,
		},
	}
}

// escapeQueryString escapes special Lucene query_string characters so user
// input is treated literally.
func escapeQueryString(s string) string {
	var b []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '\\', '+', '-', '!', '(', ')', ':', '^', '[', ']', '"', '{', '}', '~', '?', '|', '&', '/':
			b = append(b, '\\', c)
		default:
			b = append(b, c)
		}
	}
	return string(b)
}

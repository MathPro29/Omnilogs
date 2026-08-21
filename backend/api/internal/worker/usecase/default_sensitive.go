package usecase

import "strings"

var defaultSensitiveKeys = map[string]struct{}{
	"authorization":       {},
	"proxy_authorization": {},
	"x_api_key":           {},
	"api_key":             {},
	"access_token":        {},
	"refresh_token":       {},
	"password":            {},
	"secret":              {},
	"cookie":              {},
	"set_cookie":          {},
	"session_id":          {},
	"credit_card":         {},
	"cvv":                 {},
	"private_key":         {},
}

func isDefaultSensitiveKey(key string) bool {
	normalized := strings.NewReplacer("-", "_", " ", "_", ".", "_").
		Replace(strings.ToLower(strings.TrimSpace(key)))
	_, exists := defaultSensitiveKeys[normalized]
	return exists
}

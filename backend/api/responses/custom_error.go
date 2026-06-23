package responses

import "errors"

var ErrorCode = map[string]string{
	"INVALID_REQUEST":      "Invalid request",
	"INVALID_PHONE_NUMBER": "Invalid phone number",
	"INVALID_EMAIL":        "Invalid email",
	"EMAIL_ALREADY_EXISTS": "Email already exists",
	"USER_NOT_FOUND":       "User not found",
	"INVALID_CREDENTIAL":   "Invalid credential",
	"FORBIDDEN":            "Forbidden or Permission Denied",
	"INTERNAL_ERROR":       "Internal Server Error",
	"INVALID_RESET_TOKEN":  "Invalid reset token",
}

var ErrorUserCode = map[string]error{
	"EMAIL_ALREADY_EXISTS":  errors.New("EMAIL_ALREADY_EXISTS"),
	"USER_NOT_FOUND":        errors.New("USER_NOT_FOUND"),
	"INVALID_CREDENTIAL":    errors.New("INVALID_CREDENTIAL"),
	"FORBIDDEN":             errors.New("FORBIDDEN"),
	"INVALID_REFRESH_TOKEN": errors.New("INVALID_REFRESH_TOKEN"),
	"INVALID_RESET_TOKEN":   errors.New("INVALID_RESET_TOKEN"),
}

var ErrorScopeCode = map[string]error{
	"SCOPE_NOT_FOUND":  errors.New("scope resource not found"),
	"SCOPE_ACCESS_DENIED":    errors.New("scope access denied"),
	"SCOPE_CONFLICT":    errors.New("scope resource already exists"),
	"INVALID_SCOPE":   errors.New("invalid scope relationship"),
}

package responses

import "errors"

var (
	ErrNotFound  = errors.New("resource not found")
	ErrInvalid   = errors.New("invalid relationship or payload")
	ErrConflict  = errors.New("resource already exists")
	ErrForbidden = errors.New("forbidden or permission denied")
)

var ErrorCode = map[string]string{
	"INVALID_REQUEST":                "Invalid request",
	"INVALID_PHONE_NUMBER":           "Invalid phone number",
	"INVALID_EMAIL":                  "Invalid email",
	"EMAIL_ALREADY_EXISTS":           "Email already exists",
	"USERNAME_ALREADY_EXISTS":        "Username already exists",
	"USER_NOT_FOUND":                 "User not found",
	"INVALID_CREDENTIAL":             "Invalid credential",
	"FORBIDDEN":                      "Forbidden or Permission Denied",
	"INTERNAL_ERROR":                 "Internal Server Error",
	"INVALID_RESET_TOKEN":            "Invalid reset token",
	"ACCESS_RESOURCE_NOT_FOUND":      "access resource not found",
	"ACCESS_DENIED":                  "access denied",
	"ACCESS_RESOURCE_ALREADY_EXISTS": "access resource already exists",
	"INVALID_ACCESS_RELATIONSHIP":    "invalid access relationship",
}

var ErrorUserCode = map[string]error{
	"EMAIL_ALREADY_EXISTS":    errors.New("EMAIL_ALREADY_EXISTS"),
	"USERNAME_ALREADY_EXISTS": errors.New("USERNAME_ALREADY_EXISTS"),
	"USER_NOT_FOUND":          errors.New("USER_NOT_FOUND"),
	"INVALID_CREDENTIAL":      errors.New("INVALID_CREDENTIAL"),
	"FORBIDDEN":               errors.New("FORBIDDEN"),
	"INVALID_REFRESH_TOKEN":   errors.New("INVALID_REFRESH_TOKEN"),
	"INVALID_RESET_TOKEN":     errors.New("INVALID_RESET_TOKEN"),
}

var ErrorScopeCode = map[string]error{
	"SCOPE_NOT_FOUND":     errors.New("scope resource not found"),
	"SCOPE_ACCESS_DENIED": errors.New("scope access denied"),
	"SCOPE_CONFLICT":      errors.New("scope resource already exists"),
	"INVALID_SCOPE":       errors.New("invalid scope relationship"),
}

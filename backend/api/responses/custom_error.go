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
}

var ErrorUserCode = map[string]error{
	"EMAIL_ALREADY_EXISTS":  errors.New("EMAIL_ALREADY_EXISTS"),
	"USER_NOT_FOUND":        errors.New("USER_NOT_FOUND"),
	"INVALID_CREDENTIAL":    errors.New("INVALID_CREDENTIAL"),
	"FORBIDDEN":             errors.New("FORBIDDEN"),
	"INVALID_REFRESH_TOKEN": errors.New("INVALID_REFRESH_TOKEN"),
}

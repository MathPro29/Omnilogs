package responses

import "fmt"

type ScopedConflictError struct {
	Field   string
	Value   string
	Scope   string
	Message string
	Cause   error
}

func (e *ScopedConflictError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("%s %q already exists in %s", e.Field, e.Value, e.Scope)
}

func (e *ScopedConflictError) Unwrap() error {
	if e.Cause != nil {
		return e.Cause
	}
	return ErrConflict
}

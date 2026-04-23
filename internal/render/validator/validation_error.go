package validator

import "fmt"

// ValidationError represents a validation failure from a specific check.
type ValidationError struct {
	Check string
	Err   error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Check, e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

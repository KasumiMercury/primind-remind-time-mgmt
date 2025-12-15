package app

import (
	"errors"
	"fmt"
)

var (
	ErrValidation    = errors.New("validation error")
	ErrNotFound      = errors.New("resource not found")
	ErrInternalError = errors.New("internal error")
	ErrAlreadyExists = errors.New("resource already exists")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

func IsValidationError(err error) bool {
	var validationErr *ValidationError

	return errors.As(err, &validationErr)
}

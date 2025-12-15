package app_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/app"
)

func TestNewValidationErrorSuccess(t *testing.T) {
	tests := []struct {
		name            string
		field           string
		message         string
		expectedError   string
		expectedField   string
		expectedMessage string
	}{
		{
			name:            "user_id validation error",
			field:           "user_id",
			message:         "must be valid UUIDv7",
			expectedError:   "validation error: user_id - must be valid UUIDv7",
			expectedField:   "user_id",
			expectedMessage: "must be valid UUIDv7",
		},
		{
			name:            "devices validation error with index",
			field:           "devices[0]",
			message:         "device ID cannot be empty",
			expectedError:   "validation error: devices[0] - device ID cannot be empty",
			expectedField:   "devices[0]",
			expectedMessage: "device ID cannot be empty",
		},
		{
			name:            "task_id validation error",
			field:           "task_id",
			message:         "invalid task ID format",
			expectedError:   "validation error: task_id - invalid task ID format",
			expectedField:   "task_id",
			expectedMessage: "invalid task ID format",
		},
		{
			name:            "time_range validation error",
			field:           "time_range",
			message:         "start must be before end",
			expectedError:   "validation error: time_range - start must be before end",
			expectedField:   "time_range",
			expectedMessage: "start must be before end",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := app.NewValidationError(tt.field, tt.message)

			assert.Equal(t, tt.expectedField, err.Field)
			assert.Equal(t, tt.expectedMessage, err.Message)
			assert.Equal(t, tt.expectedError, err.Error())
		})
	}
}

func TestIsValidationErrorSuccess(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "is ValidationError",
			err:      app.NewValidationError("field", "message"),
			expected: true,
		},
		{
			name:     "wrapped ValidationError",
			err:      fmt.Errorf("wrapped: %w", app.NewValidationError("field", "message")),
			expected: true,
		},
		{
			name:     "double wrapped ValidationError",
			err:      fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", app.NewValidationError("field", "message"))),
			expected: true,
		},
		{
			name:     "not ValidationError - generic error",
			err:      errors.New("generic error"),
			expected: false,
		},
		{
			name:     "not ValidationError - nil",
			err:      nil,
			expected: false,
		},
		{
			name:     "not ValidationError - wrapped generic error",
			err:      fmt.Errorf("wrapped: %w", errors.New("generic error")),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.IsValidationError(tt.err)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidationErrorTypeAssertionSuccess(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "can be type asserted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := app.NewValidationError("field", "message")

			var validationErr *app.ValidationError
			assert.True(t, errors.As(err, &validationErr))
			assert.Equal(t, "field", validationErr.Field)
			assert.Equal(t, "message", validationErr.Message)
		})
	}
}

func TestSentinelErrorsSuccess(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "ErrValidation exists",
			err:  app.ErrValidation,
		},
		{
			name: "ErrNotFound exists",
			err:  app.ErrNotFound,
		},
		{
			name: "ErrInternalError exists",
			err:  app.ErrInternalError,
		},
		{
			name: "ErrAlreadyExists exists",
			err:  app.ErrAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.err)
			assert.Error(t, tt.err)
		})
	}
}

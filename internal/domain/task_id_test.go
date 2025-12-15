package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/domain"
)

func TestTaskIDFromStringSuccess(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "valid UUIDv7",
			input: uuid.Must(uuid.NewV7()).String(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := domain.TaskIDFromString(tt.input)

			assert.NoError(t, err)
			assert.Equal(t, tt.input, id.String())
			assert.False(t, id.IsZero())
		})
	}
}

func TestTaskIDFromStringError(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "invalid format",
			input: "not-a-uuid",
		},
		{
			name:  "UUID v4 (not v7)",
			input: uuid.New().String(),
		},
		{
			name:  "UUID v1 (not v7)",
			input: uuid.Must(uuid.NewUUID()).String(),
		},
		{
			name:  "partial UUID",
			input: "550e8400-e29b-41d4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.TaskIDFromString(tt.input)

			assert.ErrorIs(t, err, domain.ErrInvalidTaskID)
		})
	}
}

func TestTaskIDFromUUIDSuccess(t *testing.T) {
	tests := []struct {
		name  string
		input uuid.UUID
	}{
		{
			name:  "valid UUIDv7",
			input: uuid.Must(uuid.NewV7()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := domain.TaskIDFromUUID(tt.input)

			assert.NoError(t, err)
			assert.Equal(t, tt.input, id.UUID())
			assert.False(t, id.IsZero())
		})
	}
}

func TestTaskIDFromUUIDError(t *testing.T) {
	tests := []struct {
		name  string
		input uuid.UUID
	}{
		{
			name:  "UUID v4",
			input: uuid.New(),
		},
		{
			name:  "UUID v1",
			input: uuid.Must(uuid.NewUUID()),
		},
		{
			name:  "nil UUID",
			input: uuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.TaskIDFromUUID(tt.input)

			assert.ErrorIs(t, err, domain.ErrInvalidTaskID)
		})
	}
}

func TestTaskIDEqualsSuccess(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (domain.TaskID, domain.TaskID, error)
		expected bool
	}{
		{
			name: "same UUID returns true",
			setup: func() (domain.TaskID, domain.TaskID, error) {
				u := uuid.Must(uuid.NewV7())

				id1, err := domain.TaskIDFromUUID(u)
				if err != nil {
					return domain.TaskID{}, domain.TaskID{}, err
				}

				id2, err := domain.TaskIDFromUUID(u)

				return id1, id2, err
			},
			expected: true,
		},
		{
			name: "different UUIDs returns false",
			setup: func() (domain.TaskID, domain.TaskID, error) {
				id1, err := domain.TaskIDFromUUID(uuid.Must(uuid.NewV7()))
				if err != nil {
					return domain.TaskID{}, domain.TaskID{}, err
				}

				id2, err := domain.TaskIDFromUUID(uuid.Must(uuid.NewV7()))

				return id1, id2, err
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id1, id2, err := tt.setup()
			assert.NoError(t, err)

			assert.Equal(t, tt.expected, id1.Equals(id2))
		})
	}
}

func TestTaskIDStringSuccess(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "round-trip conversion preserves value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := uuid.Must(uuid.NewV7()).String()
			id, err := domain.TaskIDFromString(input)
			assert.NoError(t, err)

			assert.Equal(t, input, id.String())
		})
	}
}

func TestTaskIDIsZeroSuccess(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (domain.TaskID, error)
		expected bool
	}{
		{
			name: "valid UUIDv7 is not zero",
			setup: func() (domain.TaskID, error) {
				return domain.TaskIDFromUUID(uuid.Must(uuid.NewV7()))
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := tt.setup()
			assert.NoError(t, err)

			assert.Equal(t, tt.expected, id.IsZero())
		})
	}
}

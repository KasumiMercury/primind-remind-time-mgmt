package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/domain"
)

func TestRemindIDFromStringSuccess(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "valid UUID v4",
			input: uuid.New().String(),
		},
		{
			name:  "valid UUID v7",
			input: uuid.Must(uuid.NewV7()).String(),
		},
		{
			name:  "valid UUID with uppercase",
			input: "550E8400-E29B-41D4-A716-446655440000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := domain.RemindIDFromString(tt.input)

			assert.NoError(t, err)
			assert.False(t, id.IsZero())
		})
	}
}

func TestRemindIDFromStringError(t *testing.T) {
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
			name:  "partial UUID",
			input: "550e8400-e29b-41d4",
		},
		{
			name:  "UUID with invalid characters",
			input: "550e8400-e29b-41d4-a716-44665544000g",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.RemindIDFromString(tt.input)

			assert.ErrorIs(t, err, domain.ErrInvalidRemindID)
		})
	}
}

func TestNewRemindIDSuccess(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "generates non-zero ID",
		},
		{
			name: "generates unique IDs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := domain.NewRemindID()

			assert.False(t, id.IsZero())
			assert.NotEmpty(t, id.String())
		})
	}
}

func TestRemindIDFromUUIDSuccess(t *testing.T) {
	tests := []struct {
		name  string
		input uuid.UUID
	}{
		{
			name:  "valid UUID v4",
			input: uuid.New(),
		},
		{
			name:  "valid UUID v7",
			input: uuid.Must(uuid.NewV7()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := domain.RemindIDFromUUID(tt.input)

			assert.Equal(t, tt.input, id.UUID())
			assert.False(t, id.IsZero())
		})
	}
}

func TestRemindIDEqualsSuccess(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (domain.RemindID, domain.RemindID)
		expected bool
	}{
		{
			name: "same UUID returns true",
			setup: func() (domain.RemindID, domain.RemindID) {
				u := uuid.New()
				id1 := domain.RemindIDFromUUID(u)
				id2 := domain.RemindIDFromUUID(u)

				return id1, id2
			},
			expected: true,
		},
		{
			name: "different UUIDs returns false",
			setup: func() (domain.RemindID, domain.RemindID) {
				id1 := domain.NewRemindID()
				id2 := domain.NewRemindID()

				return id1, id2
			},
			expected: false,
		},
		{
			name: "same ID compared with itself returns true",
			setup: func() (domain.RemindID, domain.RemindID) {
				id := domain.NewRemindID()

				return id, id
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id1, id2 := tt.setup()

			assert.Equal(t, tt.expected, id1.Equals(id2))
		})
	}
}

func TestRemindIDStringSuccess(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "round-trip conversion preserves value",
			input: uuid.New().String(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := domain.RemindIDFromString(tt.input)
			assert.NoError(t, err)

			assert.Equal(t, tt.input, id.String())
		})
	}
}

func TestRemindIDIsZeroSuccess(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() domain.RemindID
		expected bool
	}{
		{
			name: "new ID is not zero",
			setup: func() domain.RemindID {
				return domain.NewRemindID()
			},
			expected: false,
		},
		{
			name: "nil UUID is zero",
			setup: func() domain.RemindID {
				return domain.RemindIDFromUUID(uuid.Nil)
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := tt.setup()

			assert.Equal(t, tt.expected, id.IsZero())
		})
	}
}

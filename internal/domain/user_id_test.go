package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/domain"
)

func TestUserIDFromStringSuccess(t *testing.T) {
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
			id, err := domain.UserIDFromString(tt.input)

			assert.NoError(t, err)
			assert.Equal(t, tt.input, id.String())
			assert.False(t, id.IsZero())
		})
	}
}

func TestUserIDFromStringError(t *testing.T) {
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
			_, err := domain.UserIDFromString(tt.input)

			assert.ErrorIs(t, err, domain.ErrInvalidUserID)
		})
	}
}

func TestUserIDFromUUIDSuccess(t *testing.T) {
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
			id, err := domain.UserIDFromUUID(tt.input)

			assert.NoError(t, err)
			assert.Equal(t, tt.input, id.UUID())
			assert.False(t, id.IsZero())
		})
	}
}

func TestUserIDFromUUIDError(t *testing.T) {
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
			_, err := domain.UserIDFromUUID(tt.input)

			assert.ErrorIs(t, err, domain.ErrInvalidUserID)
		})
	}
}

func TestUserIDEqualsSuccess(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (domain.UserID, domain.UserID, error)
		expected bool
	}{
		{
			name: "same UUID returns true",
			setup: func() (domain.UserID, domain.UserID, error) {
				u := uuid.Must(uuid.NewV7())

				id1, err := domain.UserIDFromUUID(u)
				if err != nil {
					return domain.UserID{}, domain.UserID{}, err
				}

				id2, err := domain.UserIDFromUUID(u)

				return id1, id2, err
			},
			expected: true,
		},
		{
			name: "different UUIDs returns false",
			setup: func() (domain.UserID, domain.UserID, error) {
				id1, err := domain.UserIDFromUUID(uuid.Must(uuid.NewV7()))
				if err != nil {
					return domain.UserID{}, domain.UserID{}, err
				}

				id2, err := domain.UserIDFromUUID(uuid.Must(uuid.NewV7()))

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

func TestUserIDStringSuccess(t *testing.T) {
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
			id, err := domain.UserIDFromString(input)
			assert.NoError(t, err)

			assert.Equal(t, input, id.String())
		})
	}
}

func TestUserIDIsZeroSuccess(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (domain.UserID, error)
		expected bool
	}{
		{
			name: "valid UUIDv7 is not zero",
			setup: func() (domain.UserID, error) {
				return domain.UserIDFromUUID(uuid.Must(uuid.NewV7()))
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

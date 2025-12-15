package domain

import (
	"errors"

	"github.com/google/uuid"
)

type UserID struct {
	value uuid.UUID
}

var ErrInvalidUserID = errors.New("invalid user ID: must be valid UUIDv7")

func UserIDFromString(s string) (UserID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return UserID{}, ErrInvalidUserID
	}

	if id.Version() != 7 {
		return UserID{}, ErrInvalidUserID
	}

	return UserID{value: id}, nil
}

func UserIDFromUUID(id uuid.UUID) (UserID, error) {
	if id.Version() != 7 {
		return UserID{}, ErrInvalidUserID
	}

	return UserID{value: id}, nil
}

func (u UserID) String() string {
	return u.value.String()
}

func (u UserID) UUID() uuid.UUID {
	return u.value
}

func (u UserID) IsZero() bool {
	return u.value == uuid.Nil
}

func (u UserID) Equals(other UserID) bool {
	return u.value == other.value
}

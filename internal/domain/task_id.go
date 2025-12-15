package domain

import (
	"errors"

	"github.com/google/uuid"
)

type TaskID struct {
	value uuid.UUID
}

var ErrInvalidTaskID = errors.New("invalid task ID: must be valid UUIDv7")

func TaskIDFromString(s string) (TaskID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return TaskID{}, ErrInvalidTaskID
	}

	if id.Version() != 7 {
		return TaskID{}, ErrInvalidTaskID
	}

	return TaskID{value: id}, nil
}

func TaskIDFromUUID(id uuid.UUID) (TaskID, error) {
	if id.Version() != 7 {
		return TaskID{}, ErrInvalidTaskID
	}

	return TaskID{value: id}, nil
}

func (t TaskID) String() string {
	return t.value.String()
}

func (t TaskID) UUID() uuid.UUID {
	return t.value
}

func (t TaskID) IsZero() bool {
	return t.value == uuid.Nil
}

func (t TaskID) Equals(other TaskID) bool {
	return t.value == other.value
}

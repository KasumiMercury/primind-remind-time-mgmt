package domain

import (
	"github.com/google/uuid"
)

type RemindID struct {
	value uuid.UUID
}

func NewRemindID() RemindID {
	return RemindID{value: uuid.New()}
}

func RemindIDFromString(s string) (RemindID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return RemindID{}, ErrInvalidRemindID
	}

	return RemindID{value: id}, nil
}

func RemindIDFromUUID(id uuid.UUID) RemindID {
	return RemindID{value: id}
}

func (r RemindID) String() string {
	return r.value.String()
}

func (r RemindID) UUID() uuid.UUID {
	return r.value
}

func (r RemindID) IsZero() bool {
	return r.value == uuid.Nil
}

func (r RemindID) Equals(other RemindID) bool {
	return r.value == other.value
}

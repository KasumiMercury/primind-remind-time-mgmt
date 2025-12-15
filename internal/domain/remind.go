package domain

import (
	"time"
)

type Remind struct {
	id        RemindID
	time      time.Time
	userID    UserID
	devices   Devices
	taskID    TaskID
	taskType  Type
	throttled bool
	createdAt time.Time
	updatedAt time.Time
}

func NewRemind(
	remindTime time.Time,
	userID UserID,
	devices Devices,
	taskID TaskID,
	taskType Type,
) (*Remind, error) {
	if remindTime.Before(time.Now().Add(-1 * time.Minute)) {
		return nil, ErrPastRemindTime
	}

	now := time.Now()

	return &Remind{
		id:        NewRemindID(),
		time:      remindTime,
		userID:    userID,
		devices:   devices,
		taskID:    taskID,
		taskType:  taskType,
		throttled: false,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func Reconstitute(
	id RemindID,
	remindTime time.Time,
	userID UserID,
	devices Devices,
	taskID TaskID,
	taskType Type,
	throttled bool,
	createdAt time.Time,
	updatedAt time.Time,
) *Remind {
	return &Remind{
		id:        id,
		time:      remindTime,
		userID:    userID,
		devices:   devices,
		taskID:    taskID,
		taskType:  taskType,
		throttled: throttled,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (r *Remind) MarkAsThrottled() error {
	if r.throttled {
		return ErrAlreadyThrottled
	}

	r.throttled = true
	r.updatedAt = time.Now()

	return nil
}

func (r *Remind) IsThrottled() bool {
	return r.throttled
}

func (r *Remind) IsDue() bool {
	return time.Now().After(r.time)
}

func (r *Remind) ID() RemindID {
	return r.id
}

func (r *Remind) Time() time.Time {
	return r.time
}

func (r *Remind) UserID() UserID {
	return r.userID
}

func (r *Remind) Devices() Devices {
	return r.devices
}

func (r *Remind) TaskID() TaskID {
	return r.taskID
}

func (r *Remind) TaskType() Type {
	return r.taskType
}

func (r *Remind) CreatedAt() time.Time {
	return r.createdAt
}

func (r *Remind) UpdatedAt() time.Time {
	return r.updatedAt
}

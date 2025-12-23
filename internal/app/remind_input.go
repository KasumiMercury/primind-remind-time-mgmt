package app

import "time"

type CreateRemindInput struct {
	Times    []time.Time
	UserID   string
	Devices  []DeviceInput
	TaskID   string
	TaskType string
}

type DeviceInput struct {
	DeviceID string
	FCMToken string
}

type GetRemindsByTimeRangeInput struct {
	Start time.Time
	End   time.Time
}

type UpdateThrottledInput struct {
	ID        string
	Throttled bool
}

type DeleteRemindInput struct {
	ID string
}

type CancelRemindByTaskIDInput struct {
	TaskID string
	UserID string
}

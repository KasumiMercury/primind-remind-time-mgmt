package domain

import "errors"

var (
	ErrRemindNotFound = errors.New("remind not found")

	ErrInvalidTimeRange = errors.New("invalid time range: start must be before end")
	ErrInvalidTaskType  = errors.New("invalid task type")

	ErrPastRemindTime   = errors.New("remind time cannot be in the past")
	ErrAlreadyThrottled = errors.New("remind is already throttled")

	ErrInvalidRemindID = errors.New("invalid remind ID")
)

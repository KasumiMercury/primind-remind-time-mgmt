package app

import (
	"time"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/domain"
)

type RemindOutput struct {
	ID        string
	Time      time.Time
	UserID    string
	Devices   []DeviceOutput
	TaskID    string
	TaskType  string
	Throttled bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DeviceOutput struct {
	DeviceID string
	FCMToken string
}

type RemindsOutput struct {
	Reminds []RemindOutput
	Count   int32
}

func FromEntity(remind *domain.Remind) RemindOutput {
	devices := make([]DeviceOutput, 0, remind.Devices().Count())
	for _, d := range remind.Devices().ToSlice() {
		devices = append(devices, DeviceOutput{
			DeviceID: d.DeviceID(),
			FCMToken: d.FCMToken(),
		})
	}

	return RemindOutput{
		ID:        remind.ID().String(),
		Time:      remind.Time(),
		UserID:    remind.UserID().String(),
		Devices:   devices,
		TaskID:    remind.TaskID().String(),
		TaskType:  string(remind.TaskType()),
		Throttled: remind.IsThrottled(),
		CreatedAt: remind.CreatedAt(),
		UpdatedAt: remind.UpdatedAt(),
	}
}

func FromEntities(reminds []*domain.Remind) RemindsOutput {
	outputs := make([]RemindOutput, 0, len(reminds))
	for _, r := range reminds {
		outputs = append(outputs, FromEntity(r))
	}

	return RemindsOutput{
		Reminds: outputs,
		Count:   int32(len(outputs)), //nolint:gosec
	}
}

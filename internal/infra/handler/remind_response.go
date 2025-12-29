package handler

import (
	"time"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/app"
)

type RemindResponse struct {
	ID               string           `json:"id"`
	Time             time.Time        `json:"time"`
	UserID           string           `json:"user_id"`
	Devices          []DeviceResponse `json:"devices"`
	TaskID           string           `json:"task_id"`
	TaskType         string           `json:"task_type"`
	Throttled        bool             `json:"throttled"`
	SlideWindowWidth int32            `json:"slide_window_width"` // slide window width in seconds (range: 60-600)
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

type DeviceResponse struct {
	DeviceID string `json:"device_id"`
	FCMToken string `json:"fcm_token"`
}

type RemindsResponse struct {
	Reminds []RemindResponse `json:"reminds"`
	Count   int32            `json:"count"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func FromDTO(output app.RemindOutput) RemindResponse {
	devices := make([]DeviceResponse, 0, len(output.Devices))
	for _, d := range output.Devices {
		devices = append(devices, DeviceResponse{
			DeviceID: d.DeviceID,
			FCMToken: d.FCMToken,
		})
	}

	return RemindResponse{
		ID:               output.ID,
		Time:             output.Time,
		UserID:           output.UserID,
		Devices:          devices,
		TaskID:           output.TaskID,
		TaskType:         output.TaskType,
		Throttled:        output.Throttled,
		SlideWindowWidth: output.SlideWindowWidth,
		CreatedAt:        output.CreatedAt,
		UpdatedAt:        output.UpdatedAt,
	}
}

func FromDTOs(output app.RemindsOutput) RemindsResponse {
	reminds := make([]RemindResponse, 0, len(output.Reminds))
	for _, r := range output.Reminds {
		reminds = append(reminds, FromDTO(r))
	}

	return RemindsResponse{
		Reminds: reminds,
		Count:   output.Count,
	}
}

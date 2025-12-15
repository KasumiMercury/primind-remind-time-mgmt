package handler

import "time"

type CreateRemindRequest struct {
	Times    []time.Time     `json:"times" binding:"required,min=1,dive"`
	UserID   string          `json:"user_id" binding:"required,uuid"`
	Devices  []DeviceRequest `json:"devices" binding:"required,min=1,dive"`
	TaskID   string          `json:"task_id" binding:"required,uuid"`
	TaskType string          `json:"task_type" binding:"required"`
}

type DeviceRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
	FCMToken string `json:"fcm_token" binding:"required"`
}

type GetRemindsByTimeRangeRequest struct {
	Start time.Time `form:"start" binding:"required" time_format:"2006-01-02T15:04:05Z07:00"`
	End   time.Time `form:"end" binding:"required,gtfield=Start" time_format:"2006-01-02T15:04:05Z07:00"`
}

type UpdateThrottledRequest struct {
	Throttled bool `json:"throttled"`
}

package handler_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/app"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/infra/handler"
)

func TestFromDTOSuccess(t *testing.T) {
	tests := []struct {
		name        string
		deviceCount int
		throttled   bool
	}{
		{
			name:        "single device not throttled",
			deviceCount: 1,
			throttled:   false,
		},
		{
			name:        "single device throttled",
			deviceCount: 1,
			throttled:   true,
		},
		{
			name:        "multiple devices",
			deviceCount: 3,
			throttled:   false,
		},
		{
			name:        "many devices",
			deviceCount: 5,
			throttled:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devices := make([]app.DeviceOutput, tt.deviceCount)
			for i := 0; i < tt.deviceCount; i++ {
				devices[i] = app.DeviceOutput{
					DeviceID: "device-" + string(rune('a'+i)),
					FCMToken: "token-" + string(rune('a'+i)),
				}
			}

			remindTime := time.Now().Add(1 * time.Hour)
			createdAt := time.Now().Add(-1 * time.Hour)
			updatedAt := time.Now()

			output := app.RemindOutput{
				ID:        "0191c7f0-7c3d-7000-8000-000000000001",
				Time:      remindTime,
				UserID:    "0191c7f0-7c3d-7000-8000-000000000002",
				Devices:   devices,
				TaskID:    "0191c7f0-7c3d-7000-8000-000000000003",
				TaskType:  "task",
				Throttled: tt.throttled,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			}

			response := handler.FromDTO(output)

			assert.Equal(t, output.ID, response.ID)
			assert.Equal(t, output.Time, response.Time)
			assert.Equal(t, output.UserID, response.UserID)
			assert.Equal(t, output.TaskID, response.TaskID)
			assert.Equal(t, output.TaskType, response.TaskType)
			assert.Equal(t, tt.throttled, response.Throttled)
			assert.Equal(t, output.CreatedAt, response.CreatedAt)
			assert.Equal(t, output.UpdatedAt, response.UpdatedAt)
			assert.Len(t, response.Devices, tt.deviceCount)
		})
	}
}

func TestFromDTODevicesSuccess(t *testing.T) {
	tests := []struct {
		name        string
		deviceCount int
	}{
		{
			name:        "converts single device correctly",
			deviceCount: 1,
		},
		{
			name:        "converts multiple devices correctly",
			deviceCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devices := make([]app.DeviceOutput, tt.deviceCount)
			for i := 0; i < tt.deviceCount; i++ {
				devices[i] = app.DeviceOutput{
					DeviceID: "device-" + string(rune('a'+i)),
					FCMToken: "token-" + string(rune('a'+i)),
				}
			}

			output := app.RemindOutput{
				ID:       "0191c7f0-7c3d-7000-8000-000000000001",
				Time:     time.Now(),
				UserID:   "0191c7f0-7c3d-7000-8000-000000000002",
				Devices:  devices,
				TaskID:   "0191c7f0-7c3d-7000-8000-000000000003",
				TaskType: "task",
			}

			response := handler.FromDTO(output)

			for i, deviceResp := range response.Devices {
				assert.Equal(t, devices[i].DeviceID, deviceResp.DeviceID)
				assert.Equal(t, devices[i].FCMToken, deviceResp.FCMToken)
			}
		})
	}
}

func TestFromDTOsSuccess(t *testing.T) {
	tests := []struct {
		name        string
		remindCount int
	}{
		{
			name:        "empty slice",
			remindCount: 0,
		},
		{
			name:        "single remind",
			remindCount: 1,
		},
		{
			name:        "multiple reminds",
			remindCount: 5,
		},
		{
			name:        "many reminds",
			remindCount: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reminds := make([]app.RemindOutput, tt.remindCount)
			for i := 0; i < tt.remindCount; i++ {
				reminds[i] = app.RemindOutput{
					ID:        "0191c7f0-7c3d-7000-8000-00000000000" + string(rune('1'+i)),
					Time:      time.Now().Add(time.Duration(i) * time.Hour),
					UserID:    "0191c7f0-7c3d-7000-8000-000000000002",
					Devices:   []app.DeviceOutput{{DeviceID: "d", FCMToken: "t"}},
					TaskID:    "0191c7f0-7c3d-7000-8000-000000000003",
					TaskType:  "task",
					Throttled: i%2 == 0,
				}
			}

			output := app.RemindsOutput{
				Reminds: reminds,
				Count:   tt.remindCount,
			}

			response := handler.FromDTOs(output)

			assert.Equal(t, tt.remindCount, response.Count)
			assert.Len(t, response.Reminds, tt.remindCount)
		})
	}
}

func TestFromDTOsPreservesOrderSuccess(t *testing.T) {
	tests := []struct {
		name        string
		remindCount int
	}{
		{
			name:        "preserves order of reminds",
			remindCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reminds := make([]app.RemindOutput, tt.remindCount)
			for i := 0; i < tt.remindCount; i++ {
				reminds[i] = app.RemindOutput{
					ID:       "id-" + string(rune('a'+i)),
					Time:     time.Now().Add(time.Duration(i) * time.Hour),
					UserID:   "0191c7f0-7c3d-7000-8000-000000000002",
					Devices:  []app.DeviceOutput{{DeviceID: "d", FCMToken: "t"}},
					TaskID:   "0191c7f0-7c3d-7000-8000-000000000003",
					TaskType: "task",
				}
			}

			output := app.RemindsOutput{
				Reminds: reminds,
				Count:   tt.remindCount,
			}

			response := handler.FromDTOs(output)

			for i, remindResp := range response.Reminds {
				assert.Equal(t, reminds[i].ID, remindResp.ID)
			}
		})
	}
}

func TestFromDTOsWithMixedThrottledSuccess(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "handles mixed throttled states",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reminds := []app.RemindOutput{
				{
					ID:        "id-1",
					Time:      time.Now(),
					UserID:    "0191c7f0-7c3d-7000-8000-000000000002",
					Devices:   []app.DeviceOutput{{DeviceID: "d", FCMToken: "t"}},
					TaskID:    "0191c7f0-7c3d-7000-8000-000000000003",
					TaskType:  "task",
					Throttled: true,
				},
				{
					ID:        "id-2",
					Time:      time.Now(),
					UserID:    "0191c7f0-7c3d-7000-8000-000000000002",
					Devices:   []app.DeviceOutput{{DeviceID: "d", FCMToken: "t"}},
					TaskID:    "0191c7f0-7c3d-7000-8000-000000000003",
					TaskType:  "task",
					Throttled: false,
				},
				{
					ID:        "id-3",
					Time:      time.Now(),
					UserID:    "0191c7f0-7c3d-7000-8000-000000000002",
					Devices:   []app.DeviceOutput{{DeviceID: "d", FCMToken: "t"}},
					TaskID:    "0191c7f0-7c3d-7000-8000-000000000003",
					TaskType:  "task",
					Throttled: true,
				},
			}

			output := app.RemindsOutput{
				Reminds: reminds,
				Count:   3,
			}

			response := handler.FromDTOs(output)

			assert.Equal(t, 3, response.Count)
			assert.True(t, response.Reminds[0].Throttled)
			assert.False(t, response.Reminds[1].Throttled)
			assert.True(t, response.Reminds[2].Throttled)
		})
	}
}

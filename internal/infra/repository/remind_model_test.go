package repository_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/domain"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/infra/repository"
)

func createValidUserID(t *testing.T) domain.UserID {
	t.Helper()

	id, err := domain.UserIDFromUUID(uuid.Must(uuid.NewV7()))
	require.NoError(t, err)

	return id
}

func createValidTaskID(t *testing.T) domain.TaskID {
	t.Helper()

	id, err := domain.TaskIDFromUUID(uuid.Must(uuid.NewV7()))
	require.NoError(t, err)

	return id
}

func createValidDevices(t *testing.T, count int) domain.Devices {
	t.Helper()

	deviceSlice := make([]domain.Device, count)
	for i := 0; i < count; i++ {
		d, err := domain.NewDevice(
			"device-"+string(rune('a'+i)),
			"token-"+string(rune('a'+i)),
		)
		require.NoError(t, err)

		deviceSlice[i] = d
	}

	devices, err := domain.NewDevices(deviceSlice)
	require.NoError(t, err)

	return devices
}

func createValidRemind(t *testing.T, deviceCount int, throttled bool) *domain.Remind {
	t.Helper()

	return domain.Reconstitute(
		domain.NewRemindID(),
		time.Now().Add(1*time.Hour),
		createValidUserID(t),
		createValidDevices(t, deviceCount),
		createValidTaskID(t),
		domain.TypeNear,
		throttled,
		domain.MustSlideWindowWidth(5*time.Minute),
		time.Now().Add(-1*time.Hour),
		time.Now(),
	)
}

func TestFromEntitySuccess(t *testing.T) {
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
			remind := createValidRemind(t, tt.deviceCount, tt.throttled)

			model := repository.FromEntity(remind)

			assert.Equal(t, remind.ID().String(), model.ID)
			assert.Equal(t, remind.Time(), model.Time)
			assert.Equal(t, remind.UserID().String(), model.UserID)
			assert.Equal(t, remind.TaskID().String(), model.TaskID)
			assert.Equal(t, string(remind.TaskType()), model.TaskType)
			assert.Equal(t, tt.throttled, model.Throttled)
			assert.Equal(t, remind.CreatedAt(), model.CreatedAt)
			assert.Equal(t, remind.UpdatedAt(), model.UpdatedAt)
			assert.Len(t, model.Devices, tt.deviceCount)
		})
	}
}

func TestFromEntityDevicesSuccess(t *testing.T) {
	tests := []struct {
		name        string
		deviceCount int
	}{
		{
			name:        "single device converts correctly",
			deviceCount: 1,
		},
		{
			name:        "multiple devices convert correctly",
			deviceCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remind := createValidRemind(t, tt.deviceCount, false)

			model := repository.FromEntity(remind)

			for i, deviceJSON := range model.Devices {
				expectedDevice := remind.Devices().ToSlice()[i]
				assert.Equal(t, expectedDevice.DeviceID(), deviceJSON.DeviceID)
				assert.Equal(t, expectedDevice.FCMToken(), deviceJSON.FCMToken)
			}
		})
	}
}

func TestToEntitySuccess(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devicesJSON := make(repository.DevicesJSONB, tt.deviceCount)
			for i := 0; i < tt.deviceCount; i++ {
				devicesJSON[i] = repository.DeviceJSON{
					DeviceID: "device-" + string(rune('a'+i)),
					FCMToken: "token-" + string(rune('a'+i)),
				}
			}

			remindID := domain.NewRemindID()
			userID := createValidUserID(t)
			taskID := createValidTaskID(t)
			remindTime := time.Now().Add(1 * time.Hour)
			createdAt := time.Now().Add(-1 * time.Hour)
			updatedAt := time.Now()

			model := &repository.RemindModel{
				ID:               remindID.String(),
				Time:             remindTime,
				UserID:           userID.String(),
				Devices:          devicesJSON,
				TaskID:           taskID.String(),
				TaskType:         "near",
				Throttled:        tt.throttled,
				SlideWindowWidth: 300, // 5 minutes in seconds
				CreatedAt:        createdAt,
				UpdatedAt:        updatedAt,
			}

			entity, err := model.ToEntity()

			assert.NoError(t, err)
			assert.Equal(t, remindID.String(), entity.ID().String())
			assert.Equal(t, remindTime, entity.Time())
			assert.Equal(t, userID.String(), entity.UserID().String())
			assert.Equal(t, taskID.String(), entity.TaskID().String())
			assert.Equal(t, domain.TypeNear, entity.TaskType())
			assert.Equal(t, tt.throttled, entity.IsThrottled())
			assert.Equal(t, createdAt, entity.CreatedAt())
			assert.Equal(t, updatedAt, entity.UpdatedAt())
			assert.Equal(t, tt.deviceCount, entity.Devices().Count())
		})
	}
}

func TestToEntityError(t *testing.T) {
	tests := []struct {
		name        string
		setupModel  func(t *testing.T) *repository.RemindModel
		expectedErr string
	}{
		{
			name: "invalid remind ID",
			setupModel: func(t *testing.T) *repository.RemindModel {
				return &repository.RemindModel{
					ID:               "invalid-uuid",
					Time:             time.Now().Add(1 * time.Hour),
					UserID:           createValidUserID(t).String(),
					Devices:          repository.DevicesJSONB{{DeviceID: "d", FCMToken: "t"}},
					TaskID:           createValidTaskID(t).String(),
					TaskType:         "near",
					SlideWindowWidth: 300, // 5 minutes in seconds
				}
			},
			expectedErr: "invalid",
		},
		{
			name: "invalid user ID - not UUIDv7",
			setupModel: func(t *testing.T) *repository.RemindModel {
				return &repository.RemindModel{
					ID:               domain.NewRemindID().String(),
					Time:             time.Now().Add(1 * time.Hour),
					UserID:           uuid.Must(uuid.NewRandom()).String(), // UUIDv4
					Devices:          repository.DevicesJSONB{{DeviceID: "d", FCMToken: "t"}},
					TaskID:           createValidTaskID(t).String(),
					TaskType:         "near",
					SlideWindowWidth: 300, // 5 minutes in seconds
				}
			},
			expectedErr: "UUIDv7",
		},
		{
			name: "invalid task ID - not UUIDv7",
			setupModel: func(t *testing.T) *repository.RemindModel {
				return &repository.RemindModel{
					ID:               domain.NewRemindID().String(),
					Time:             time.Now().Add(1 * time.Hour),
					UserID:           createValidUserID(t).String(),
					Devices:          repository.DevicesJSONB{{DeviceID: "d", FCMToken: "t"}},
					TaskID:           uuid.Must(uuid.NewRandom()).String(), // UUIDv4
					TaskType:         "near",
					SlideWindowWidth: 300, // 5 minutes in seconds
				}
			},
			expectedErr: "UUIDv7",
		},
		{
			name: "empty device ID",
			setupModel: func(t *testing.T) *repository.RemindModel {
				return &repository.RemindModel{
					ID:               domain.NewRemindID().String(),
					Time:             time.Now().Add(1 * time.Hour),
					UserID:           createValidUserID(t).String(),
					Devices:          repository.DevicesJSONB{{DeviceID: "", FCMToken: "t"}},
					TaskID:           createValidTaskID(t).String(),
					TaskType:         "near",
					SlideWindowWidth: 300, // 5 minutes in seconds
				}
			},
			expectedErr: "device ID",
		},
		{
			name: "empty FCM token",
			setupModel: func(t *testing.T) *repository.RemindModel {
				return &repository.RemindModel{
					ID:               domain.NewRemindID().String(),
					Time:             time.Now().Add(1 * time.Hour),
					UserID:           createValidUserID(t).String(),
					Devices:          repository.DevicesJSONB{{DeviceID: "d", FCMToken: ""}},
					TaskID:           createValidTaskID(t).String(),
					TaskType:         "near",
					SlideWindowWidth: 300, // 5 minutes in seconds
				}
			},
			expectedErr: "FCM token",
		},
		{
			name: "empty devices slice",
			setupModel: func(t *testing.T) *repository.RemindModel {
				return &repository.RemindModel{
					ID:               domain.NewRemindID().String(),
					Time:             time.Now().Add(1 * time.Hour),
					UserID:           createValidUserID(t).String(),
					Devices:          repository.DevicesJSONB{},
					TaskID:           createValidTaskID(t).String(),
					TaskType:         "near",
					SlideWindowWidth: 300, // 5 minutes in seconds
				}
			},
			expectedErr: "at least one device",
		},
		{
			name: "invalid task type",
			setupModel: func(t *testing.T) *repository.RemindModel {
				return &repository.RemindModel{
					ID:               domain.NewRemindID().String(),
					Time:             time.Now().Add(1 * time.Hour),
					UserID:           createValidUserID(t).String(),
					Devices:          repository.DevicesJSONB{{DeviceID: "d", FCMToken: "t"}},
					TaskID:           createValidTaskID(t).String(),
					TaskType:         "invalid_type",
					SlideWindowWidth: 300, // 5 minutes in seconds
				}
			},
			expectedErr: "invalid task type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := tt.setupModel(t)

			_, err := model.ToEntity()

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestDevicesJSONBScanSuccess(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expectNil   bool
		expectCount int
	}{
		{
			name:        "scan valid JSON",
			input:       []byte(`[{"device_id":"d1","fcm_token":"t1"},{"device_id":"d2","fcm_token":"t2"}]`),
			expectNil:   false,
			expectCount: 2,
		},
		{
			name:        "scan single device",
			input:       []byte(`[{"device_id":"device","fcm_token":"token"}]`),
			expectNil:   false,
			expectCount: 1,
		},
		{
			name:        "scan nil value",
			input:       nil,
			expectNil:   true,
			expectCount: 0,
		},
		{
			name:        "scan empty array",
			input:       []byte(`[]`),
			expectNil:   false,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var devices repository.DevicesJSONB

			err := devices.Scan(tt.input)

			assert.NoError(t, err)

			if tt.expectNil {
				assert.Nil(t, devices)
			} else {
				assert.Len(t, devices, tt.expectCount)
			}
		})
	}
}

func TestDevicesJSONBScanError(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{
			name:  "invalid type - string",
			input: "not bytes",
		},
		{
			name:  "invalid type - int",
			input: 123,
		},
		{
			name:  "invalid JSON",
			input: []byte(`{invalid json}`),
		},
		{
			name:  "wrong JSON structure",
			input: []byte(`{"not":"array"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var devices repository.DevicesJSONB

			err := devices.Scan(tt.input)

			assert.Error(t, err)
		})
	}
}

func TestDevicesJSONBValueSuccess(t *testing.T) {
	tests := []struct {
		name      string
		devices   repository.DevicesJSONB
		expectNil bool
	}{
		{
			name: "value with devices",
			devices: repository.DevicesJSONB{
				{DeviceID: "d1", FCMToken: "t1"},
				{DeviceID: "d2", FCMToken: "t2"},
			},
			expectNil: false,
		},
		{
			name:      "value with nil devices",
			devices:   nil,
			expectNil: true,
		},
		{
			name:      "value with empty devices",
			devices:   repository.DevicesJSONB{},
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.devices.Value()

			assert.NoError(t, err)

			if tt.expectNil {
				assert.Nil(t, value)
			} else {
				assert.NotNil(t, value)
			}
		})
	}
}

func TestTableNameSuccess(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "returns correct table name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := repository.RemindModel{}

			assert.Equal(t, "reminds", model.TableName())
		})
	}
}

func TestRoundTripConversionSuccess(t *testing.T) {
	tests := []struct {
		name        string
		deviceCount int
		throttled   bool
	}{
		{
			name:        "round trip with single device",
			deviceCount: 1,
			throttled:   false,
		},
		{
			name:        "round trip with multiple devices",
			deviceCount: 3,
			throttled:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := createValidRemind(t, tt.deviceCount, tt.throttled)

			model := repository.FromEntity(original)
			restored, err := model.ToEntity()

			assert.NoError(t, err)
			assert.Equal(t, original.ID().String(), restored.ID().String())
			assert.Equal(t, original.Time(), restored.Time())
			assert.Equal(t, original.UserID().String(), restored.UserID().String())
			assert.Equal(t, original.TaskID().String(), restored.TaskID().String())
			assert.Equal(t, original.TaskType(), restored.TaskType())
			assert.Equal(t, original.IsThrottled(), restored.IsThrottled())
			assert.Equal(t, original.CreatedAt(), restored.CreatedAt())
			assert.Equal(t, original.UpdatedAt(), restored.UpdatedAt())
			assert.Equal(t, original.Devices().Count(), restored.Devices().Count())
		})
	}
}

package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/domain"
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

func TestNewRemindSuccess(t *testing.T) {
	tests := []struct {
		name       string
		remindTime time.Time
		taskType   domain.Type
	}{
		{
			name:       "future time - 1 hour ahead with urgent type",
			remindTime: time.Now().Add(1 * time.Hour),
			taskType:   domain.TypeUrgent,
		},
		{
			name:       "future time - 1 minute ahead with normal type",
			remindTime: time.Now().Add(1 * time.Minute),
			taskType:   domain.TypeNormal,
		},
		{
			name:       "future time - 24 hours ahead with low type",
			remindTime: time.Now().Add(24 * time.Hour),
			taskType:   domain.TypeLow,
		},
		{
			name:       "within tolerance - 30 seconds in past with scheduled type",
			remindTime: time.Now().Add(-30 * time.Second),
			taskType:   domain.TypeScheduled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := createValidUserID(t)
			taskID := createValidTaskID(t)
			devices := createValidDevices(t, 1)

			remind, err := domain.NewRemind(tt.remindTime, userID, devices, taskID, tt.taskType)

			assert.NoError(t, err)
			assert.NotNil(t, remind)
			assert.False(t, remind.ID().IsZero())
			assert.Equal(t, tt.remindTime, remind.Time())
			assert.Equal(t, userID, remind.UserID())
			assert.Equal(t, taskID, remind.TaskID())
			assert.Equal(t, tt.taskType, remind.TaskType())
			assert.False(t, remind.IsThrottled())
			assert.Equal(t, 1, remind.Devices().Count())
		})
	}
}

func TestNewRemindError(t *testing.T) {
	tests := []struct {
		name        string
		remindTime  time.Time
		taskType    domain.Type
		expectedErr error
	}{
		{
			name:        "past time - 2 minutes ago",
			remindTime:  time.Now().Add(-2 * time.Minute),
			taskType:    domain.TypeNormal,
			expectedErr: domain.ErrPastRemindTime,
		},
		{
			name:        "past time - 1 hour ago",
			remindTime:  time.Now().Add(-1 * time.Hour),
			taskType:    domain.TypeNormal,
			expectedErr: domain.ErrPastRemindTime,
		},
		{
			name:        "past time - 24 hours ago",
			remindTime:  time.Now().Add(-24 * time.Hour),
			taskType:    domain.TypeNormal,
			expectedErr: domain.ErrPastRemindTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := createValidUserID(t)
			taskID := createValidTaskID(t)
			devices := createValidDevices(t, 1)

			_, err := domain.NewRemind(tt.remindTime, userID, devices, taskID, tt.taskType)

			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestNewRemindWithMultipleDevicesSuccess(t *testing.T) {
	tests := []struct {
		name        string
		deviceCount int
	}{
		{
			name:        "single device",
			deviceCount: 1,
		},
		{
			name:        "two devices",
			deviceCount: 2,
		},
		{
			name:        "five devices",
			deviceCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := createValidUserID(t)
			taskID := createValidTaskID(t)
			devices := createValidDevices(t, tt.deviceCount)

			remind, err := domain.NewRemind(
				time.Now().Add(1*time.Hour),
				userID,
				devices,
				taskID,
				domain.TypeNormal,
			)

			assert.NoError(t, err)
			assert.Equal(t, tt.deviceCount, remind.Devices().Count())
		})
	}
}

func TestMarkAsThrottledSuccess(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "mark non-throttled remind as throttled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := createValidUserID(t)
			taskID := createValidTaskID(t)
			devices := createValidDevices(t, 1)

			remind, err := domain.NewRemind(
				time.Now().Add(1*time.Hour),
				userID,
				devices,
				taskID,
				domain.TypeNormal,
			)
			require.NoError(t, err)
			assert.False(t, remind.IsThrottled())

			err = remind.MarkAsThrottled()

			assert.NoError(t, err)
			assert.True(t, remind.IsThrottled())
		})
	}
}

func TestMarkAsThrottledError(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "already throttled remind returns error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := createValidUserID(t)
			taskID := createValidTaskID(t)
			devices := createValidDevices(t, 1)

			remind, err := domain.NewRemind(
				time.Now().Add(1*time.Hour),
				userID,
				devices,
				taskID,
				domain.TypeNormal,
			)
			require.NoError(t, err)

			err = remind.MarkAsThrottled()
			require.NoError(t, err)

			err = remind.MarkAsThrottled()

			assert.ErrorIs(t, err, domain.ErrAlreadyThrottled)
		})
	}
}

func TestIsDueSuccess(t *testing.T) {
	tests := []struct {
		name       string
		remindTime time.Time
		expected   bool
	}{
		{
			name:       "past time is due",
			remindTime: time.Now().Add(-1 * time.Hour),
			expected:   true,
		},
		{
			name:       "future time is not due",
			remindTime: time.Now().Add(1 * time.Hour),
			expected:   false,
		},
		{
			name:       "far future time is not due",
			remindTime: time.Now().Add(24 * time.Hour),
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := createValidUserID(t)
			taskID := createValidTaskID(t)
			devices := createValidDevices(t, 1)

			remind := domain.Reconstitute(
				domain.NewRemindID(),
				tt.remindTime,
				userID,
				devices,
				taskID,
				domain.TypeNormal,
				false,
				time.Now(),
				time.Now(),
			)

			assert.Equal(t, tt.expected, remind.IsDue())
		})
	}
}

func TestReconstituteSuccess(t *testing.T) {
	tests := []struct {
		name      string
		throttled bool
	}{
		{
			name:      "reconstitute non-throttled remind",
			throttled: false,
		},
		{
			name:      "reconstitute throttled remind",
			throttled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := domain.NewRemindID()
			remindTime := time.Now().Add(1 * time.Hour)
			userID := createValidUserID(t)
			taskID := createValidTaskID(t)
			devices := createValidDevices(t, 2)
			taskType := domain.TypeNormal
			createdAt := time.Now().Add(-1 * time.Hour)
			updatedAt := time.Now()

			remind := domain.Reconstitute(
				id,
				remindTime,
				userID,
				devices,
				taskID,
				taskType,
				tt.throttled,
				createdAt,
				updatedAt,
			)

			assert.Equal(t, id, remind.ID())
			assert.Equal(t, remindTime, remind.Time())
			assert.Equal(t, userID, remind.UserID())
			assert.Equal(t, taskID, remind.TaskID())
			assert.Equal(t, taskType, remind.TaskType())
			assert.Equal(t, tt.throttled, remind.IsThrottled())
			assert.Equal(t, createdAt, remind.CreatedAt())
			assert.Equal(t, updatedAt, remind.UpdatedAt())
			assert.Equal(t, 2, remind.Devices().Count())
		})
	}
}

func TestReconstituteWithPastTimeSuccess(t *testing.T) {
	tests := []struct {
		name       string
		remindTime time.Time
	}{
		{
			name:       "reconstitute with past time succeeds",
			remindTime: time.Now().Add(-24 * time.Hour),
		},
		{
			name:       "reconstitute with far past time succeeds",
			remindTime: time.Now().Add(-7 * 24 * time.Hour),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := createValidUserID(t)
			taskID := createValidTaskID(t)
			devices := createValidDevices(t, 1)

			remind := domain.Reconstitute(
				domain.NewRemindID(),
				tt.remindTime,
				userID,
				devices,
				taskID,
				domain.TypeNormal,
				false,
				time.Now(),
				time.Now(),
			)

			assert.NotNil(t, remind)
			assert.Equal(t, tt.remindTime, remind.Time())
		})
	}
}

func TestRemindGettersSuccess(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "all getters return correct values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := domain.NewRemindID()
			remindTime := time.Now().Add(1 * time.Hour)
			userID := createValidUserID(t)
			taskID := createValidTaskID(t)
			devices := createValidDevices(t, 1)
			taskType := domain.TypeNormal
			createdAt := time.Now().Add(-1 * time.Hour)
			updatedAt := time.Now()

			remind := domain.Reconstitute(
				id,
				remindTime,
				userID,
				devices,
				taskID,
				taskType,
				true,
				createdAt,
				updatedAt,
			)

			assert.Equal(t, id, remind.ID())
			assert.Equal(t, remindTime, remind.Time())
			assert.Equal(t, userID, remind.UserID())
			assert.Equal(t, taskID, remind.TaskID())
			assert.Equal(t, taskType, remind.TaskType())
			assert.True(t, remind.IsThrottled())
			assert.Equal(t, createdAt, remind.CreatedAt())
			assert.Equal(t, updatedAt, remind.UpdatedAt())
			assert.NotNil(t, remind.Devices())
		})
	}
}

func TestNewRemindGeneratesUniqueIDsSuccess(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{
			name:  "multiple reminds have unique IDs",
			count: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := createValidUserID(t)
			taskID := createValidTaskID(t)
			devices := createValidDevices(t, 1)

			ids := make(map[string]bool)

			for i := 0; i < tt.count; i++ {
				remind, err := domain.NewRemind(
					time.Now().Add(time.Duration(i+1)*time.Hour),
					userID,
					devices,
					taskID,
					domain.TypeNormal,
				)
				require.NoError(t, err)

				idStr := remind.ID().String()
				assert.False(t, ids[idStr], "duplicate ID found: %s", idStr)
				ids[idStr] = true
			}
		})
	}
}

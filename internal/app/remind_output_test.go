package app_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/app"
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
			name:        "multiple devices not throttled",
			deviceCount: 3,
			throttled:   false,
		},
		{
			name:        "multiple devices throttled",
			deviceCount: 5,
			throttled:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remind := createValidRemind(t, tt.deviceCount, tt.throttled)

			output := app.FromEntity(remind)

			assert.Equal(t, remind.ID().String(), output.ID)
			assert.Equal(t, remind.Time(), output.Time)
			assert.Equal(t, remind.UserID().String(), output.UserID)
			assert.Equal(t, remind.TaskID().String(), output.TaskID)
			assert.Equal(t, string(remind.TaskType()), output.TaskType)
			assert.Equal(t, tt.throttled, output.Throttled)
			assert.Equal(t, remind.CreatedAt(), output.CreatedAt)
			assert.Equal(t, remind.UpdatedAt(), output.UpdatedAt)
			assert.Len(t, output.Devices, tt.deviceCount)
		})
	}
}

func TestFromEntityDevicesSuccess(t *testing.T) {
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
			remind := createValidRemind(t, tt.deviceCount, false)

			output := app.FromEntity(remind)

			for i, deviceOutput := range output.Devices {
				expectedDevice := remind.Devices().ToSlice()[i]
				assert.Equal(t, expectedDevice.DeviceID(), deviceOutput.DeviceID)
				assert.Equal(t, expectedDevice.FCMToken(), deviceOutput.FCMToken)
			}
		})
	}
}

func TestFromEntitiesSuccess(t *testing.T) {
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
			reminds := make([]*domain.Remind, tt.remindCount)
			for i := 0; i < tt.remindCount; i++ {
				reminds[i] = createValidRemind(t, 1, i%2 == 0)
			}

			output := app.FromEntities(reminds)

			assert.Equal(t, tt.remindCount, output.Count)
			assert.Len(t, output.Reminds, tt.remindCount)
		})
	}
}

func TestFromEntitiesPreservesOrderSuccess(t *testing.T) {
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
			reminds := make([]*domain.Remind, tt.remindCount)
			for i := 0; i < tt.remindCount; i++ {
				reminds[i] = createValidRemind(t, 1, false)
			}

			output := app.FromEntities(reminds)

			for i, remindOutput := range output.Reminds {
				assert.Equal(t, reminds[i].ID().String(), remindOutput.ID)
			}
		})
	}
}

func TestFromEntitiesWithMixedThrottledSuccess(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "handles mixed throttled states",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reminds := []*domain.Remind{
				createValidRemind(t, 1, true),
				createValidRemind(t, 2, false),
				createValidRemind(t, 1, true),
			}

			output := app.FromEntities(reminds)

			assert.Equal(t, 3, output.Count)
			assert.True(t, output.Reminds[0].Throttled)
			assert.False(t, output.Reminds[1].Throttled)
			assert.True(t, output.Reminds[2].Throttled)
		})
	}
}

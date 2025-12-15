package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/domain"
)

func TestNewDeviceSuccess(t *testing.T) {
	tests := []struct {
		name     string
		deviceID string
		fcmToken string
	}{
		{
			name:     "valid device",
			deviceID: "device-123",
			fcmToken: "fcm-token-abc",
		},
		{
			name:     "minimal values",
			deviceID: "a",
			fcmToken: "b",
		},
		{
			name:     "UUID-like device ID",
			deviceID: "0191c7f0-7c3d-7000-8000-000000000001",
			fcmToken: "long-fcm-token-value-here",
		},
		{
			name:     "device with special characters",
			deviceID: "device_123-abc",
			fcmToken: "token:with:colons",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device, err := domain.NewDevice(tt.deviceID, tt.fcmToken)

			assert.NoError(t, err)
			assert.Equal(t, tt.deviceID, device.DeviceID())
			assert.Equal(t, tt.fcmToken, device.FCMToken())
		})
	}
}

func TestNewDeviceError(t *testing.T) {
	tests := []struct {
		name        string
		deviceID    string
		fcmToken    string
		expectedErr error
	}{
		{
			name:        "empty device ID",
			deviceID:    "",
			fcmToken:    "valid-token",
			expectedErr: domain.ErrEmptyDeviceID,
		},
		{
			name:        "empty FCM token",
			deviceID:    "valid-device",
			fcmToken:    "",
			expectedErr: domain.ErrEmptyFCMToken,
		},
		{
			name:        "both empty - device ID checked first",
			deviceID:    "",
			fcmToken:    "",
			expectedErr: domain.ErrEmptyDeviceID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewDevice(tt.deviceID, tt.fcmToken)

			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestDeviceEqualsSuccess(t *testing.T) {
	tests := []struct {
		name     string
		device1  func() domain.Device
		device2  func() domain.Device
		expected bool
	}{
		{
			name: "equal devices",
			device1: func() domain.Device {
				d, _ := domain.NewDevice("id", "token")

				return d
			},
			device2: func() domain.Device {
				d, _ := domain.NewDevice("id", "token")

				return d
			},
			expected: true,
		},
		{
			name: "different device IDs",
			device1: func() domain.Device {
				d, _ := domain.NewDevice("id1", "token")

				return d
			},
			device2: func() domain.Device {
				d, _ := domain.NewDevice("id2", "token")

				return d
			},
			expected: false,
		},
		{
			name: "different FCM tokens",
			device1: func() domain.Device {
				d, _ := domain.NewDevice("id", "token1")

				return d
			},
			device2: func() domain.Device {
				d, _ := domain.NewDevice("id", "token2")

				return d
			},
			expected: false,
		},
		{
			name: "completely different devices",
			device1: func() domain.Device {
				d, _ := domain.NewDevice("id1", "token1")

				return d
			},
			device2: func() domain.Device {
				d, _ := domain.NewDevice("id2", "token2")

				return d
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d1 := tt.device1()
			d2 := tt.device2()

			assert.Equal(t, tt.expected, d1.Equals(d2))
		})
	}
}

func TestNewDevicesSuccess(t *testing.T) {
	tests := []struct {
		name          string
		deviceCount   int
		expectedCount int
	}{
		{
			name:          "single device",
			deviceCount:   1,
			expectedCount: 1,
		},
		{
			name:          "two devices",
			deviceCount:   2,
			expectedCount: 2,
		},
		{
			name:          "multiple devices",
			deviceCount:   5,
			expectedCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test devices
			deviceSlice := make([]domain.Device, tt.deviceCount)
			for i := 0; i < tt.deviceCount; i++ {
				d, err := domain.NewDevice(
					"device-"+string(rune('a'+i)),
					"token-"+string(rune('a'+i)),
				)
				assert.NoError(t, err)

				deviceSlice[i] = d
			}

			devices, err := domain.NewDevices(deviceSlice)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, devices.Count())
			assert.Len(t, devices.ToSlice(), tt.expectedCount)
		})
	}
}

func TestNewDevicesError(t *testing.T) {
	tests := []struct {
		name    string
		devices []domain.Device
	}{
		{
			name:    "empty slice",
			devices: []domain.Device{},
		},
		{
			name:    "nil slice",
			devices: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewDevices(tt.devices)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "at least one device is required")
		})
	}
}

func TestDevicesToSliceSuccess(t *testing.T) {
	tests := []struct {
		name        string
		deviceCount int
	}{
		{
			name:        "single device returns correct slice",
			deviceCount: 1,
		},
		{
			name:        "multiple devices returns correct slice",
			deviceCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test devices
			deviceSlice := make([]domain.Device, tt.deviceCount)
			for i := 0; i < tt.deviceCount; i++ {
				d, _ := domain.NewDevice(
					"device-"+string(rune('a'+i)),
					"token-"+string(rune('a'+i)),
				)
				deviceSlice[i] = d
			}

			devices, err := domain.NewDevices(deviceSlice)
			assert.NoError(t, err)

			result := devices.ToSlice()

			assert.Equal(t, deviceSlice, result)
		})
	}
}

func TestDevicesCountSuccess(t *testing.T) {
	tests := []struct {
		name          string
		deviceCount   int
		expectedCount int
	}{
		{
			name:          "single device",
			deviceCount:   1,
			expectedCount: 1,
		},
		{
			name:          "multiple devices",
			deviceCount:   10,
			expectedCount: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deviceSlice := make([]domain.Device, tt.deviceCount)
			for i := 0; i < tt.deviceCount; i++ {
				d, _ := domain.NewDevice(
					"device-"+string(rune('a'+i)),
					"token-"+string(rune('a'+i)),
				)
				deviceSlice[i] = d
			}

			devices, err := domain.NewDevices(deviceSlice)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedCount, devices.Count())
		})
	}
}

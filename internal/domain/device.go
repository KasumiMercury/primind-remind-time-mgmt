package domain

import "errors"

type Device struct {
	deviceID string
	fcmToken string
}

var (
	ErrEmptyDeviceID = errors.New("device ID cannot be empty")
	ErrEmptyFCMToken = errors.New("FCM token cannot be empty")
)

func NewDevice(deviceID, fcmToken string) (Device, error) {
	if deviceID == "" {
		return Device{}, ErrEmptyDeviceID
	}

	if fcmToken == "" {
		return Device{}, ErrEmptyFCMToken
	}

	return Device{
		deviceID: deviceID,
		fcmToken: fcmToken,
	}, nil
}

func (d Device) DeviceID() string {
	return d.deviceID
}

func (d Device) FCMToken() string {
	return d.fcmToken
}

func (d Device) Equals(other Device) bool {
	return d.deviceID == other.deviceID && d.fcmToken == other.fcmToken
}

type Devices []Device

func NewDevices(devices []Device) (Devices, error) {
	if len(devices) == 0 {
		return nil, errors.New("at least one device is required")
	}

	return Devices(devices), nil
}

func (d Devices) ToSlice() []Device {
	return d
}

func (d Devices) Count() int {
	return len(d)
}

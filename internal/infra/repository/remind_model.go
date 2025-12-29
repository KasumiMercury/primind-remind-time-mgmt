package repository

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/domain"
)

type DeviceJSON struct {
	DeviceID string `json:"device_id"`
	FCMToken string `json:"fcm_token"`
}

type DevicesJSONB []DeviceJSON

func (d *DevicesJSONB) Scan(value interface{}) error {
	if value == nil {
		*d = nil

		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan DevicesJSONB: expected []byte")
	}

	return json.Unmarshal(bytes, d)
}

func (d DevicesJSONB) Value() (driver.Value, error) {
	if d == nil {
		return nil, nil //nolint:nilnil
	}

	return json.Marshal(d)
}

type RemindModel struct {
	ID          string       `gorm:"column:id;type:uuid;primaryKey"`
	Time        time.Time    `gorm:"column:time;type:timestamptz;not null;index:idx_reminds_time;uniqueIndex:idx_reminds_task_id_time"`
	UserID      string       `gorm:"column:user_id;type:uuid;not null;index:idx_reminds_user_id"`
	Devices     DevicesJSONB `gorm:"column:devices;type:jsonb;not null"`
	TaskID      string       `gorm:"column:task_id;type:uuid;not null;uniqueIndex:idx_reminds_task_id_time"`
	TaskType    string       `gorm:"column:task_type;type:varchar(255);not null"`
	Throttled        bool  `gorm:"column:throttled;type:boolean;not null;default:false;index:idx_reminds_throttled"`
	SlideWindowWidth int32 `gorm:"column:slide_window_width;type:integer;not null"` // stored as seconds
	CreatedAt        time.Time    `gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt   time.Time    `gorm:"column:updated_at;type:timestamptz;not null"`
}

func (RemindModel) TableName() string {
	return "reminds"
}

func (m *RemindModel) ToEntity() (*domain.Remind, error) {
	remindID, err := domain.RemindIDFromString(m.ID)
	if err != nil {
		return nil, err
	}

	userID, err := domain.UserIDFromString(m.UserID)
	if err != nil {
		return nil, err
	}

	taskID, err := domain.TaskIDFromString(m.TaskID)
	if err != nil {
		return nil, err
	}

	devices := make([]domain.Device, 0, len(m.Devices))
	for _, d := range m.Devices {
		device, err := domain.NewDevice(d.DeviceID, d.FCMToken)
		if err != nil {
			return nil, err
		}

		devices = append(devices, device)
	}

	deviceCollection, err := domain.NewDevices(devices)
	if err != nil {
		return nil, err
	}

	taskType, err := domain.NewType(m.TaskType)
	if err != nil {
		return nil, err
	}

	slideWindowWidth, err := domain.SlideWindowWidthFromSeconds(m.SlideWindowWidth)
	if err != nil {
		return nil, err
	}

	return domain.Reconstitute(
		remindID,
		m.Time,
		userID,
		deviceCollection,
		taskID,
		taskType,
		m.Throttled,
		slideWindowWidth,
		m.CreatedAt,
		m.UpdatedAt,
	), nil
}

func FromEntity(e *domain.Remind) *RemindModel {
	devices := make(DevicesJSONB, 0, e.Devices().Count())
	for _, d := range e.Devices().ToSlice() {
		devices = append(devices, DeviceJSON{
			DeviceID: d.DeviceID(),
			FCMToken: d.FCMToken(),
		})
	}

	return &RemindModel{
		ID:               e.ID().String(),
		Time:             e.Time(),
		UserID:           e.UserID().String(),
		Devices:          devices,
		TaskID:           e.TaskID().String(),
		TaskType:         string(e.TaskType()),
		Throttled:        e.IsThrottled(),
		SlideWindowWidth: e.SlideWindowWidth().Seconds(),
		CreatedAt:        e.CreatedAt(),
		UpdatedAt:        e.UpdatedAt(),
	}
}

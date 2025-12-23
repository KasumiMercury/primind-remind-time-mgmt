package domain

import (
	"context"
	"time"
)

type TimeRange struct {
	Start time.Time
	End   time.Time
}

type RemindRepository interface {
	Save(ctx context.Context, remind *Remind) error
	FindByID(ctx context.Context, id RemindID) (*Remind, error)
	FindByTaskID(ctx context.Context, taskID TaskID) ([]*Remind, error)
	FindByTimeRange(ctx context.Context, timeRange TimeRange) ([]*Remind, error)
	Update(ctx context.Context, remind *Remind) error
	Delete(ctx context.Context, id RemindID) error
	DeleteByTaskID(ctx context.Context, taskID TaskID) (int64, error)
	WithTx(ctx context.Context, fn func(repo RemindRepository) error) error
}

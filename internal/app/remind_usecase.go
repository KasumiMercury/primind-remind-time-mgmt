package app

import (
	"context"
)

type RemindUseCase interface {
	CreateRemind(ctx context.Context, input CreateRemindInput) (RemindsOutput, error)
	GetRemindsByTimeRange(ctx context.Context, input GetRemindsByTimeRangeInput) (RemindsOutput, error)
	UpdateThrottled(ctx context.Context, input UpdateThrottledInput) (RemindOutput, error)
	DeleteRemind(ctx context.Context, input DeleteRemindInput) error
	CancelRemindByTaskID(ctx context.Context, input CancelRemindByTaskIDInput) error
}

package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/domain"
	throttlev1 "github.com/KasumiMercury/primind-remind-time-mgmt/internal/gen/throttle/v1"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/infra/pubsub"
)

type remindUseCaseImpl struct {
	repo      domain.RemindRepository
	publisher pubsub.Publisher
}

func NewRemindUseCase(repo domain.RemindRepository, publisher pubsub.Publisher) RemindUseCase {
	return &remindUseCaseImpl{
		repo:      repo,
		publisher: publisher,
	}
}

func (uc *remindUseCaseImpl) CreateRemind(ctx context.Context, input CreateRemindInput) (RemindsOutput, error) {
	slog.Debug("creating reminds",
		"task_id", input.TaskID,
		"user_id", input.UserID,
		"times_count", len(input.Times),
	)

	if len(input.Times) == 0 {
		return RemindsOutput{}, NewValidationError("times", "at least one time is required")
	}

	userID, err := domain.UserIDFromString(input.UserID)
	if err != nil {
		return RemindsOutput{}, NewValidationError("user_id", err.Error())
	}

	taskID, err := domain.TaskIDFromString(input.TaskID)
	if err != nil {
		return RemindsOutput{}, NewValidationError("task_id", err.Error())
	}

	existing, err := uc.repo.FindByTaskID(ctx, taskID)
	if err != nil {
		slog.Error("failed to check existing reminds",
			"error", err,
			"task_id", input.TaskID,
		)

		return RemindsOutput{}, fmt.Errorf("%w: %v", ErrInternalError, err)
	}

	if len(existing) > 0 {
		slog.Info("returning existing reminds (idempotency)",
			"task_id", input.TaskID,
			"count", len(existing),
		)

		return FromEntities(existing), nil
	}

	devices := make([]domain.Device, 0, len(input.Devices))
	for i, d := range input.Devices {
		device, err := domain.NewDevice(d.DeviceID, d.FCMToken)
		if err != nil {
			return RemindsOutput{}, NewValidationError(
				fmt.Sprintf("devices[%d]", i), err.Error(),
			)
		}

		devices = append(devices, device)
	}

	deviceCollection, err := domain.NewDevices(devices)
	if err != nil {
		return RemindsOutput{}, NewValidationError("devices", err.Error())
	}

	taskType, err := domain.NewType(input.TaskType)
	if err != nil {
		return RemindsOutput{}, NewValidationError("task_type", err.Error())
	}

	reminds := make([]*domain.Remind, 0, len(input.Times))
	for i, t := range input.Times {
		remind, err := domain.NewRemind(
			t,
			userID,
			deviceCollection,
			taskID,
			taskType,
		)
		if err != nil {
			return RemindsOutput{}, NewValidationError(
				fmt.Sprintf("times[%d]", i), err.Error(),
			)
		}

		reminds = append(reminds, remind)
	}

	if err := uc.repo.WithTx(ctx, func(txRepo domain.RemindRepository) error {
		for _, remind := range reminds {
			if err := txRepo.Save(ctx, remind); err != nil {
				slog.Error("failed to save remind",
					"error", err,
					"task_id", input.TaskID,
					"remind_id", remind.ID().String(),
				)

				return err
			}
		}

		return nil
	}); err != nil {
		return RemindsOutput{}, fmt.Errorf("%w: %v", ErrInternalError, err)
	}

	slog.Debug("reminds created",
		"task_id", input.TaskID,
		"count", len(reminds),
	)

	return FromEntities(reminds), nil
}

func (uc *remindUseCaseImpl) GetRemindsByTimeRange(ctx context.Context, input GetRemindsByTimeRangeInput) (RemindsOutput, error) {
	slog.Debug("getting reminds by time range",
		"start", input.Start,
		"end", input.End,
	)

	if input.Start.After(input.End) {
		return RemindsOutput{}, NewValidationError("time_range", domain.ErrInvalidTimeRange.Error())
	}

	timeRange := domain.TimeRange{
		Start: input.Start,
		End:   input.End,
	}

	reminds, err := uc.repo.FindByTimeRange(ctx, timeRange)
	if err != nil {
		slog.Error("failed to get reminds by time range",
			"error", err,
			"start", input.Start,
			"end", input.End,
		)

		return RemindsOutput{}, fmt.Errorf("%w: %v", ErrInternalError, err)
	}

	slog.Debug("reminds retrieved",
		"count", len(reminds),
		"start", input.Start,
		"end", input.End,
	)

	return FromEntities(reminds), nil
}

func (uc *remindUseCaseImpl) UpdateThrottled(ctx context.Context, input UpdateThrottledInput) (RemindOutput, error) {
	slog.Debug("updating throttled status",
		"remind_id", input.ID,
		"throttled", input.Throttled,
	)

	remindID, err := domain.RemindIDFromString(input.ID)
	if err != nil {
		return RemindOutput{}, NewValidationError("id", err.Error())
	}

	remind, err := uc.repo.FindByID(ctx, remindID)
	if err != nil {
		slog.Warn("remind not found for throttled update",
			"remind_id", input.ID,
			"error", err,
		)

		return RemindOutput{}, fmt.Errorf("%w: %v", ErrNotFound, err)
	}

	if input.Throttled {
		if err := remind.MarkAsThrottled(); err != nil {
			if !errors.Is(err, domain.ErrAlreadyThrottled) {
				return RemindOutput{}, NewValidationError("throttled", err.Error())
			}

			slog.Info("remind already throttled (idempotency)",
				"remind_id", input.ID,
			)
		}
	}

	if err := uc.repo.Update(ctx, remind); err != nil {
		slog.Error("failed to update throttled status",
			"error", err,
			"remind_id", input.ID,
		)

		return RemindOutput{}, fmt.Errorf("%w: %v", ErrInternalError, err)
	}

	slog.Debug("throttled status updated",
		"remind_id", input.ID,
		"throttled", remind.IsThrottled(),
	)

	return FromEntity(remind), nil
}

func (uc *remindUseCaseImpl) DeleteRemind(ctx context.Context, input DeleteRemindInput) error {
	slog.Debug("deleting remind",
		"remind_id", input.ID,
	)

	remindID, err := domain.RemindIDFromString(input.ID)
	if err != nil {
		return NewValidationError("id", err.Error())
	}

	if err := uc.repo.Delete(ctx, remindID); err != nil {
		if !errors.Is(err, domain.ErrRemindNotFound) {
			slog.Error("failed to delete remind",
				"error", err,
				"remind_id", input.ID,
			)

			return fmt.Errorf("%w: %v", ErrInternalError, err)
		}

		slog.Info("remind not found for deletion (idempotency)",
			"remind_id", input.ID,
		)
	}

	slog.Debug("remind deleted",
		"remind_id", input.ID,
	)

	return nil
}

func (uc *remindUseCaseImpl) CancelRemindByTaskID(ctx context.Context, input CancelRemindByTaskIDInput) error {
	slog.Debug("canceling reminds by task ID",
		"task_id", input.TaskID,
		"user_id", input.UserID,
	)

	taskID, err := domain.TaskIDFromString(input.TaskID)
	if err != nil {
		return NewValidationError("task_id", err.Error())
	}

	if _, err := domain.UserIDFromString(input.UserID); err != nil {
		return NewValidationError("user_id", err.Error())
	}

	deletedIDs, err := uc.repo.DeleteByTaskID(ctx, taskID)
	if err != nil {
		slog.Error("failed to cancel reminds by task ID",
			"error", err,
			"task_id", input.TaskID,
			"user_id", input.UserID,
		)

		return fmt.Errorf("%w: %v", ErrInternalError, err)
	}

	if uc.publisher != nil && len(deletedIDs) > 0 {
		remindIDStrings := make([]string, len(deletedIDs))
		for i, id := range deletedIDs {
			remindIDStrings[i] = id.String()
		}

		req := &throttlev1.CancelRemindRequest{
			TaskId:       input.TaskID,
			UserId:       input.UserID,
			DeletedCount: int64(len(deletedIDs)),
			CancelledAt:  timestamppb.Now(),
			RemindIds:    remindIDStrings,
		}
		if pubErr := uc.publisher.PublishRemindCancelled(ctx, req); pubErr != nil {
			slog.Error("failed to publish remind cancelled event",
				"task_id", input.TaskID,
				"error", pubErr.Error(),
			)
		}
	}

	slog.Info("reminds canceled by task ID",
		"task_id", input.TaskID,
		"user_id", input.UserID,
		"deleted_count", len(deletedIDs),
	)

	return nil
}

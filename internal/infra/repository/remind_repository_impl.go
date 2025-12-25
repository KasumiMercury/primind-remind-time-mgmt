package repository

import (
	"context"
	"errors"
	"log/slog"

	"gorm.io/gorm"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/domain"
)

type remindRepositoryImpl struct {
	db *gorm.DB
}

func NewRemindRepository(db *gorm.DB) domain.RemindRepository {
	return &remindRepositoryImpl{
		db: db,
	}
}

func (r *remindRepositoryImpl) Save(ctx context.Context, remind *domain.Remind) error {
	slog.Debug("saving remind to database",
		"remind_id", remind.ID().String(),
	)

	m := FromEntity(remind)

	result := r.db.WithContext(ctx).Create(m)
	if result.Error != nil {
		slog.Error("failed to save remind to database",
			"remind_id", remind.ID().String(),
			"error", result.Error,
		)

		return result.Error
	}

	slog.Debug("remind saved to database",
		"remind_id", remind.ID().String(),
	)

	return nil
}

func (r *remindRepositoryImpl) FindByID(ctx context.Context, id domain.RemindID) (*domain.Remind, error) {
	slog.Debug("finding remind by ID",
		"remind_id", id.String(),
	)

	var m RemindModel

	result := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&m)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			slog.Debug("remind not found",
				"remind_id", id.String(),
			)

			return nil, domain.ErrRemindNotFound
		}

		slog.Error("failed to find remind by ID",
			"remind_id", id.String(),
			"error", result.Error,
		)

		return nil, result.Error
	}

	slog.Debug("remind found",
		"remind_id", id.String(),
	)

	return m.ToEntity()
}

func (r *remindRepositoryImpl) FindByTaskID(ctx context.Context, taskID domain.TaskID) ([]*domain.Remind, error) {
	slog.Debug("finding reminds by task ID",
		"task_id", taskID.String(),
	)

	var models []RemindModel

	result := r.db.WithContext(ctx).Where("task_id = ?", taskID.String()).Order("time ASC").Find(&models)
	if result.Error != nil {
		slog.Error("failed to find reminds by task ID",
			"task_id", taskID.String(),
			"error", result.Error,
		)

		return nil, result.Error
	}

	reminds := make([]*domain.Remind, 0, len(models))
	for _, m := range models {
		remind, err := m.ToEntity()
		if err != nil {
			slog.Error("failed to convert model to entity",
				"remind_id", m.ID,
				"error", err,
			)

			return nil, err
		}

		reminds = append(reminds, remind)
	}

	slog.Debug("reminds found by task ID",
		"task_id", taskID.String(),
		"count", len(reminds),
	)

	return reminds, nil
}

func (r *remindRepositoryImpl) FindByTimeRange(ctx context.Context, timeRange domain.TimeRange) ([]*domain.Remind, error) {
	slog.Debug("finding reminds by time range",
		"start", timeRange.Start,
		"end", timeRange.End,
	)

	var models []RemindModel

	result := r.db.WithContext(ctx).
		Where("time >= ? AND time <= ?", timeRange.Start, timeRange.End).
		Order("time ASC").
		Find(&models)

	if result.Error != nil {
		slog.Error("failed to find reminds by time range",
			"start", timeRange.Start,
			"end", timeRange.End,
			"error", result.Error,
		)

		return nil, result.Error
	}

	reminds := make([]*domain.Remind, 0, len(models))
	for _, m := range models {
		remind, err := m.ToEntity()
		if err != nil {
			slog.Error("failed to convert model to entity",
				"remind_id", m.ID,
				"error", err,
			)

			return nil, err
		}

		reminds = append(reminds, remind)
	}

	slog.Debug("reminds found by time range",
		"count", len(reminds),
		"start", timeRange.Start,
		"end", timeRange.End,
	)

	return reminds, nil
}

func (r *remindRepositoryImpl) Update(ctx context.Context, remind *domain.Remind) error {
	slog.Debug("updating remind in database",
		"remind_id", remind.ID().String(),
	)

	m := FromEntity(remind)

	result := r.db.WithContext(ctx).Model(&RemindModel{}).Where("id = ?", m.ID).Updates(m)
	if result.Error != nil {
		slog.Error("failed to update remind in database",
			"remind_id", remind.ID().String(),
			"error", result.Error,
		)

		return result.Error
	}

	if result.RowsAffected == 0 {
		slog.Debug("remind not found for update",
			"remind_id", remind.ID().String(),
		)

		return domain.ErrRemindNotFound
	}

	slog.Debug("remind updated in database",
		"remind_id", remind.ID().String(),
	)

	return nil
}

func (r *remindRepositoryImpl) Delete(ctx context.Context, id domain.RemindID) error {
	slog.Debug("deleting remind from database",
		"remind_id", id.String(),
	)

	result := r.db.WithContext(ctx).Where("id = ?", id.String()).Delete(&RemindModel{})
	if result.Error != nil {
		slog.Error("failed to delete remind from database",
			"remind_id", id.String(),
			"error", result.Error,
		)

		return result.Error
	}

	if result.RowsAffected == 0 {
		slog.Debug("remind not found for deletion",
			"remind_id", id.String(),
		)

		return domain.ErrRemindNotFound
	}

	slog.Debug("remind deleted from database",
		"remind_id", id.String(),
	)

	return nil
}

func (r *remindRepositoryImpl) DeleteByTaskID(ctx context.Context, taskID domain.TaskID) ([]domain.RemindID, error) {
	slog.Debug("deleting reminds by task ID",
		"task_id", taskID.String(),
	)

	var models []RemindModel
	if err := r.db.WithContext(ctx).
		Select("id").
		Where("task_id = ?", taskID.String()).
		Find(&models).Error; err != nil {
		slog.Error("failed to find reminds by task ID",
			"task_id", taskID.String(),
			"error", err,
		)

		return nil, err
	}

	if len(models) == 0 {
		slog.Debug("no reminds found for task ID",
			"task_id", taskID.String(),
		)

		return nil, nil
	}

	ids := make([]domain.RemindID, len(models))
	for i, m := range models {
		id, err := domain.RemindIDFromString(m.ID)
		if err != nil {
			slog.Error("failed to parse remind ID",
				"id", m.ID,
				"error", err,
			)

			return nil, err
		}

		ids[i] = id
	}

	result := r.db.WithContext(ctx).Where("task_id = ?", taskID.String()).Delete(&RemindModel{})
	if result.Error != nil {
		slog.Error("failed to delete reminds by task ID",
			"task_id", taskID.String(),
			"error", result.Error,
		)

		return nil, result.Error
	}

	slog.Debug("reminds deleted by task ID",
		"task_id", taskID.String(),
		"count", len(ids),
	)

	return ids, nil
}

func (r *remindRepositoryImpl) WithTx(ctx context.Context, fn func(repo domain.RemindRepository) error) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		slog.Error("failed to begin transaction",
			"error", tx.Error,
		)

		return tx.Error
	}

	txRepo := &remindRepositoryImpl{db: tx}

	if err := fn(txRepo); err != nil {
		if rbErr := tx.Rollback().Error; rbErr != nil {
			slog.Error("failed to rollback transaction",
				"error", rbErr,
				"original_error", err,
			)
		}

		return err
	}

	if err := tx.Commit().Error; err != nil {
		slog.Error("failed to commit transaction",
			"error", err,
		)

		return err
	}

	return nil
}

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/domain"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/infra/repository"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/testutil"
)

func TestSaveSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	repo := repository.NewRemindRepository(testDB.DB)
	ctx := context.Background()

	tests := []struct {
		name        string
		deviceCount int
	}{
		{
			name:        "save remind with single device",
			deviceCount: 1,
		},
		{
			name:        "save remind with multiple devices",
			deviceCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			userID, err := domain.UserIDFromUUID(uuid.Must(uuid.NewV7()))
			require.NoError(t, err)
			taskID, err := domain.TaskIDFromUUID(uuid.Must(uuid.NewV7()))
			require.NoError(t, err)

			deviceSlice := make([]domain.Device, tt.deviceCount)
			for i := 0; i < tt.deviceCount; i++ {
				d, err := domain.NewDevice(
					"device-"+string(rune('a'+i)),
					"token-"+string(rune('a'+i)),
				)
				require.NoError(t, err)

				deviceSlice[i] = d
			}

			devices, err := domain.NewDevices(deviceSlice)
			require.NoError(t, err)

			remind, err := domain.NewRemind(
				time.Now().Add(1*time.Hour),
				userID,
				devices,
				taskID,
				domain.TypeNear,
			)
			require.NoError(t, err)

			err = repo.Save(ctx, remind)

			assert.NoError(t, err)

			found, err := repo.FindByID(ctx, remind.ID())
			assert.NoError(t, err)
			assert.Equal(t, remind.ID().String(), found.ID().String())
			assert.Equal(t, tt.deviceCount, found.Devices().Count())
		})
	}
}

func TestSaveError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	repo := repository.NewRemindRepository(testDB.DB)
	ctx := context.Background()

	tests := []struct {
		name string
	}{
		{
			name: "duplicate ID causes error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			userID, err := domain.UserIDFromUUID(uuid.Must(uuid.NewV7()))
			require.NoError(t, err)
			taskID, err := domain.TaskIDFromUUID(uuid.Must(uuid.NewV7()))
			require.NoError(t, err)

			d, err := domain.NewDevice("device", "token")
			require.NoError(t, err)
			devices, err := domain.NewDevices([]domain.Device{d})
			require.NoError(t, err)

			remind, err := domain.NewRemind(
				time.Now().Add(1*time.Hour),
				userID,
				devices,
				taskID,
				domain.TypeNear,
			)
			require.NoError(t, err)

			err = repo.Save(ctx, remind)
			require.NoError(t, err)

			err = repo.Save(ctx, remind)

			assert.Error(t, err)
		})
	}
}

func TestFindByIDSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	repo := repository.NewRemindRepository(testDB.DB)
	ctx := context.Background()

	tests := []struct {
		name      string
		throttled bool
	}{
		{
			name:      "find non-throttled remind",
			throttled: false,
		},
		{
			name:      "find throttled remind",
			throttled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			userID, err := domain.UserIDFromUUID(uuid.Must(uuid.NewV7()))
			require.NoError(t, err)
			taskID, err := domain.TaskIDFromUUID(uuid.Must(uuid.NewV7()))
			require.NoError(t, err)

			d, err := domain.NewDevice("device", "token")
			require.NoError(t, err)
			devices, err := domain.NewDevices([]domain.Device{d})
			require.NoError(t, err)

			remind := domain.Reconstitute(
				domain.NewRemindID(),
				time.Now().Add(1*time.Hour),
				userID,
				devices,
				taskID,
				domain.TypeNear,
				tt.throttled,
				time.Now().Add(-1*time.Hour),
				time.Now(),
			)

			err = repo.Save(ctx, remind)
			require.NoError(t, err)

			found, err := repo.FindByID(ctx, remind.ID())

			assert.NoError(t, err)
			assert.Equal(t, remind.ID().String(), found.ID().String())
			assert.Equal(t, remind.UserID().String(), found.UserID().String())
			assert.Equal(t, remind.TaskID().String(), found.TaskID().String())
			assert.Equal(t, remind.TaskType(), found.TaskType())
			assert.Equal(t, tt.throttled, found.IsThrottled())
		})
	}
}

func TestFindByIDError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	repo := repository.NewRemindRepository(testDB.DB)
	ctx := context.Background()

	tests := []struct {
		name string
	}{
		{
			name: "not found returns error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			nonExistentID := domain.NewRemindID()

			_, err := repo.FindByID(ctx, nonExistentID)

			assert.ErrorIs(t, err, domain.ErrRemindNotFound)
		})
	}
}

func TestFindByTaskIDSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	repo := repository.NewRemindRepository(testDB.DB)
	ctx := context.Background()

	tests := []struct {
		name          string
		remindsToSave int
	}{
		{
			name:          "find single remind by task ID",
			remindsToSave: 1,
		},
		{
			name:          "find multiple reminds by task ID",
			remindsToSave: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			userID, err := domain.UserIDFromUUID(uuid.Must(uuid.NewV7()))
			require.NoError(t, err)
			taskID, err := domain.TaskIDFromUUID(uuid.Must(uuid.NewV7()))
			require.NoError(t, err)

			d, err := domain.NewDevice("device", "token")
			require.NoError(t, err)
			devices, err := domain.NewDevices([]domain.Device{d})
			require.NoError(t, err)

			for i := 0; i < tt.remindsToSave; i++ {
				remindTime := time.Now().Add(time.Duration(i+1) * time.Hour).Truncate(time.Microsecond)
				remind := domain.Reconstitute(
					domain.NewRemindID(),
					remindTime,
					userID,
					devices,
					taskID,
					domain.TypeNear,
					false,
					time.Now().Add(-1*time.Hour),
					time.Now(),
				)
				err = repo.Save(ctx, remind)
				require.NoError(t, err)
			}

			found, err := repo.FindByTaskID(ctx, taskID)

			assert.NoError(t, err)
			assert.Len(t, found, tt.remindsToSave)

			for _, r := range found {
				assert.Equal(t, taskID.String(), r.TaskID().String())
			}
		})
	}
}

func TestFindByTaskIDEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	repo := repository.NewRemindRepository(testDB.DB)
	ctx := context.Background()

	tests := []struct {
		name string
	}{
		{
			name: "empty result for non-existent task ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			nonExistentTaskID, err := domain.TaskIDFromUUID(uuid.Must(uuid.NewV7()))
			require.NoError(t, err)

			found, err := repo.FindByTaskID(ctx, nonExistentTaskID)

			assert.NoError(t, err)
			assert.Empty(t, found)
		})
	}
}

func TestFindByTimeRangeSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	repo := repository.NewRemindRepository(testDB.DB)
	ctx := context.Background()

	tests := []struct {
		name          string
		remindsToSave int
		expectedCount int
	}{
		{
			name:          "find no reminds in empty range",
			remindsToSave: 0,
			expectedCount: 0,
		},
		{
			name:          "find single remind",
			remindsToSave: 1,
			expectedCount: 1,
		},
		{
			name:          "find multiple reminds",
			remindsToSave: 5,
			expectedCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			baseTime := time.Now().Add(1 * time.Hour).Truncate(time.Microsecond)
			startTime := baseTime.Add(-30 * time.Minute)
			endTime := baseTime.Add(30 * time.Minute)

			for i := 0; i < tt.remindsToSave; i++ {
				userID, err := domain.UserIDFromUUID(uuid.Must(uuid.NewV7()))
				require.NoError(t, err)
				taskID, err := domain.TaskIDFromUUID(uuid.Must(uuid.NewV7()))
				require.NoError(t, err)

				d, err := domain.NewDevice("device-"+string(rune('a'+i)), "token-"+string(rune('a'+i)))
				require.NoError(t, err)
				devices, err := domain.NewDevices([]domain.Device{d})
				require.NoError(t, err)

				remindTime := baseTime.Add(time.Duration(i) * time.Minute)
				remind := domain.Reconstitute(
					domain.NewRemindID(),
					remindTime,
					userID,
					devices,
					taskID,
					domain.TypeNear,
					false,
					time.Now().Add(-1*time.Hour),
					time.Now(),
				)

				err = repo.Save(ctx, remind)
				require.NoError(t, err)
			}

			timeRange := domain.TimeRange{Start: startTime, End: endTime}

			found, err := repo.FindByTimeRange(ctx, timeRange)

			assert.NoError(t, err)
			assert.Len(t, found, tt.expectedCount)
		})
	}
}

func TestFindByTimeRangeOrderSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	repo := repository.NewRemindRepository(testDB.DB)
	ctx := context.Background()

	tests := []struct {
		name string
	}{
		{
			name: "reminds are ordered by time ascending",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			baseTime := time.Now().Add(1 * time.Hour).Truncate(time.Microsecond)
			times := []time.Time{
				baseTime.Add(10 * time.Minute),
				baseTime.Add(5 * time.Minute),
				baseTime.Add(15 * time.Minute),
			}

			// Save in non-ordered manner
			for i, remindTime := range times {
				userID, err := domain.UserIDFromUUID(uuid.Must(uuid.NewV7()))
				require.NoError(t, err)
				taskID, err := domain.TaskIDFromUUID(uuid.Must(uuid.NewV7()))
				require.NoError(t, err)

				d, err := domain.NewDevice("device-"+string(rune('a'+i)), "token-"+string(rune('a'+i)))
				require.NoError(t, err)
				devices, err := domain.NewDevices([]domain.Device{d})
				require.NoError(t, err)

				remind := domain.Reconstitute(
					domain.NewRemindID(),
					remindTime,
					userID,
					devices,
					taskID,
					domain.TypeNear,
					false,
					time.Now().Add(-1*time.Hour),
					time.Now(),
				)

				err = repo.Save(ctx, remind)
				require.NoError(t, err)
			}

			timeRange := domain.TimeRange{Start: baseTime, End: baseTime.Add(20 * time.Minute)}

			found, err := repo.FindByTimeRange(ctx, timeRange)

			assert.NoError(t, err)
			assert.Len(t, found, 3)

			for i := 0; i < len(found)-1; i++ {
				assert.True(t, found[i].Time().Before(found[i+1].Time()) || found[i].Time().Equal(found[i+1].Time()))
			}
		})
	}
}

func TestUpdateSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	repo := repository.NewRemindRepository(testDB.DB)
	ctx := context.Background()

	tests := []struct {
		name string
	}{
		{
			name: "update throttled status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			userID, err := domain.UserIDFromUUID(uuid.Must(uuid.NewV7()))
			require.NoError(t, err)
			taskID, err := domain.TaskIDFromUUID(uuid.Must(uuid.NewV7()))
			require.NoError(t, err)

			d, err := domain.NewDevice("device", "token")
			require.NoError(t, err)
			devices, err := domain.NewDevices([]domain.Device{d})
			require.NoError(t, err)

			remind, err := domain.NewRemind(
				time.Now().Add(1*time.Hour),
				userID,
				devices,
				taskID,
				domain.TypeNear,
			)
			require.NoError(t, err)

			err = repo.Save(ctx, remind)
			require.NoError(t, err)

			err = remind.MarkAsThrottled()
			require.NoError(t, err)

			err = repo.Update(ctx, remind)

			assert.NoError(t, err)

			found, err := repo.FindByID(ctx, remind.ID())
			assert.NoError(t, err)
			assert.True(t, found.IsThrottled())
		})
	}
}

func TestUpdateNotFoundError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	repo := repository.NewRemindRepository(testDB.DB)
	ctx := context.Background()

	tests := []struct {
		name string
	}{
		{
			name: "update non-existent remind returns not found error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			userID, err := domain.UserIDFromUUID(uuid.Must(uuid.NewV7()))
			require.NoError(t, err)
			taskID, err := domain.TaskIDFromUUID(uuid.Must(uuid.NewV7()))
			require.NoError(t, err)

			d, err := domain.NewDevice("device", "token")
			require.NoError(t, err)
			devices, err := domain.NewDevices([]domain.Device{d})
			require.NoError(t, err)

			remind := domain.Reconstitute(
				domain.NewRemindID(),
				time.Now().Add(1*time.Hour),
				userID,
				devices,
				taskID,
				domain.TypeNear,
				false,
				time.Now().Add(-1*time.Hour),
				time.Now(),
			)

			err = repo.Update(ctx, remind)

			assert.ErrorIs(t, err, domain.ErrRemindNotFound)
		})
	}
}

func TestDeleteSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	repo := repository.NewRemindRepository(testDB.DB)
	ctx := context.Background()

	tests := []struct {
		name string
	}{
		{
			name: "delete existing remind",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			userID, err := domain.UserIDFromUUID(uuid.Must(uuid.NewV7()))
			require.NoError(t, err)
			taskID, err := domain.TaskIDFromUUID(uuid.Must(uuid.NewV7()))
			require.NoError(t, err)

			d, err := domain.NewDevice("device", "token")
			require.NoError(t, err)
			devices, err := domain.NewDevices([]domain.Device{d})
			require.NoError(t, err)

			remind, err := domain.NewRemind(
				time.Now().Add(1*time.Hour),
				userID,
				devices,
				taskID,
				domain.TypeNear,
			)
			require.NoError(t, err)

			err = repo.Save(ctx, remind)
			require.NoError(t, err)

			err = repo.Delete(ctx, remind.ID())

			assert.NoError(t, err)

			_, err = repo.FindByID(ctx, remind.ID())
			assert.ErrorIs(t, err, domain.ErrRemindNotFound)
		})
	}
}

func TestDeleteError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	repo := repository.NewRemindRepository(testDB.DB)
	ctx := context.Background()

	tests := []struct {
		name string
	}{
		{
			name: "delete non-existent remind returns error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			nonExistentID := domain.NewRemindID()

			err := repo.Delete(ctx, nonExistentID)

			assert.ErrorIs(t, err, domain.ErrRemindNotFound)
		})
	}
}

func TestWithTxCommitSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	repo := repository.NewRemindRepository(testDB.DB)
	ctx := context.Background()

	tests := []struct {
		name        string
		remindCount int
	}{
		{
			name:        "commit single remind",
			remindCount: 1,
		},
		{
			name:        "commit multiple reminds",
			remindCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			userID, err := domain.UserIDFromUUID(uuid.Must(uuid.NewV7()))
			require.NoError(t, err)
			taskID, err := domain.TaskIDFromUUID(uuid.Must(uuid.NewV7()))
			require.NoError(t, err)

			d, err := domain.NewDevice("device", "token")
			require.NoError(t, err)
			devices, err := domain.NewDevices([]domain.Device{d})
			require.NoError(t, err)

			reminds := make([]*domain.Remind, tt.remindCount)
			for i := 0; i < tt.remindCount; i++ {
				remind, err := domain.NewRemind(
					time.Now().Add(time.Duration(i+1)*time.Hour),
					userID,
					devices,
					taskID,
					domain.TypeNear,
				)
				require.NoError(t, err)

				reminds[i] = remind
			}

			err = repo.WithTx(ctx, func(txRepo domain.RemindRepository) error {
				for _, remind := range reminds {
					if err := txRepo.Save(ctx, remind); err != nil {
						return err
					}
				}

				return nil
			})

			assert.NoError(t, err)

			for _, remind := range reminds {
				found, err := repo.FindByID(ctx, remind.ID())
				assert.NoError(t, err)
				assert.Equal(t, remind.ID().String(), found.ID().String())
			}
		})
	}
}

func TestWithTxRollbackOnError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	repo := repository.NewRemindRepository(testDB.DB)
	ctx := context.Background()

	tests := []struct {
		name string
	}{
		{
			name: "rollback on error leaves no data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			userID, err := domain.UserIDFromUUID(uuid.Must(uuid.NewV7()))
			require.NoError(t, err)
			taskID, err := domain.TaskIDFromUUID(uuid.Must(uuid.NewV7()))
			require.NoError(t, err)

			d, err := domain.NewDevice("device", "token")
			require.NoError(t, err)
			devices, err := domain.NewDevices([]domain.Device{d})
			require.NoError(t, err)

			remind1, err := domain.NewRemind(
				time.Now().Add(1*time.Hour),
				userID,
				devices,
				taskID,
				domain.TypeNear,
			)
			require.NoError(t, err)

			remind2, err := domain.NewRemind(
				time.Now().Add(2*time.Hour),
				userID,
				devices,
				taskID,
				domain.TypeNear,
			)
			require.NoError(t, err)

			simulatedError := errors.New("simulated error")

			err = repo.WithTx(ctx, func(txRepo domain.RemindRepository) error {
				if err := txRepo.Save(ctx, remind1); err != nil {
					return err
				}

				if err := txRepo.Save(ctx, remind2); err != nil {
					return err
				}

				return simulatedError
			})

			assert.ErrorIs(t, err, simulatedError)

			_, err = repo.FindByID(ctx, remind1.ID())
			assert.ErrorIs(t, err, domain.ErrRemindNotFound)

			_, err = repo.FindByID(ctx, remind2.ID())
			assert.ErrorIs(t, err, domain.ErrRemindNotFound)
		})
	}
}

func TestWithTxRollbackOnSaveError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	repo := repository.NewRemindRepository(testDB.DB)
	ctx := context.Background()

	tests := []struct {
		name string
	}{
		{
			name: "rollback when second save fails due to duplicate ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			userID, err := domain.UserIDFromUUID(uuid.Must(uuid.NewV7()))
			require.NoError(t, err)
			taskID, err := domain.TaskIDFromUUID(uuid.Must(uuid.NewV7()))
			require.NoError(t, err)

			d, err := domain.NewDevice("device", "token")
			require.NoError(t, err)
			devices, err := domain.NewDevices([]domain.Device{d})
			require.NoError(t, err)

			remind1, err := domain.NewRemind(
				time.Now().Add(1*time.Hour),
				userID,
				devices,
				taskID,
				domain.TypeNear,
			)
			require.NoError(t, err)

			remind2, err := domain.NewRemind(
				time.Now().Add(2*time.Hour),
				userID,
				devices,
				taskID,
				domain.TypeNear,
			)
			require.NoError(t, err)

			err = repo.WithTx(ctx, func(txRepo domain.RemindRepository) error {
				if err := txRepo.Save(ctx, remind1); err != nil {
					return err
				}

				if err := txRepo.Save(ctx, remind2); err != nil {
					return err
				}

				// Try to save remind1 again - should fail due to duplicate primary key
				return txRepo.Save(ctx, remind1)
			})

			assert.Error(t, err)

			_, err = repo.FindByID(ctx, remind1.ID())
			assert.ErrorIs(t, err, domain.ErrRemindNotFound)

			_, err = repo.FindByID(ctx, remind2.ID())
			assert.ErrorIs(t, err, domain.ErrRemindNotFound)
		})
	}
}

func TestDeleteByTaskIDSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	repo := repository.NewRemindRepository(testDB.DB)
	ctx := context.Background()

	tests := []struct {
		name          string
		remindsToSave int
	}{
		{
			name:          "delete single remind by task ID",
			remindsToSave: 1,
		},
		{
			name:          "delete multiple reminds by task ID",
			remindsToSave: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			userID, err := domain.UserIDFromUUID(uuid.Must(uuid.NewV7()))
			require.NoError(t, err)
			taskID, err := domain.TaskIDFromUUID(uuid.Must(uuid.NewV7()))
			require.NoError(t, err)

			d, err := domain.NewDevice("device", "token")
			require.NoError(t, err)
			devices, err := domain.NewDevices([]domain.Device{d})
			require.NoError(t, err)

			for i := 0; i < tt.remindsToSave; i++ {
				remindTime := time.Now().Add(time.Duration(i+1) * time.Hour).Truncate(time.Microsecond)
				remind := domain.Reconstitute(
					domain.NewRemindID(),
					remindTime,
					userID,
					devices,
					taskID,
					domain.TypeNear,
					false,
					time.Now().Add(-1*time.Hour),
					time.Now(),
				)
				err = repo.Save(ctx, remind)
				require.NoError(t, err)
			}

			deletedCount, err := repo.DeleteByTaskID(ctx, taskID)

			assert.NoError(t, err)
			assert.Equal(t, int64(tt.remindsToSave), deletedCount)

			found, err := repo.FindByTaskID(ctx, taskID)
			assert.NoError(t, err)
			assert.Empty(t, found)
		})
	}
}

func TestDeleteByTaskIDNoRecords(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	repo := repository.NewRemindRepository(testDB.DB)
	ctx := context.Background()

	tests := []struct {
		name string
	}{
		{
			name: "delete non-existent task ID returns zero count without error (idempotency)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			nonExistentTaskID, err := domain.TaskIDFromUUID(uuid.Must(uuid.NewV7()))
			require.NoError(t, err)

			deletedCount, err := repo.DeleteByTaskID(ctx, nonExistentTaskID)

			assert.NoError(t, err)
			assert.Equal(t, int64(0), deletedCount)
		})
	}
}

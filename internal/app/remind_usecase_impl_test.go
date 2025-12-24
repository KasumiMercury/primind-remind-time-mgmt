package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/app"
	throttlev1 "github.com/KasumiMercury/primind-remind-time-mgmt/internal/gen/throttle/v1"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/infra/pubsub"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/infra/repository"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/testutil"
)

func generateUUIDv7String() string {
	return uuid.Must(uuid.NewV7()).String()
}

func setupUseCaseTest(t *testing.T) (app.RemindUseCase, func()) {
	t.Helper()
	testDB := testutil.SetupTestDB(t)
	repo := repository.NewRemindRepository(testDB.DB)
	useCase := app.NewRemindUseCase(repo, nil)

	return useCase, func() {
		testDB.CleanTable(t)
		testDB.TeardownTestDB(t)
	}
}

func TestCreateRemindSuccess(t *testing.T) {
	tests := []struct {
		name        string
		deviceCount int
		timesCount  int
		taskType    string
	}{
		{
			name:        "valid remind with single device and single time",
			deviceCount: 1,
			timesCount:  1,
			taskType:    "near",
		},
		{
			name:        "valid remind with multiple devices and single time",
			deviceCount: 3,
			timesCount:  1,
			taskType:    "short",
		},
		{
			name:        "valid remind with multiple times",
			deviceCount: 1,
			timesCount:  3,
			taskType:    "scheduled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, cleanup := setupUseCaseTest(t)
			defer cleanup()

			devices := make([]app.DeviceInput, tt.deviceCount)
			for i := 0; i < tt.deviceCount; i++ {
				devices[i] = app.DeviceInput{
					DeviceID: "device-" + string(rune('a'+i)),
					FCMToken: "token-" + string(rune('a'+i)),
				}
			}

			times := make([]time.Time, tt.timesCount)
			for i := 0; i < tt.timesCount; i++ {
				times[i] = time.Now().Add(time.Duration(i+1) * time.Hour)
			}

			input := app.CreateRemindInput{
				Times:    times,
				UserID:   generateUUIDv7String(),
				Devices:  devices,
				TaskID:   generateUUIDv7String(),
				TaskType: tt.taskType,
			}

			output, err := useCase.CreateRemind(context.Background(), input)

			assert.NoError(t, err)
			assert.Equal(t, int32(tt.timesCount), output.Count)
			assert.Len(t, output.Reminds, tt.timesCount)

			for _, remind := range output.Reminds {
				assert.NotEmpty(t, remind.ID)
				assert.Equal(t, input.UserID, remind.UserID)
				assert.Equal(t, input.TaskID, remind.TaskID)
				assert.Equal(t, tt.taskType, remind.TaskType)
				assert.False(t, remind.Throttled)
				assert.Len(t, remind.Devices, tt.deviceCount)
			}
		})
	}
}

func TestCreateRemindIdempotencySuccess(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "duplicate TaskID returns existing reminds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, cleanup := setupUseCaseTest(t)
			defer cleanup()

			taskID := generateUUIDv7String()

			input := app.CreateRemindInput{
				Times:    []time.Time{time.Now().Add(1 * time.Hour), time.Now().Add(2 * time.Hour)},
				UserID:   generateUUIDv7String(),
				Devices:  []app.DeviceInput{{DeviceID: "device-1", FCMToken: "token-1"}},
				TaskID:   taskID,
				TaskType: "near",
			}

			output1, err := useCase.CreateRemind(context.Background(), input)
			require.NoError(t, err)
			require.Equal(t, int32(2), output1.Count)

			input.UserID = generateUUIDv7String()
			input.Times = []time.Time{time.Now().Add(3 * time.Hour)}
			output2, err := useCase.CreateRemind(context.Background(), input)

			assert.NoError(t, err)
			assert.Equal(t, output1.Count, output2.Count)
			assert.Equal(t, output1.Reminds[0].ID, output2.Reminds[0].ID)
		})
	}
}

func TestCreateRemindError(t *testing.T) {
	tests := []struct {
		name          string
		input         app.CreateRemindInput
		expectedField string
	}{
		{
			name: "empty times array",
			input: app.CreateRemindInput{
				Times:    []time.Time{},
				UserID:   generateUUIDv7String(),
				Devices:  []app.DeviceInput{{DeviceID: "d", FCMToken: "t"}},
				TaskID:   generateUUIDv7String(),
				TaskType: "near",
			},
			expectedField: "times",
		},
		{
			name: "invalid user_id not UUIDv7",
			input: app.CreateRemindInput{
				Times:    []time.Time{time.Now().Add(1 * time.Hour)},
				UserID:   uuid.New().String(),
				Devices:  []app.DeviceInput{{DeviceID: "d", FCMToken: "t"}},
				TaskID:   generateUUIDv7String(),
				TaskType: "near",
			},
			expectedField: "user_id",
		},
		{
			name: "invalid task_id not UUIDv7",
			input: app.CreateRemindInput{
				Times:    []time.Time{time.Now().Add(1 * time.Hour)},
				UserID:   generateUUIDv7String(),
				Devices:  []app.DeviceInput{{DeviceID: "d", FCMToken: "t"}},
				TaskID:   uuid.New().String(),
				TaskType: "near",
			},
			expectedField: "task_id",
		},
		{
			name: "empty device ID",
			input: app.CreateRemindInput{
				Times:    []time.Time{time.Now().Add(1 * time.Hour)},
				UserID:   generateUUIDv7String(),
				Devices:  []app.DeviceInput{{DeviceID: "", FCMToken: "t"}},
				TaskID:   generateUUIDv7String(),
				TaskType: "near",
			},
			expectedField: "devices[0]",
		},
		{
			name: "empty FCM token",
			input: app.CreateRemindInput{
				Times:    []time.Time{time.Now().Add(1 * time.Hour)},
				UserID:   generateUUIDv7String(),
				Devices:  []app.DeviceInput{{DeviceID: "d", FCMToken: ""}},
				TaskID:   generateUUIDv7String(),
				TaskType: "near",
			},
			expectedField: "devices[0]",
		},
		{
			name: "empty devices array",
			input: app.CreateRemindInput{
				Times:    []time.Time{time.Now().Add(1 * time.Hour)},
				UserID:   generateUUIDv7String(),
				Devices:  []app.DeviceInput{},
				TaskID:   generateUUIDv7String(),
				TaskType: "near",
			},
			expectedField: "devices",
		},
		{
			name: "empty task type",
			input: app.CreateRemindInput{
				Times:    []time.Time{time.Now().Add(1 * time.Hour)},
				UserID:   generateUUIDv7String(),
				Devices:  []app.DeviceInput{{DeviceID: "d", FCMToken: "t"}},
				TaskID:   generateUUIDv7String(),
				TaskType: "",
			},
			expectedField: "task_type",
		},
		{
			name: "invalid task type",
			input: app.CreateRemindInput{
				Times:    []time.Time{time.Now().Add(1 * time.Hour)},
				UserID:   generateUUIDv7String(),
				Devices:  []app.DeviceInput{{DeviceID: "d", FCMToken: "t"}},
				TaskID:   generateUUIDv7String(),
				TaskType: "invalid_type",
			},
			expectedField: "task_type",
		},
		{
			name: "past remind time",
			input: app.CreateRemindInput{
				Times:    []time.Time{time.Now().Add(-2 * time.Minute)},
				UserID:   generateUUIDv7String(),
				Devices:  []app.DeviceInput{{DeviceID: "d", FCMToken: "t"}},
				TaskID:   generateUUIDv7String(),
				TaskType: "near",
			},
			expectedField: "times[0]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, cleanup := setupUseCaseTest(t)
			defer cleanup()

			_, err := useCase.CreateRemind(context.Background(), tt.input)

			assert.Error(t, err)
			assert.True(t, app.IsValidationError(err))

			var validationErr *app.ValidationError
			if errors.As(err, &validationErr) {
				assert.Equal(t, tt.expectedField, validationErr.Field)
			}
		})
	}
}

func TestGetRemindsByTimeRangeSuccess(t *testing.T) {
	tests := []struct {
		name          string
		setupReminds  int
		queryDuration time.Duration
		expectedCount int
	}{
		{
			name:          "empty result",
			setupReminds:  0,
			queryDuration: 2 * time.Hour,
			expectedCount: 0,
		},
		{
			name:          "single result",
			setupReminds:  1,
			queryDuration: 2 * time.Hour,
			expectedCount: 1,
		},
		{
			name:          "multiple results",
			setupReminds:  5,
			queryDuration: 6 * time.Hour,
			expectedCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, cleanup := setupUseCaseTest(t)
			defer cleanup()

			now := time.Now()

			// Setup test data
			for i := 0; i < tt.setupReminds; i++ {
				input := app.CreateRemindInput{
					Times:    []time.Time{now.Add(time.Duration(i+1) * time.Hour)},
					UserID:   generateUUIDv7String(),
					Devices:  []app.DeviceInput{{DeviceID: "d", FCMToken: "t"}},
					TaskID:   generateUUIDv7String(),
					TaskType: "near",
				}
				_, err := useCase.CreateRemind(context.Background(), input)
				require.NoError(t, err)
			}

			input := app.GetRemindsByTimeRangeInput{
				Start: now,
				End:   now.Add(tt.queryDuration),
			}

			output, err := useCase.GetRemindsByTimeRange(context.Background(), input)

			assert.NoError(t, err)
			assert.Equal(t, int32(tt.expectedCount), output.Count)
			assert.Len(t, output.Reminds, tt.expectedCount)
		})
	}
}

func TestGetRemindsByTimeRangePartialSuccess(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "returns only reminds within range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, cleanup := setupUseCaseTest(t)
			defer cleanup()

			now := time.Now()

			offsets := []time.Duration{1 * time.Hour, 2 * time.Hour, 5 * time.Hour}
			for _, offset := range offsets {
				input := app.CreateRemindInput{
					Times:    []time.Time{now.Add(offset)},
					UserID:   generateUUIDv7String(),
					Devices:  []app.DeviceInput{{DeviceID: "d", FCMToken: "t"}},
					TaskID:   generateUUIDv7String(),
					TaskType: "near",
				}
				_, err := useCase.CreateRemind(context.Background(), input)
				require.NoError(t, err)
			}

			input := app.GetRemindsByTimeRangeInput{
				Start: now,
				End:   now.Add(3 * time.Hour),
			}

			output, err := useCase.GetRemindsByTimeRange(context.Background(), input)

			assert.NoError(t, err)
			assert.Equal(t, int32(2), output.Count)
		})
	}
}

func TestGetRemindsByTimeRangeError(t *testing.T) {
	tests := []struct {
		name  string
		input app.GetRemindsByTimeRangeInput
	}{
		{
			name: "start after end",
			input: app.GetRemindsByTimeRangeInput{
				Start: time.Now().Add(2 * time.Hour),
				End:   time.Now(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, cleanup := setupUseCaseTest(t)
			defer cleanup()

			_, err := useCase.GetRemindsByTimeRange(context.Background(), tt.input)

			assert.Error(t, err)
			assert.True(t, app.IsValidationError(err))
		})
	}
}

func TestUpdateThrottledSuccess(t *testing.T) {
	tests := []struct {
		name      string
		throttled bool
	}{
		{
			name:      "set throttled to true",
			throttled: true,
		},
		{
			name:      "set throttled to false (no-op on new remind)",
			throttled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, cleanup := setupUseCaseTest(t)
			defer cleanup()

			createInput := app.CreateRemindInput{
				Times:    []time.Time{time.Now().Add(1 * time.Hour)},
				UserID:   generateUUIDv7String(),
				Devices:  []app.DeviceInput{{DeviceID: "d", FCMToken: "t"}},
				TaskID:   generateUUIDv7String(),
				TaskType: "near",
			}
			created, err := useCase.CreateRemind(context.Background(), createInput)
			require.NoError(t, err)
			require.Equal(t, int32(1), created.Count)

			input := app.UpdateThrottledInput{
				ID:        created.Reminds[0].ID,
				Throttled: tt.throttled,
			}

			output, err := useCase.UpdateThrottled(context.Background(), input)

			assert.NoError(t, err)
			assert.Equal(t, tt.throttled, output.Throttled)
		})
	}
}

func TestUpdateThrottledIdempotencySuccess(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "double throttling succeeds (idempotent)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, cleanup := setupUseCaseTest(t)
			defer cleanup()

			createInput := app.CreateRemindInput{
				Times:    []time.Time{time.Now().Add(1 * time.Hour)},
				UserID:   generateUUIDv7String(),
				Devices:  []app.DeviceInput{{DeviceID: "d", FCMToken: "t"}},
				TaskID:   generateUUIDv7String(),
				TaskType: "near",
			}
			created, err := useCase.CreateRemind(context.Background(), createInput)
			require.NoError(t, err)
			require.Equal(t, int32(1), created.Count)

			input := app.UpdateThrottledInput{ID: created.Reminds[0].ID, Throttled: true}

			_, err = useCase.UpdateThrottled(context.Background(), input)
			assert.NoError(t, err)

			output, err := useCase.UpdateThrottled(context.Background(), input)

			assert.NoError(t, err)
			assert.True(t, output.Throttled)
		})
	}
}

func TestUpdateThrottledError(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		expectedErr error
	}{
		{
			name:        "invalid ID format",
			id:          "not-a-uuid",
			expectedErr: nil,
		},
		{
			name:        "non-existent ID",
			id:          uuid.New().String(),
			expectedErr: app.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, cleanup := setupUseCaseTest(t)
			defer cleanup()

			input := app.UpdateThrottledInput{ID: tt.id, Throttled: true}

			_, err := useCase.UpdateThrottled(context.Background(), input)

			assert.Error(t, err)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.True(t, app.IsValidationError(err))
			}
		})
	}
}

func TestDeleteRemindSuccess(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "delete existing remind",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, cleanup := setupUseCaseTest(t)
			defer cleanup()

			createInput := app.CreateRemindInput{
				Times:    []time.Time{time.Now().Add(1 * time.Hour)},
				UserID:   generateUUIDv7String(),
				Devices:  []app.DeviceInput{{DeviceID: "d", FCMToken: "t"}},
				TaskID:   generateUUIDv7String(),
				TaskType: "near",
			}
			created, err := useCase.CreateRemind(context.Background(), createInput)
			require.NoError(t, err)
			require.Equal(t, int32(1), created.Count)

			input := app.DeleteRemindInput{ID: created.Reminds[0].ID}

			err = useCase.DeleteRemind(context.Background(), input)

			assert.NoError(t, err)
		})
	}
}

func TestDeleteRemindIdempotencySuccess(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "double delete succeeds (idempotent)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, cleanup := setupUseCaseTest(t)
			defer cleanup()

			createInput := app.CreateRemindInput{
				Times:    []time.Time{time.Now().Add(1 * time.Hour)},
				UserID:   generateUUIDv7String(),
				Devices:  []app.DeviceInput{{DeviceID: "d", FCMToken: "t"}},
				TaskID:   generateUUIDv7String(),
				TaskType: "near",
			}
			created, err := useCase.CreateRemind(context.Background(), createInput)
			require.NoError(t, err)
			require.Equal(t, int32(1), created.Count)

			input := app.DeleteRemindInput{ID: created.Reminds[0].ID}

			err = useCase.DeleteRemind(context.Background(), input)
			assert.NoError(t, err)

			err = useCase.DeleteRemind(context.Background(), input)

			assert.NoError(t, err)
		})
	}
}

func TestDeleteRemindNonExistentSuccess(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "delete non-existent remind succeeds (idempotent)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, cleanup := setupUseCaseTest(t)
			defer cleanup()

			input := app.DeleteRemindInput{ID: uuid.New().String()}

			err := useCase.DeleteRemind(context.Background(), input)

			assert.NoError(t, err)
		})
	}
}

func TestDeleteRemindError(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{
			name: "invalid ID format",
			id:   "not-a-uuid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, cleanup := setupUseCaseTest(t)
			defer cleanup()

			input := app.DeleteRemindInput{ID: tt.id}

			err := useCase.DeleteRemind(context.Background(), input)

			assert.Error(t, err)
			assert.True(t, app.IsValidationError(err))
		})
	}
}

func TestCreateRemindTransactionCommitSuccess(t *testing.T) {
	tests := []struct {
		name       string
		timesCount int
	}{
		{
			name:       "all reminds persisted on successful commit",
			timesCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, cleanup := setupUseCaseTest(t)
			defer cleanup()

			times := make([]time.Time, tt.timesCount)
			for i := 0; i < tt.timesCount; i++ {
				times[i] = time.Now().Add(time.Duration(i+1) * time.Hour)
			}

			input := app.CreateRemindInput{
				Times:    times,
				UserID:   generateUUIDv7String(),
				Devices:  []app.DeviceInput{{DeviceID: "device-1", FCMToken: "token-1"}},
				TaskID:   generateUUIDv7String(),
				TaskType: "near",
			}

			output, err := useCase.CreateRemind(context.Background(), input)

			require.NoError(t, err)
			assert.Equal(t, int32(tt.timesCount), output.Count)

			rangeInput := app.GetRemindsByTimeRangeInput{
				Start: time.Now(),
				End:   time.Now().Add(time.Duration(tt.timesCount+1) * time.Hour),
			}
			rangeOutput, err := useCase.GetRemindsByTimeRange(context.Background(), rangeInput)
			require.NoError(t, err)
			assert.Equal(t, int32(tt.timesCount), rangeOutput.Count)
		})
	}
}

func TestCancelRemindByTaskIDSuccess(t *testing.T) {
	tests := []struct {
		name       string
		timesCount int
	}{
		{
			name:       "cancel single remind by task ID",
			timesCount: 1,
		},
		{
			name:       "cancel multiple reminds by task ID",
			timesCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, cleanup := setupUseCaseTest(t)
			defer cleanup()

			taskID := generateUUIDv7String()
			userID := generateUUIDv7String()

			times := make([]time.Time, tt.timesCount)
			for i := 0; i < tt.timesCount; i++ {
				times[i] = time.Now().Add(time.Duration(i+1) * time.Hour)
			}

			createInput := app.CreateRemindInput{
				Times:    times,
				UserID:   userID,
				Devices:  []app.DeviceInput{{DeviceID: "d", FCMToken: "t"}},
				TaskID:   taskID,
				TaskType: "near",
			}
			created, err := useCase.CreateRemind(context.Background(), createInput)
			require.NoError(t, err)
			require.Equal(t, int32(tt.timesCount), created.Count)

			input := app.CancelRemindByTaskIDInput{
				TaskID: taskID,
				UserID: userID,
			}

			err = useCase.CancelRemindByTaskID(context.Background(), input)

			assert.NoError(t, err)

			rangeInput := app.GetRemindsByTimeRangeInput{
				Start: time.Now(),
				End:   time.Now().Add(time.Duration(tt.timesCount+1) * time.Hour),
			}
			rangeOutput, err := useCase.GetRemindsByTimeRange(context.Background(), rangeInput)
			require.NoError(t, err)
			assert.Equal(t, int32(0), rangeOutput.Count)
		})
	}
}

func TestCancelRemindByTaskIDIdempotencySuccess(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "double cancel succeeds (idempotent)",
		},
		{
			name: "cancel non-existent task ID succeeds (idempotent)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, cleanup := setupUseCaseTest(t)
			defer cleanup()

			taskID := generateUUIDv7String()
			userID := generateUUIDv7String()

			if tt.name == "double cancel succeeds (idempotent)" {
				createInput := app.CreateRemindInput{
					Times:    []time.Time{time.Now().Add(1 * time.Hour)},
					UserID:   userID,
					Devices:  []app.DeviceInput{{DeviceID: "d", FCMToken: "t"}},
					TaskID:   taskID,
					TaskType: "near",
				}
				_, err := useCase.CreateRemind(context.Background(), createInput)
				require.NoError(t, err)
			}

			input := app.CancelRemindByTaskIDInput{
				TaskID: taskID,
				UserID: userID,
			}

			err := useCase.CancelRemindByTaskID(context.Background(), input)
			assert.NoError(t, err)

			err = useCase.CancelRemindByTaskID(context.Background(), input)

			assert.NoError(t, err)
		})
	}
}

func TestCancelRemindByTaskIDError(t *testing.T) {
	tests := []struct {
		name          string
		taskID        string
		userID        string
		expectedField string
	}{
		{
			name:          "invalid task_id format",
			taskID:        "not-a-uuid",
			userID:        generateUUIDv7String(),
			expectedField: "task_id",
		},
		{
			name:          "invalid task_id not UUIDv7",
			taskID:        uuid.New().String(),
			userID:        generateUUIDv7String(),
			expectedField: "task_id",
		},
		{
			name:          "invalid user_id format",
			taskID:        generateUUIDv7String(),
			userID:        "not-a-uuid",
			expectedField: "user_id",
		},
		{
			name:          "invalid user_id not UUIDv7",
			taskID:        generateUUIDv7String(),
			userID:        uuid.New().String(),
			expectedField: "user_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, cleanup := setupUseCaseTest(t)
			defer cleanup()

			input := app.CancelRemindByTaskIDInput{
				TaskID: tt.taskID,
				UserID: tt.userID,
			}

			err := useCase.CancelRemindByTaskID(context.Background(), input)

			assert.Error(t, err)
			assert.True(t, app.IsValidationError(err))

			var validationErr *app.ValidationError
			if errors.As(err, &validationErr) {
				assert.Equal(t, tt.expectedField, validationErr.Field)
			}
		})
	}
}

// setupUseCaseTestWithPublisher creates a use case with a custom publisher for testing.
func setupUseCaseTestWithPublisher(t *testing.T, publisher pubsub.Publisher) (app.RemindUseCase, func()) {
	t.Helper()
	testDB := testutil.SetupTestDB(t)
	repo := repository.NewRemindRepository(testDB.DB)
	useCase := app.NewRemindUseCase(repo, publisher)

	return useCase, func() {
		testDB.CleanTable(t)
		testDB.TeardownTestDB(t)
	}
}

func TestCancelRemindByTaskID_PublishesEvent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPublisher := pubsub.NewMockPublisher(ctrl)

	// Expect PublishRemindCancelled to be called with matching request
	mockPublisher.EXPECT().
		PublishRemindCancelled(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, req *throttlev1.CancelRemindRequest) error {
			assert.NotEmpty(t, req.GetTaskId())
			assert.NotEmpty(t, req.GetUserId())
			assert.Greater(t, req.GetDeletedCount(), int64(0))
			assert.NotNil(t, req.GetCancelledAt())
			return nil
		}).
		Times(1)

	useCase, cleanup := setupUseCaseTestWithPublisher(t, mockPublisher)
	defer cleanup()

	// Create a remind first
	taskID := generateUUIDv7String()
	userID := generateUUIDv7String()

	createInput := app.CreateRemindInput{
		Times:    []time.Time{time.Now().Add(1 * time.Hour)},
		UserID:   userID,
		Devices:  []app.DeviceInput{{DeviceID: "device-a", FCMToken: "token-a"}},
		TaskID:   taskID,
		TaskType: "near",
	}
	_, err := useCase.CreateRemind(context.Background(), createInput)
	require.NoError(t, err)

	// Cancel the remind - should trigger publish
	input := app.CancelRemindByTaskIDInput{
		TaskID: taskID,
		UserID: userID,
	}
	err = useCase.CancelRemindByTaskID(context.Background(), input)
	assert.NoError(t, err)
}

func TestCancelRemindByTaskID_PublishError_DoesNotFailOperation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPublisher := pubsub.NewMockPublisher(ctrl)

	mockPublisher.EXPECT().
		PublishRemindCancelled(gomock.Any(), gomock.Any()).
		Return(errors.New("publish failed")).
		Times(1)

	useCase, cleanup := setupUseCaseTestWithPublisher(t, mockPublisher)
	defer cleanup()

	// Create a remind first
	taskID := generateUUIDv7String()
	userID := generateUUIDv7String()

	createInput := app.CreateRemindInput{
		Times:    []time.Time{time.Now().Add(1 * time.Hour)},
		UserID:   userID,
		Devices:  []app.DeviceInput{{DeviceID: "device-a", FCMToken: "token-a"}},
		TaskID:   taskID,
		TaskType: "near",
	}
	_, err := useCase.CreateRemind(context.Background(), createInput)
	require.NoError(t, err)

	// Even though publish fails, operation should succeed
	input := app.CancelRemindByTaskIDInput{
		TaskID: taskID,
		UserID: userID,
	}
	err = useCase.CancelRemindByTaskID(context.Background(), input)
	assert.NoError(t, err)
}

func TestCancelRemindByTaskID_NoRemindsDeleted_DoesNotPublish(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPublisher := pubsub.NewMockPublisher(ctrl)

	// PublishRemindCancelled should NOT be called when deletedCount = 0
	mockPublisher.EXPECT().
		PublishRemindCancelled(gomock.Any(), gomock.Any()).
		Times(0)

	useCase, cleanup := setupUseCaseTestWithPublisher(t, mockPublisher)
	defer cleanup()

	// Cancel non-existent reminds
	input := app.CancelRemindByTaskIDInput{
		TaskID: generateUUIDv7String(),
		UserID: generateUUIDv7String(),
	}
	err := useCase.CancelRemindByTaskID(context.Background(), input)
	assert.NoError(t, err)
}

func TestCancelRemindByTaskID_NilPublisher_Succeeds(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// This tests the existing behavior with nil publisher
	useCase, cleanup := setupUseCaseTest(t)
	defer cleanup()

	taskID := generateUUIDv7String()
	userID := generateUUIDv7String()

	createInput := app.CreateRemindInput{
		Times:    []time.Time{time.Now().Add(1 * time.Hour)},
		UserID:   userID,
		Devices:  []app.DeviceInput{{DeviceID: "device-a", FCMToken: "token-a"}},
		TaskID:   taskID,
		TaskType: "near",
	}
	_, err := useCase.CreateRemind(context.Background(), createInput)
	require.NoError(t, err)

	input := app.CancelRemindByTaskIDInput{
		TaskID: taskID,
		UserID: userID,
	}
	err = useCase.CancelRemindByTaskID(context.Background(), input)
	assert.NoError(t, err)
}

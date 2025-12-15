package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/app"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/domain"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/infra/handler"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/infra/repository"
	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/testutil"
)

func setupTestRouter(t *testing.T, testDB *testutil.TestDB) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	repo := repository.NewRemindRepository(testDB.DB)
	useCase := app.NewRemindUseCase(repo)
	h := handler.NewRemindHandler(useCase)

	router := gin.New()
	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	return router
}

func TestCreateRemindHandlerSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	router := setupTestRouter(t, testDB)

	tests := []struct {
		name        string
		deviceCount int
		timesCount  int
	}{
		{
			name:        "create remind with single device and single time",
			deviceCount: 1,
			timesCount:  1,
		},
		{
			name:        "create remind with multiple devices and single time",
			deviceCount: 3,
			timesCount:  1,
		},
		{
			name:        "create remind with multiple times",
			deviceCount: 1,
			timesCount:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			devices := make([]map[string]string, tt.deviceCount)
			for i := 0; i < tt.deviceCount; i++ {
				devices[i] = map[string]string{
					"device_id": "device-" + string(rune('a'+i)),
					"fcm_token": "token-" + string(rune('a'+i)),
				}
			}

			times := make([]string, tt.timesCount)
			for i := 0; i < tt.timesCount; i++ {
				times[i] = time.Now().Add(time.Duration(i+1) * time.Hour).Format(time.RFC3339)
			}

			reqBody := map[string]any{
				"times":     times,
				"user_id":   uuid.Must(uuid.NewV7()).String(),
				"devices":   devices,
				"task_id":   uuid.Must(uuid.NewV7()).String(),
				"task_type": "normal",
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/reminds", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusCreated, rec.Code)

			var response handler.RemindsResponse

			err := json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.timesCount, response.Count)
			assert.Len(t, response.Reminds, tt.timesCount)

			for _, remind := range response.Reminds {
				assert.NotEmpty(t, remind.ID)
				assert.Len(t, remind.Devices, tt.deviceCount)
				assert.False(t, remind.Throttled)
			}
		})
	}
}

func TestCreateRemindHandlerError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	router := setupTestRouter(t, testDB)

	tests := []struct {
		name           string
		requestBody    map[string]any
		expectedStatus int
	}{
		{
			name: "missing times",
			requestBody: map[string]any{
				"user_id":   uuid.Must(uuid.NewV7()).String(),
				"devices":   []map[string]string{{"device_id": "d", "fcm_token": "t"}},
				"task_id":   uuid.Must(uuid.NewV7()).String(),
				"task_type": "normal",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty times",
			requestBody: map[string]any{
				"times":     []string{},
				"user_id":   uuid.Must(uuid.NewV7()).String(),
				"devices":   []map[string]string{{"device_id": "d", "fcm_token": "t"}},
				"task_id":   uuid.Must(uuid.NewV7()).String(),
				"task_type": "normal",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing user_id",
			requestBody: map[string]any{
				"times":     []string{time.Now().Add(1 * time.Hour).Format(time.RFC3339)},
				"devices":   []map[string]string{{"device_id": "d", "fcm_token": "t"}},
				"task_id":   uuid.Must(uuid.NewV7()).String(),
				"task_type": "normal",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing devices",
			requestBody: map[string]any{
				"times":     []string{time.Now().Add(1 * time.Hour).Format(time.RFC3339)},
				"user_id":   uuid.Must(uuid.NewV7()).String(),
				"task_id":   uuid.Must(uuid.NewV7()).String(),
				"task_type": "normal",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty devices",
			requestBody: map[string]any{
				"times":     []string{time.Now().Add(1 * time.Hour).Format(time.RFC3339)},
				"user_id":   uuid.Must(uuid.NewV7()).String(),
				"devices":   []map[string]string{},
				"task_id":   uuid.Must(uuid.NewV7()).String(),
				"task_type": "normal",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing task_id",
			requestBody: map[string]any{
				"times":     []string{time.Now().Add(1 * time.Hour).Format(time.RFC3339)},
				"user_id":   uuid.Must(uuid.NewV7()).String(),
				"devices":   []map[string]string{{"device_id": "d", "fcm_token": "t"}},
				"task_type": "normal",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing task_type",
			requestBody: map[string]any{
				"times":   []string{time.Now().Add(1 * time.Hour).Format(time.RFC3339)},
				"user_id": uuid.Must(uuid.NewV7()).String(),
				"devices": []map[string]string{{"device_id": "d", "fcm_token": "t"}},
				"task_id": uuid.Must(uuid.NewV7()).String(),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid user_id format",
			requestBody: map[string]any{
				"times":     []string{time.Now().Add(1 * time.Hour).Format(time.RFC3339)},
				"user_id":   "not-a-uuid",
				"devices":   []map[string]string{{"device_id": "d", "fcm_token": "t"}},
				"task_id":   uuid.Must(uuid.NewV7()).String(),
				"task_type": "normal",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid task_id format",
			requestBody: map[string]any{
				"times":     []string{time.Now().Add(1 * time.Hour).Format(time.RFC3339)},
				"user_id":   uuid.Must(uuid.NewV7()).String(),
				"devices":   []map[string]string{{"device_id": "d", "fcm_token": "t"}},
				"task_id":   "not-a-uuid",
				"task_type": "normal",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "past time",
			requestBody: map[string]any{
				"times":     []string{time.Now().Add(-1 * time.Hour).Format(time.RFC3339)},
				"user_id":   uuid.Must(uuid.NewV7()).String(),
				"devices":   []map[string]string{{"device_id": "d", "fcm_token": "t"}},
				"task_id":   uuid.Must(uuid.NewV7()).String(),
				"task_type": "normal",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			body, _ := json.Marshal(tt.requestBody)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/reminds", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response handler.ErrorResponse

			err := json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.NotEmpty(t, response.Error)
		})
	}
}

func TestCreateRemindIdempotencySuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	router := setupTestRouter(t, testDB)

	tests := []struct {
		name string
	}{
		{
			name: "same task_id returns same reminds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			taskID := uuid.Must(uuid.NewV7()).String()

			reqBody := map[string]any{
				"times":     []string{time.Now().Add(1 * time.Hour).Format(time.RFC3339), time.Now().Add(2 * time.Hour).Format(time.RFC3339)},
				"user_id":   uuid.Must(uuid.NewV7()).String(),
				"devices":   []map[string]string{{"device_id": "d", "fcm_token": "t"}},
				"task_id":   taskID,
				"task_type": "normal",
			}
			body, _ := json.Marshal(reqBody)

			// First request
			req1 := httptest.NewRequest(http.MethodPost, "/api/v1/reminds", bytes.NewReader(body))
			req1.Header.Set("Content-Type", "application/json")

			rec1 := httptest.NewRecorder()
			router.ServeHTTP(rec1, req1)

			assert.Equal(t, http.StatusCreated, rec1.Code)

			var response1 handler.RemindsResponse

			err := json.Unmarshal(rec1.Body.Bytes(), &response1)
			require.NoError(t, err)
			require.Equal(t, 2, response1.Count)

			// Second request with same task_id (different times)
			reqBody2 := map[string]any{
				"times":     []string{time.Now().Add(3 * time.Hour).Format(time.RFC3339)},
				"user_id":   uuid.Must(uuid.NewV7()).String(),
				"devices":   []map[string]string{{"device_id": "d2", "fcm_token": "t2"}},
				"task_id":   taskID,
				"task_type": "normal",
			}
			body2, _ := json.Marshal(reqBody2)

			req2 := httptest.NewRequest(http.MethodPost, "/api/v1/reminds", bytes.NewReader(body2))
			req2.Header.Set("Content-Type", "application/json")

			rec2 := httptest.NewRecorder()
			router.ServeHTTP(rec2, req2)

			assert.Equal(t, http.StatusCreated, rec2.Code)

			var response2 handler.RemindsResponse

			err = json.Unmarshal(rec2.Body.Bytes(), &response2)
			require.NoError(t, err)

			// Should return same IDs (idempotent)
			assert.Equal(t, response1.Count, response2.Count)
			assert.Equal(t, response1.Reminds[0].ID, response2.Reminds[0].ID)
		})
	}
}

func TestGetRemindsByTimeRangeHandlerSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	router := setupTestRouter(t, testDB)

	tests := []struct {
		name          string
		remindsToSave int
	}{
		{
			name:          "get empty range",
			remindsToSave: 0,
		},
		{
			name:          "get single remind",
			remindsToSave: 1,
		},
		{
			name:          "get multiple reminds",
			remindsToSave: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			baseTime := time.Now().Add(1 * time.Hour).Truncate(time.Second)
			startTime := baseTime.Add(-30 * time.Minute)
			endTime := baseTime.Add(30 * time.Minute)

			// Create reminds
			for i := 0; i < tt.remindsToSave; i++ {
				reqBody := map[string]any{
					"times":     []string{baseTime.Add(time.Duration(i) * time.Minute).Format(time.RFC3339)},
					"user_id":   uuid.Must(uuid.NewV7()).String(),
					"devices":   []map[string]string{{"device_id": "d-" + string(rune('a'+i)), "fcm_token": "t-" + string(rune('a'+i))}},
					"task_id":   uuid.Must(uuid.NewV7()).String(),
					"task_type": "normal",
				}
				body, _ := json.Marshal(reqBody)

				req := httptest.NewRequest(http.MethodPost, "/api/v1/reminds", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)
				require.Equal(t, http.StatusCreated, rec.Code)
			}

			// Get by time range
			queryParams := url.Values{}
			queryParams.Set("start", startTime.Format(time.RFC3339))
			queryParams.Set("end", endTime.Format(time.RFC3339))
			req := httptest.NewRequest(http.MethodGet,
				"/api/v1/reminds?"+queryParams.Encode(),
				nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)

			var response handler.RemindsResponse

			err := json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.remindsToSave, response.Count)
			assert.Len(t, response.Reminds, tt.remindsToSave)
		})
	}
}

func TestGetRemindsByTimeRangeHandlerError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	router := setupTestRouter(t, testDB)

	tests := []struct {
		name           string
		setupQuery     func() string
		expectedStatus int
	}{
		{
			name: "missing start",
			setupQuery: func() string {
				params := url.Values{}
				params.Set("end", time.Now().Format(time.RFC3339))

				return params.Encode()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing end",
			setupQuery: func() string {
				params := url.Values{}
				params.Set("start", time.Now().Format(time.RFC3339))

				return params.Encode()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid start format",
			setupQuery: func() string {
				params := url.Values{}
				params.Set("start", "invalid")
				params.Set("end", time.Now().Format(time.RFC3339))

				return params.Encode()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid end format",
			setupQuery: func() string {
				params := url.Values{}
				params.Set("start", time.Now().Format(time.RFC3339))
				params.Set("end", "invalid")

				return params.Encode()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "end before start",
			setupQuery: func() string {
				params := url.Values{}
				params.Set("start", time.Now().Add(1*time.Hour).Format(time.RFC3339))
				params.Set("end", time.Now().Format(time.RFC3339))

				return params.Encode()
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/reminds?"+tt.setupQuery(), nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestUpdateThrottledHandlerSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	router := setupTestRouter(t, testDB)

	tests := []struct {
		name string
	}{
		{
			name: "update throttled to true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			// Create a remind first
			createBody := map[string]any{
				"times":     []string{time.Now().Add(1 * time.Hour).Format(time.RFC3339)},
				"user_id":   uuid.Must(uuid.NewV7()).String(),
				"devices":   []map[string]string{{"device_id": "d", "fcm_token": "t"}},
				"task_id":   uuid.Must(uuid.NewV7()).String(),
				"task_type": "normal",
			}
			body, _ := json.Marshal(createBody)

			createReq := httptest.NewRequest(http.MethodPost, "/api/v1/reminds", bytes.NewReader(body))
			createReq.Header.Set("Content-Type", "application/json")

			createRec := httptest.NewRecorder()
			router.ServeHTTP(createRec, createReq)
			require.Equal(t, http.StatusCreated, createRec.Code)

			var createResp handler.RemindsResponse

			err := json.Unmarshal(createRec.Body.Bytes(), &createResp)
			require.NoError(t, err)
			require.Equal(t, 1, createResp.Count)

			// Update throttled
			updateBody := map[string]any{
				"throttled": true,
			}
			updateBodyBytes, _ := json.Marshal(updateBody)

			updateReq := httptest.NewRequest(http.MethodPost, "/api/v1/reminds/"+createResp.Reminds[0].ID+"/throttled", bytes.NewReader(updateBodyBytes))
			updateReq.Header.Set("Content-Type", "application/json")

			updateRec := httptest.NewRecorder()

			router.ServeHTTP(updateRec, updateReq)

			assert.Equal(t, http.StatusOK, updateRec.Code)

			var updateResp handler.RemindResponse

			err = json.Unmarshal(updateRec.Body.Bytes(), &updateResp)
			assert.NoError(t, err)
			assert.True(t, updateResp.Throttled)
		})
	}
}

func TestUpdateThrottledHandlerError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	router := setupTestRouter(t, testDB)

	tests := []struct {
		name           string
		remindID       string
		requestBody    map[string]any
		expectedStatus int
	}{
		{
			name:           "non-existent remind",
			remindID:       domain.NewRemindID().String(),
			requestBody:    map[string]any{"throttled": true},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid remind ID format",
			remindID:       "invalid-uuid",
			requestBody:    map[string]any{"throttled": true},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			body, _ := json.Marshal(tt.requestBody)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/reminds/"+tt.remindID+"/throttled", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestUpdateThrottledIdempotencySuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	router := setupTestRouter(t, testDB)

	tests := []struct {
		name string
	}{
		{
			name: "multiple updates to true are idempotent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			// Create a remind first
			createBody := map[string]any{
				"times":     []string{time.Now().Add(1 * time.Hour).Format(time.RFC3339)},
				"user_id":   uuid.Must(uuid.NewV7()).String(),
				"devices":   []map[string]string{{"device_id": "d", "fcm_token": "t"}},
				"task_id":   uuid.Must(uuid.NewV7()).String(),
				"task_type": "normal",
			}
			body, _ := json.Marshal(createBody)

			createReq := httptest.NewRequest(http.MethodPost, "/api/v1/reminds", bytes.NewReader(body))
			createReq.Header.Set("Content-Type", "application/json")

			createRec := httptest.NewRecorder()
			router.ServeHTTP(createRec, createReq)
			require.Equal(t, http.StatusCreated, createRec.Code)

			var createResp handler.RemindsResponse

			err := json.Unmarshal(createRec.Body.Bytes(), &createResp)
			require.NoError(t, err)
			require.Equal(t, 1, createResp.Count)

			updateBody := map[string]any{
				"throttled": true,
			}
			updateBodyBytes, _ := json.Marshal(updateBody)

			// First update
			updateReq1 := httptest.NewRequest(http.MethodPost, "/api/v1/reminds/"+createResp.Reminds[0].ID+"/throttled", bytes.NewReader(updateBodyBytes))
			updateReq1.Header.Set("Content-Type", "application/json")

			updateRec1 := httptest.NewRecorder()
			router.ServeHTTP(updateRec1, updateReq1)
			assert.Equal(t, http.StatusOK, updateRec1.Code)

			// Second update (should be idempotent)
			updateReq2 := httptest.NewRequest(http.MethodPost, "/api/v1/reminds/"+createResp.Reminds[0].ID+"/throttled", bytes.NewReader(updateBodyBytes))
			updateReq2.Header.Set("Content-Type", "application/json")

			updateRec2 := httptest.NewRecorder()
			router.ServeHTTP(updateRec2, updateReq2)

			assert.Equal(t, http.StatusOK, updateRec2.Code)

			var updateResp2 handler.RemindResponse

			err = json.Unmarshal(updateRec2.Body.Bytes(), &updateResp2)
			assert.NoError(t, err)
			assert.True(t, updateResp2.Throttled)
		})
	}
}

func TestDeleteRemindHandlerSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	router := setupTestRouter(t, testDB)

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

			// Create a remind first
			createBody := map[string]any{
				"times":     []string{time.Now().Add(1 * time.Hour).Format(time.RFC3339)},
				"user_id":   uuid.Must(uuid.NewV7()).String(),
				"devices":   []map[string]string{{"device_id": "d", "fcm_token": "t"}},
				"task_id":   uuid.Must(uuid.NewV7()).String(),
				"task_type": "normal",
			}
			body, _ := json.Marshal(createBody)

			createReq := httptest.NewRequest(http.MethodPost, "/api/v1/reminds", bytes.NewReader(body))
			createReq.Header.Set("Content-Type", "application/json")

			createRec := httptest.NewRecorder()
			router.ServeHTTP(createRec, createReq)
			require.Equal(t, http.StatusCreated, createRec.Code)

			var createResp handler.RemindsResponse

			err := json.Unmarshal(createRec.Body.Bytes(), &createResp)
			require.NoError(t, err)
			require.Equal(t, 1, createResp.Count)

			// Delete the remind
			deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/reminds/"+createResp.Reminds[0].ID, nil)
			deleteRec := httptest.NewRecorder()

			router.ServeHTTP(deleteRec, deleteReq)

			assert.Equal(t, http.StatusNoContent, deleteRec.Code)
		})
	}
}

func TestDeleteRemindHandlerError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	router := setupTestRouter(t, testDB)

	tests := []struct {
		name           string
		remindID       string
		expectedStatus int
	}{
		{
			name:           "invalid remind ID format",
			remindID:       "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/reminds/"+tt.remindID, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestDeleteRemindIdempotentSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	router := setupTestRouter(t, testDB)

	tests := []struct {
		name string
	}{
		{
			name: "delete non-existent remind returns no content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			nonExistentID := domain.NewRemindID().String()

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/reminds/"+nonExistentID, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			// Should be idempotent - return no content even for non-existent
			assert.Equal(t, http.StatusNoContent, rec.Code)
		})
	}
}

func TestDeleteRemindDoubleDeleteSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	router := setupTestRouter(t, testDB)

	tests := []struct {
		name string
	}{
		{
			name: "double delete is idempotent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.CleanTable(t)

			// Create a remind first
			createBody := map[string]any{
				"times":     []string{time.Now().Add(1 * time.Hour).Format(time.RFC3339)},
				"user_id":   uuid.Must(uuid.NewV7()).String(),
				"devices":   []map[string]string{{"device_id": "d", "fcm_token": "t"}},
				"task_id":   uuid.Must(uuid.NewV7()).String(),
				"task_type": "normal",
			}
			body, _ := json.Marshal(createBody)

			createReq := httptest.NewRequest(http.MethodPost, "/api/v1/reminds", bytes.NewReader(body))
			createReq.Header.Set("Content-Type", "application/json")

			createRec := httptest.NewRecorder()
			router.ServeHTTP(createRec, createReq)
			require.Equal(t, http.StatusCreated, createRec.Code)

			var createResp handler.RemindsResponse

			err := json.Unmarshal(createRec.Body.Bytes(), &createResp)
			require.NoError(t, err)
			require.Equal(t, 1, createResp.Count)

			// First delete
			deleteReq1 := httptest.NewRequest(http.MethodDelete, "/api/v1/reminds/"+createResp.Reminds[0].ID, nil)
			deleteRec1 := httptest.NewRecorder()
			router.ServeHTTP(deleteRec1, deleteReq1)
			assert.Equal(t, http.StatusNoContent, deleteRec1.Code)

			// Second delete (should be idempotent)
			deleteReq2 := httptest.NewRequest(http.MethodDelete, "/api/v1/reminds/"+createResp.Reminds[0].ID, nil)
			deleteRec2 := httptest.NewRecorder()

			router.ServeHTTP(deleteRec2, deleteReq2)

			assert.Equal(t, http.StatusNoContent, deleteRec2.Code)
		})
	}
}

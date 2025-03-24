package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"medication-scheduler/internal/domain"
	myerrors "medication-scheduler/internal/errors"
	"medication-scheduler/internal/handlers"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockScheduleService struct {
	mock.Mock
}

func (m *MockScheduleService) CreateSchedule(ctx context.Context, schedule *domain.Schedule) error {
	args := m.Called(ctx, schedule)
	return args.Error(0)
}

func (m *MockScheduleService) GetSchedulesByUserID(ctx context.Context, userID int) ([]domain.Schedule, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Schedule), args.Error(1)
}

func (m *MockScheduleService) GetScheduleByIDs(ctx context.Context, userID, scheduleID int) (*domain.Schedule, error) {
	args := m.Called(ctx, userID, scheduleID)
	return args.Get(0).(*domain.Schedule), args.Error(1)
}

func (m *MockScheduleService) GetNextTakings(ctx context.Context, userID int, now time.Time) ([]domain.Schedule, error) {
	args := m.Called(ctx, userID, now)
	return args.Get(0).([]domain.Schedule), args.Error(1)
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestCreateSchedule_Success(t *testing.T) {
	mockService := new(MockScheduleService)
	logger := slog.Default()
	handler := handlers.New(mockService, logger)

	router := setupRouter()
	router.POST("/schedules", handler.CreateSchedule)

	now := time.Now().UTC()
	schedule := &domain.Schedule{
		UserID:     1,
		Medication: "Aspirin",
		Frequency:  time.Hour,
		Duration:   24 * time.Hour,
		StartTime:  now,
		EndTime:    now.Add(24 * time.Hour),
	}

	mockService.On("CreateSchedule", mock.Anything, mock.AnythingOfType("*domain.Schedule")).
		Run(func(args mock.Arguments) {
			arg := args.Get(1).(*domain.Schedule)
			arg.ID = 123
			arg.StartTime = schedule.StartTime
			arg.EndTime = schedule.EndTime
		}).
		Return(nil)

	reqBody := `{
		"user_id": 1,
		"medication": "Aspirin",
		"frequency": "1h",
		"duration": "24h"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/schedules", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, float64(123), response["id"])

	mockService.AssertExpectations(t)
}

func TestCreateSchedule_InvalidRequest(t *testing.T) {
	mockService := new(MockScheduleService)
	logger := slog.Default()
	handler := handlers.New(mockService, logger)

	router := setupRouter()
	router.POST("/schedules", handler.CreateSchedule)

	testCases := []struct {
		name     string
		body     string
		expected int
	}{
		{
			name:     "Invalid JSON",
			body:     `{"user_id": "not_a_number"}`,
			expected: http.StatusBadRequest,
		},
		{
			name:     "Invalid frequency",
			body:     `{"user_id": 1, "medication": "Aspirin", "frequency": "invalid", "duration": "24h"}`,
			expected: http.StatusBadRequest,
		},
		{
			name:     "Invalid duration",
			body:     `{"user_id": 1, "medication": "Aspirin", "frequency": "1h", "duration": "invalid"}`,
			expected: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/schedules", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expected, w.Code)
		})
	}
}

func TestGetSchedules_Success(t *testing.T) {
	mockService := new(MockScheduleService)
	logger := slog.Default()
	handler := handlers.New(mockService, logger)

	router := setupRouter()
	router.GET("/schedules", handler.GetSchedules)

	expectedSchedules := []domain.Schedule{
		{ID: 1, UserID: 1, Medication: "Aspirin"},
		{ID: 2, UserID: 1, Medication: "Ibuprofen"},
	}

	mockService.On("GetSchedulesByUserID", mock.Anything, 1).Return(expectedSchedules, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/schedules?user_id=1", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response handlers.ScheduleResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, []int{1, 2}, response.ScheduleIDs)
	mockService.AssertExpectations(t)
}

func TestGetSchedules_InvalidUserID(t *testing.T) {
	mockService := new(MockScheduleService)
	logger := slog.Default()
	handler := handlers.New(mockService, logger)

	router := setupRouter()
	router.GET("/schedules", handler.GetSchedules)

	testCases := []struct {
		name     string
		query    string
		expected int
	}{
		{
			name:     "Missing user_id",
			query:    "",
			expected: http.StatusBadRequest,
		},
		{
			name:     "Invalid user_id",
			query:    "user_id=invalid",
			expected: http.StatusBadRequest,
		},
		{
			name:     "Negative user_id",
			query:    "user_id=-1",
			expected: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/schedules?"+tc.query, nil)

			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expected, w.Code)
		})
	}
}

func TestGetExactSchedule_Success(t *testing.T) {
	mockService := new(MockScheduleService)
	logger := slog.Default()
	handler := handlers.New(mockService, logger)

	router := setupRouter()
	router.GET("/schedule", handler.GetExactSchedule)

	expectedSchedule := &domain.Schedule{
		ID:         1,
		UserID:     1,
		Medication: "Aspirin",
		Frequency:  time.Hour,
		Duration:   24 * time.Hour,
	}

	mockService.On("GetScheduleByIDs", mock.Anything, 1, 1).Return(expectedSchedule, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/schedule?user_id=1&schedule_id=1", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.Schedule
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, expectedSchedule.ID, response.ID)
	assert.Equal(t, expectedSchedule.Medication, response.Medication)
	mockService.AssertExpectations(t)
}

func TestGetExactSchedule_NotFound(t *testing.T) {
	mockService := new(MockScheduleService)
	logger := slog.Default()
	handler := handlers.New(mockService, logger)

	router := setupRouter()
	router.GET("/schedule", handler.GetExactSchedule)

	mockService.On("GetScheduleByIDs", mock.Anything, 1, 999).Return((*domain.Schedule)(nil), myerrors.ErrScheduleNotFound)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/schedule?user_id=1&schedule_id=999", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetNextTakings_Success(t *testing.T) {
	mockService := new(MockScheduleService)
	logger := slog.Default()
	handler := handlers.New(mockService, logger)

	router := setupRouter()
	router.GET("/takings", handler.GetNextTakings)

	now := time.Now().UTC()
	takings := []time.Time{now.Add(1 * time.Hour), now.Add(2 * time.Hour)}

	expectedSchedules := []domain.Schedule{
		{
			Medication: "Aspirin",
			Takings:    takings,
		},
	}

	mockService.On("GetNextTakings", mock.Anything, 1, mock.AnythingOfType("time.Time")).Return(expectedSchedules, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/takings?user_id=1", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []handlers.TakingsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Len(t, response, 1)
	assert.Equal(t, "Aspirin", response[0].Medication)
	assert.Len(t, response[0].Takings, 2)
	mockService.AssertExpectations(t)
}

func TestGetNextTakings_InvalidUserID(t *testing.T) {
	mockService := new(MockScheduleService)
	logger := slog.Default()
	handler := handlers.New(mockService, logger)

	router := setupRouter()
	router.GET("/takings", handler.GetNextTakings)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/takings?user_id=invalid", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetNextTakings_ServiceError(t *testing.T) {
	mockService := new(MockScheduleService)
	logger := slog.Default()
	handler := handlers.New(mockService, logger)

	router := setupRouter()
	router.GET("/takings", handler.GetNextTakings)

	mockService.On("GetNextTakings", mock.Anything, 1, mock.AnythingOfType("time.Time")).
		Return([]domain.Schedule{}, errors.New("service error"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/takings?user_id=1", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

package service_test

import (
	"context"
	"medication-scheduler/internal/domain"
	"medication-scheduler/internal/service"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockScheduleRepository struct {
	mock.Mock
}

func (m *MockScheduleRepository) Create(ctx context.Context, schedule *domain.Schedule) error {
	args := m.Called(ctx, schedule)
	return args.Error(0)
}

func (m *MockScheduleRepository) GetByIDs(ctx context.Context, userID, scheduleID int) (*domain.Schedule, error) {
	args := m.Called(ctx, userID, scheduleID)
	return args.Get(0).(*domain.Schedule), args.Error(1)
}

func (m *MockScheduleRepository) GetByUserID(ctx context.Context, userID int) ([]domain.Schedule, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Schedule), args.Error(1)
}

func TestCreateSchedule(t *testing.T) {
	mockRepo := new(MockScheduleRepository)
	svc := service.New(mockRepo, time.Hour)
	ctx := context.Background()

	schedule := &domain.Schedule{
		UserID:     1,
		Medication: "Aspirin",
		Frequency:  30 * time.Minute,
		Duration:   24 * time.Hour,
	}

	mockRepo.On("Create", ctx, schedule).Return(nil)
	err := svc.CreateSchedule(ctx, schedule)

	assert.NoError(t, err)
	assert.False(t, schedule.StartTime.IsZero())
	assert.False(t, schedule.EndTime.IsZero())
	mockRepo.AssertExpectations(t)
}

func TestGetScheduleByIDs(t *testing.T) {
	mockRepo := new(MockScheduleRepository)
	svc := service.New(mockRepo, time.Hour)
	ctx := context.Background()

	schedule := &domain.Schedule{
		ID:         1,
		UserID:     1,
		Medication: "Aspirin",
		Frequency:  30 * time.Minute,
		Duration:   24 * time.Hour,
		StartTime:  time.Now().Add(-time.Hour),
		EndTime:    time.Now().Add(23 * time.Hour),
	}

	mockRepo.On("GetByIDs", ctx, 1, 1).Return(schedule, nil)
	res, err := svc.GetScheduleByIDs(ctx, 1, 1)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, schedule.ID, res.ID)
	mockRepo.AssertExpectations(t)
}

func TestGetSchedulesByUserID(t *testing.T) {
	mockRepo := new(MockScheduleRepository)
	svc := service.New(mockRepo, time.Hour)
	ctx := context.Background()

	schedules := []domain.Schedule{{ID: 1, UserID: 1, Medication: "Aspirin"}}

	mockRepo.On("GetByUserID", ctx, 1).Return(schedules, nil)
	res, err := svc.GetSchedulesByUserID(ctx, 1)

	assert.NoError(t, err)
	assert.Len(t, res, 1)
	mockRepo.AssertExpectations(t)
}

func TestGetNextTakings(t *testing.T) {
	mockRepo := new(MockScheduleRepository)
	svc := service.New(mockRepo, time.Hour)
	ctx := context.Background()

	schedules := []domain.Schedule{{ID: 1, UserID: 1, Medication: "Aspirin", Frequency: 30 * time.Minute}}

	mockRepo.On("GetByUserID", ctx, 1).Return(schedules, nil)
	now := time.Date(2025, 1, 1, 9, 15, 0, 0, time.UTC)
	res, err := svc.GetNextTakings(ctx, 1, now)

	assert.NoError(t, err)
	assert.Len(t, res, 1)
	mockRepo.AssertExpectations(t)
}

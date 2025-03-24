package repository_test

import (
	"context"
	"testing"
	"time"

	"medication-scheduler/internal/domain"
	myerrors "medication-scheduler/internal/errors"
	"medication-scheduler/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockDB struct {
	mock.Mock
}

func (m *MockDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	argsMock := m.Called(ctx, sql, args)
	return argsMock.Get(0).(pgx.Row)
}

func (m *MockDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	argsMock := m.Called(ctx, sql, args)
	return argsMock.Get(0).(pgx.Rows), argsMock.Error(1)
}

func (m *MockDB) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	argsMock := m.Called(ctx, sql, args)
	return argsMock.Get(0).(pgconn.CommandTag), argsMock.Error(1)
}

func TestCreateSchedule(t *testing.T) {
	baseSchedule := &domain.Schedule{
		UserID:     1,
		Medication: "Aspirin",
		Frequency:  time.Hour,
		Duration:   24 * time.Hour,
	}

	t.Run("Success", func(t *testing.T) {
		mockDB := new(MockDB)
		repo := repository.New(mockDB)

		expectedSQL := `
        INSERT INTO schedules 
            (user_id, medication, frequency, duration, start_time, end_time)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id`

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.AnythingOfType("*int")).Run(func(args mock.Arguments) {
			*args.Get(0).(*int) = 123
		}).Return(nil)

		mockDB.On("QueryRow",
			mock.Anything,
			expectedSQL,
			mock.MatchedBy(func(args []interface{}) bool {
				return len(args) == 6 &&
					args[0] == baseSchedule.UserID &&
					args[1] == baseSchedule.Medication &&
					args[2] == baseSchedule.Frequency.Milliseconds() &&
					args[3] == baseSchedule.Duration.Milliseconds()
			}),
		).Return(mockRow)

		err := repo.Create(context.Background(), baseSchedule)
		require.NoError(t, err)
		assert.Equal(t, 123, baseSchedule.ID)
		mockDB.AssertExpectations(t)
	})
}

func TestGetByIDs(t *testing.T) {
	now := time.Now().UTC()
	validSchedule := domain.Schedule{
		ID:         1,
		UserID:     1,
		Medication: "Aspirin",
		Frequency:  time.Hour,
		Duration:   24 * time.Hour,
		StartTime:  now,
		EndTime:    now.Add(24 * time.Hour),
	}

	t.Run("Success", func(t *testing.T) {
		mockDB := new(MockDB)
		repo := repository.New(mockDB)

		mockRow := new(MockRow)
		mockRow.On("Scan",
			mock.AnythingOfType("*int"),
			mock.AnythingOfType("*int"),
			mock.AnythingOfType("*string"),
			mock.AnythingOfType("*int64"),
			mock.AnythingOfType("*int64"),
			mock.AnythingOfType("*time.Time"),
			mock.AnythingOfType("*time.Time"),
		).Run(func(args mock.Arguments) {
			*args.Get(0).(*int) = validSchedule.ID
			*args.Get(1).(*int) = validSchedule.UserID
			*args.Get(2).(*string) = validSchedule.Medication
			*args.Get(3).(*int64) = validSchedule.Frequency.Milliseconds()
			*args.Get(4).(*int64) = validSchedule.Duration.Milliseconds()
			*args.Get(5).(*time.Time) = validSchedule.StartTime
			*args.Get(6).(*time.Time) = validSchedule.EndTime
		}).Return(nil)

		mockDB.On("QueryRow",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(mockRow)

		schedule, err := repo.GetByIDs(context.Background(), 1, 1)
		require.NoError(t, err)
		assert.Equal(t, "Aspirin", schedule.Medication)
	})

	t.Run("Not found", func(t *testing.T) {
		mockDB := new(MockDB)
		repo := repository.New(mockDB)

		mockRow := new(MockRow)
		mockRow.On("Scan",
			mock.AnythingOfType("*int"),
			mock.AnythingOfType("*int"),
			mock.AnythingOfType("*string"),
			mock.AnythingOfType("*int64"),
			mock.AnythingOfType("*int64"),
			mock.AnythingOfType("*time.Time"),
			mock.AnythingOfType("*time.Time"),
		).Return(pgx.ErrNoRows)

		mockDB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).
			Return(mockRow)

		_, err := repo.GetByIDs(context.Background(), 1, 999)
		assert.ErrorIs(t, err, myerrors.ErrScheduleNotFound)
	})
}

func TestGetByUserID_Success(t *testing.T) {
	mockDB := new(MockDB)
	repo := repository.New(mockDB)

	expectedSQL := `
        SELECT id, user_id, medication, frequency, duration, start_time, end_time
        FROM schedules
        WHERE user_id = $1 AND (end_time > NOW() OR duration = 0)`

	mockRows := new(MockRows)
	mockRows.On("Next").Once().Return(true)
	mockRows.On("Next").Once().Return(false)
	mockRows.On("Close").Return(nil)

	mockRows.On("Scan",
		mock.AnythingOfType("*int"),
		mock.AnythingOfType("*int"),
		mock.AnythingOfType("*string"),
		mock.AnythingOfType("*time.Duration"),
		mock.AnythingOfType("*time.Duration"),
		mock.AnythingOfType("*time.Time"),
		mock.AnythingOfType("*time.Time")).
		Run(func(args mock.Arguments) {
			*args.Get(0).(*int) = 1
			*args.Get(1).(*int) = 1
			*args.Get(2).(*string) = "Aspirin"
			*args.Get(3).(*time.Duration) = time.Hour
			*args.Get(4).(*time.Duration) = 24 * time.Hour
			*args.Get(5).(*time.Time) = time.Now().Add(-1 * time.Hour)
			*args.Get(6).(*time.Time) = time.Now().Add(23 * time.Hour)
		}).Return(nil)

	mockDB.On("Query",
		mock.Anything,
		expectedSQL,
		[]interface{}{1}).
		Return(mockRows, nil)

	schedules, err := repo.GetByUserID(context.Background(), 1)

	require.NoError(t, err)
	require.Len(t, schedules, 1)
	assert.Equal(t, 1, schedules[0].ID)
	assert.Equal(t, "Aspirin", schedules[0].Medication)
	assert.Equal(t, time.Hour, schedules[0].Frequency)

	mockDB.AssertExpectations(t)
	mockRows.AssertExpectations(t)
}

type MockRow struct {
	mock.Mock
}

func (m *MockRow) Scan(dest ...interface{}) error {
	args := m.Called(dest...)
	return args.Error(0)
}

func (m *MockRow) Values() ([]interface{}, error) {
	args := m.Called()
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockRow) RawValues() [][]byte {
	args := m.Called()
	return args.Get(0).([][]byte)
}

type MockRows struct {
	mock.Mock
}

func (m *MockRows) Next() bool {
	return m.Called().Bool(0)
}

func (m *MockRows) Scan(dest ...interface{}) error {
	return m.Called(dest...).Error(0)
}

func (m *MockRows) Close() {
	m.Called()
}

func (m *MockRows) Err() error {
	return m.Called().Error(0)
}

func (m *MockRows) CommandTag() pgconn.CommandTag {
	args := m.Called()
	return args.Get(0).(pgconn.CommandTag)
}

func (m *MockRows) FieldDescriptions() []pgconn.FieldDescription {
	args := m.Called()
	return args.Get(0).([]pgconn.FieldDescription)
}

func (m *MockRows) Values() ([]interface{}, error) {
	args := m.Called()
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockRows) RawValues() [][]byte {
	args := m.Called()
	return args.Get(0).([][]byte)
}

func (m *MockRows) Conn() *pgx.Conn {
	args := m.Called()
	return args.Get(0).(*pgx.Conn)
}

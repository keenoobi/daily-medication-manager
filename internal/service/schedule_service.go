package service

import (
	"context"
	"fmt"
	"medication-scheduler/internal/domain"
	"time"
)

type ScheduleRepository interface {
	Create(ctx context.Context, schedule *domain.Schedule) error
	GetByIDs(ctx context.Context, userID, scheduleID int) (*domain.Schedule, error)
	GetByUserID(ctx context.Context, userID int) ([]domain.Schedule, error)
}

type ScheduleService struct {
	repo   ScheduleRepository
	period time.Duration
}

func New(repo ScheduleRepository, period time.Duration) *ScheduleService {
	return &ScheduleService{repo: repo, period: period}
}

func (s *ScheduleService) CreateSchedule(ctx context.Context, schedule *domain.Schedule) error {
	if err := schedule.Validate(); err != nil {
		return fmt.Errorf("invalid schedule: %w", err)
	}

	schedule.StartTime = time.Now().UTC()
	if schedule.Duration > 0 {
		schedule.EndTime = schedule.StartTime.Add(schedule.Duration)
	} else {
		// Для бессрочного приема
		schedule.EndTime = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
	}

	return s.repo.Create(ctx, schedule)
}

func (s *ScheduleService) GetScheduleByIDs(ctx context.Context, userID, scheduleID int) (*domain.Schedule, error) {
	schedule, err := s.repo.GetByIDs(ctx, userID, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}

	now := time.Now().UTC()
	schedule.Takings = schedule.CalculateTakings(now)
	return schedule, nil
}

func (s *ScheduleService) GetSchedulesByUserID(ctx context.Context, userID int) ([]domain.Schedule, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *ScheduleService) GetNextTakings(ctx context.Context, userID int, now time.Time) ([]domain.Schedule, error) {
	schedules, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	periodEnd := now.Add(s.period)
	fmt.Println(now, periodEnd)
	result := make([]domain.Schedule, 0, len(schedules))

	for i, schedule := range schedules {
		if taking, found := schedule.FindNextTaking(now, periodEnd); found {
			result = append(result, schedule)
			result[i].Takings = append(result[i].Takings, taking)
		}
	}

	fmt.Println(result)

	return result, nil
}

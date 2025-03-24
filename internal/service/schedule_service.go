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
	Update(ctx context.Context, schedule *domain.Schedule) error
	Delete(ctx context.Context, id int) error
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

func (s *ScheduleService) GetNextTakings(ctx context.Context, userID int) ([]domain.Schedule, error) {
	schedules, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	result := make([]domain.Schedule, 0)

	for _, schedule := range schedules {
		if !schedule.IsActive(now) {
			continue
		}

		dayStart := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, now.Location())
		dayEnd := dayStart.Add(14 * time.Hour)
		takings := schedule.CalculateTakings(now)

		for _, t := range takings {
			rounded := domain.RoundToNearest15(t)

			// Проверка временного окна
			if rounded.Before(dayStart) || rounded.After(dayEnd) {
				continue
			}

			if rounded.After(now) && rounded.Before(now.Add(s.period)) {
				schedule.Takings = []time.Time{rounded}
				result = append(result, schedule)
				break
			}
		}
	}

	return result, nil
}

package repository

import (
	"context"
	"errors"
	"fmt"
	"medication-scheduler/internal/domain"
	myerrors "medication-scheduler/internal/errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScheduleRepository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

func (r *ScheduleRepository) Create(ctx context.Context, schedule *domain.Schedule) error {
	err := r.db.QueryRow(ctx, `
        INSERT INTO schedules 
            (user_id, medication, frequency, duration, start_time, end_time)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id`,
		schedule.UserID,
		schedule.Medication,
		schedule.Frequency.Milliseconds(),
		schedule.Duration.Milliseconds(),
		schedule.StartTime,
		schedule.EndTime,
	).Scan(&schedule.ID)

	return err
}

func (r *ScheduleRepository) GetByIDs(ctx context.Context, userID, scheduleID int) (*domain.Schedule, error) {
	var (
		freqMs   int64
		durMs    int64
		schedule domain.Schedule
	)

	err := r.db.QueryRow(ctx, `
        SELECT id, user_id, medication, frequency, duration, start_time, end_time
        FROM schedules
        WHERE user_id = $1 AND id = $2`,
		userID, scheduleID,
	).Scan(
		&schedule.ID,
		&schedule.UserID,
		&schedule.Medication,
		&freqMs,
		&durMs,
		&schedule.StartTime,
		&schedule.EndTime,
	)

	schedule.Frequency = time.Duration(freqMs) * time.Millisecond
	schedule.Duration = time.Duration(durMs) * time.Millisecond

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, myerrors.ErrScheduleNotFound
		}
		return nil, fmt.Errorf("failed to fetch schedule: %w", err)
	}
	return &schedule, nil
}

func (r *ScheduleRepository) GetByUserID(ctx context.Context, userID int) ([]domain.Schedule, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id, user_id, medication, frequency, duration, start_time, end_time
        FROM schedules
        WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schedules: %w", err) // TODO: Сделать обработку ошибок получше
	}
	defer rows.Close()

	var schedules []domain.Schedule
	for rows.Next() {
		var schedule domain.Schedule
		if err := rows.Scan(&schedule.ID,
			&schedule.UserID,
			&schedule.Medication,
			&schedule.Frequency,
			&schedule.Duration,
			&schedule.StartTime,
			&schedule.EndTime,
		); err != nil {
			return nil, fmt.Errorf("failed to scan schedule: %w", err)
		}
		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

func (r *ScheduleRepository) Update(ctx context.Context, schedule *domain.Schedule) error {
	_, err := r.db.Exec(ctx, `
        UPDATE schedules
        SET user_id = $1, medication = $2, start_time = $3, end_time = $4, frequency = $5
        WHERE id = $6`,
		schedule.UserID, schedule.Medication, schedule.StartTime, schedule.EndTime, schedule.Frequency, schedule.ID,
	)
	return err
}

func (r *ScheduleRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `
        DELETE FROM schedules
        WHERE id = $1`,
		id,
	)
	return err
}

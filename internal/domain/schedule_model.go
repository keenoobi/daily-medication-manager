package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidFrequency = errors.New("frequency must be at least 15 minutes")
	ErrInvalidDuration  = errors.New("duration must be positive or zero for perpetual")
)

const (
	AvailableTime = 14
	RoundTo       = 15
)

type Schedule struct {
	ID         int
	UserID     int
	Medication string
	Frequency  time.Duration
	Duration   time.Duration
	StartTime  time.Time
	EndTime    time.Time
	Takings    []time.Time
}

func (s *Schedule) Validate() error {
	if s.Frequency < 15*time.Minute {
		return ErrInvalidFrequency
	}
	if s.Duration < 0 {
		return ErrInvalidDuration
	}
	return nil
}

func (s *Schedule) CalculateTakings(now time.Time) []time.Time {
	if !s.IsActive(now) {
		return nil
	}

	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, now.Location())
	dayEnd := dayStart.Add(AvailableTime * time.Hour)

	var takings []time.Time
	current := dayStart

	for current.Before(dayEnd) {
		rounded := RoundToNearest15(current)
		if rounded.After(dayEnd) {
			break
		}

		if len(takings) == 0 || rounded.After(takings[len(takings)-1]) {
			takings = append(takings, rounded)
		}

		current = current.Add(s.Frequency)
	}
	return takings
}

func (s *Schedule) IsActive(now time.Time) bool {
	if s.Duration == 0 {
		return true
	}
	return now.After(s.StartTime) && now.Before(s.EndTime)
}

func RoundToNearest15(t time.Time) time.Time {
	minutes := t.Minute()
	remainder := minutes % RoundTo
	if remainder == 0 {
		return t.Truncate(RoundTo * time.Minute)
	}
	return t.Add(time.Duration(RoundTo-remainder) * time.Minute).Truncate(RoundTo * time.Minute)
}

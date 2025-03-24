package domain_test

import (
	"testing"
	"time"

	"medication-scheduler/internal/domain"
)

func TestRoundToNearest15(t *testing.T) {
	tests := []struct {
		input  time.Time
		expect time.Time
	}{
		{
			time.Date(2025, 1, 1, 9, 7, 0, 0, time.UTC),
			time.Date(2025, 1, 1, 9, 15, 0, 0, time.UTC),
		},
		{
			time.Date(2025, 1, 1, 14, 52, 0, 0, time.UTC),
			time.Date(2025, 1, 1, 15, 0, 0, 0, time.UTC),
		},
		{
			time.Date(2025, 1, 1, 22, 5, 0, 0, time.UTC),
			time.Date(2025, 1, 1, 22, 15, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.input.String(), func(t *testing.T) {
			got := domain.RoundToNearest15(tt.input)
			if !got.Equal(tt.expect) {
				t.Errorf("Expected %v, got %v", tt.expect, got)
			}
		})
	}
}

func TestCalculateTakings(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		schedule domain.Schedule
		expected int
	}{
		{
			"Hourly schedule",
			domain.Schedule{
				Frequency: time.Hour,
				Duration:  24 * time.Hour,
				StartTime: now.Add(-2 * time.Hour),
				EndTime:   now.Add(24 * time.Hour),
			},
			14,
		},
		{
			"Every 30 mins",
			domain.Schedule{
				Frequency: 30 * time.Minute,
				Duration:  0,
			},
			28,
		},
		{
			"Inactive schedule",
			domain.Schedule{
				Frequency: time.Hour,
				Duration:  1 * time.Hour,
				StartTime: now.Add(-48 * time.Hour),
				EndTime:   now.Add(-24 * time.Hour),
			},
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			takings := tt.schedule.CalculateTakings(now)
			if len(takings) != tt.expected {
				t.Errorf("Expected %d takings, got %d", tt.expected, len(takings))
			}

			for _, tm := range takings {
				if tm.Hour() < 8 || tm.Hour() >= 22 {
					t.Errorf("Invalid taking time: %v", tm)
				}
			}
		})
	}
}

func TestFindNextTaking(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 15, 0, 0, time.UTC)
	schedule := domain.Schedule{
		Frequency: 2 * time.Hour,
		Duration:  0,
	}

	tests := []struct {
		name        string
		period      time.Duration
		expected    time.Time
		expectFound bool
	}{
		{
			"Within period",
			2 * time.Hour,
			time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC),
			true,
		},
		{
			"Outside period",
			30 * time.Minute,
			time.Time{},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			end := now.Add(tt.period)
			taking, found := schedule.FindNextTaking(now, end)

			if found != tt.expectFound {
				t.Errorf("Expected found=%v, got %v", tt.expectFound, found)
			}

			if found && !taking.Equal(tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, taking)
			}
		})
	}
}

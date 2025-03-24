package config_test

import (
	"os"
	"testing"
	"time"

	"medication-scheduler/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	t.Run("Default values", func(t *testing.T) {
		os.Clearenv()
		cfg := config.LoadConfig()

		assert.Equal(t, "8080", cfg.ServerPort)
		assert.Equal(t, "info", cfg.LogLevel)
		assert.Equal(t, time.Hour, cfg.NextTakingsPeriod)
	})

	t.Run("Environment variables", func(t *testing.T) {
		os.Setenv("SERVER_PORT", "3000")
		os.Setenv("LOG_LEVEL", "debug")
		os.Setenv("NEXT_TAKINGS_PERIOD", "2h")

		cfg := config.LoadConfig()

		assert.Equal(t, "3000", cfg.ServerPort)
		assert.Equal(t, "debug", cfg.LogLevel)
		assert.Equal(t, 2*time.Hour, cfg.NextTakingsPeriod)

		os.Clearenv()
	})
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{"1h", time.Hour, false},
		{"30m", 30 * time.Minute, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if tt.wantErr {
				assert.Panics(t, func() { config.ParseDuration(tt.input) })
			} else {
				assert.Equal(t, tt.expected, config.ParseDuration(tt.input))
			}
		})
	}
}

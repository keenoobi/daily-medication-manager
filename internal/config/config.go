package config

import (
	"log"
	"medication-scheduler/internal/database"
	"os"
	"time"
)

type Config struct {
	DBConfig          database.Config
	ServerPort        string
	LogLevel          string
	NextTakingsPeriod time.Duration
}

func LoadConfig() *Config {
	return &Config{
		DBConfig: database.Config{
			DBHost:     getEnv("DB_HOST", "localhost"),
			DBPort:     getEnv("DB_PORT", "5432"),
			DBUser:     getEnv("POSTGRES_USER", "postgres"),
			DBPassword: getEnv("POSTGRES_PASSWORD", "password"),
			DBName:     getEnv("POSTGRES_DB", "scheduler"),
		},
		ServerPort:        getEnv("SERVER_PORT", "8080"),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		NextTakingsPeriod: ParseDuration(getEnv("NEXT_TAKINGS_PERIOD", "1h")),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func ParseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Panicf("invalid duration format: %v", err)
	}
	return d
}

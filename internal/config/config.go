package config

import (
	"log"
	"medication-scheduler/internal/database"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBConfig          database.Config
	ServerPort        string
	LogLevel          string
	NextTakingsPeriod time.Duration
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file") // TODO: Заменить на что-то более вменяемое
	}
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
		NextTakingsPeriod: parseDuration(getEnv("NEXT_TAKINGS_PERIOD", "1h")),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Fatalf("invalid duration format: %v", err)
	}
	return d
}

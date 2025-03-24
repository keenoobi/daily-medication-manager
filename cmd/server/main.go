package main

import (
	"medication-scheduler/internal/app"
	"medication-scheduler/internal/config"
	"medication-scheduler/pkg/logger"
)

func main() {
	cfg := config.LoadConfig()

	logger := logger.NewSlogLogger(cfg.LogLevel)

	app, err := app.New(cfg, logger)
	if err != nil {
		logger.Error("Faile to create app", "error", err)
		return
	}

	if err := app.Run(); err != nil {
		logger.Error("Failed to run app", "error", err)
		return
	}

}

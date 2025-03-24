package main

import (
	"medication-scheduler/internal/app"
	"medication-scheduler/internal/config"
	"medication-scheduler/pkg/logger"
)

func main() {
	// Загрузка конфигурации
	cfg := config.LoadConfig()

	// Инициализация логгера
	logger := logger.NewSlogLogger(cfg.LogLevel)

	// // Создаем и запускаем приложение
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

// if app, err := app.New(cfg); err != nil {
// 	slog.Error("Failed to create app")
// 	return
// }

// if err := app.Run(); err != nil {
// 	slog.Error("Failed to run app: %v", err)
// 	return
// }

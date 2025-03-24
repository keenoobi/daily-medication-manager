package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"medication-scheduler/internal/config"
	"medication-scheduler/internal/database"
	"medication-scheduler/internal/handlers"
	"medication-scheduler/internal/repository"
	"medication-scheduler/internal/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	sloggin "github.com/samber/slog-gin"
)

type App struct {
	cfg     *config.Config
	logger  *slog.Logger
	router  *gin.Engine
	server  *http.Server
	dbPool  *pgxpool.Pool
	handler *handlers.ScheduleHandler
}

func New(cfg *config.Config, logger *slog.Logger) (*App, error) {
	router := gin.New()
	router.Use(sloggin.New(logger))
	router.Use(gin.Recovery())

	dbPool, err := database.NewPostgresDB(cfg.DBConfig)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	repo := repository.New(dbPool)
	service := service.New(repo, cfg.NextTakingsPeriod)

	handler := handlers.New(service, logger)

	return &App{
		cfg:     cfg,
		logger:  logger,
		router:  router,
		dbPool:  dbPool,
		handler: handler,
	}, nil
}

func (a *App) setupRouters() {
	a.router.GET("health", func(c *gin.Context) {
		a.logger.Info("This is health handler")
		c.JSON(200, gin.H{
			"message": "OK",
		})
	})

	a.router.POST("schedule", a.handler.CreateSchedule)
	a.router.GET("schedules", a.handler.GetSchedules)
	a.router.GET("schedule", a.handler.GetExactSchedule)
	a.router.GET("next_takings", a.handler.GetNextTakings)
}

func (a *App) Run() error {
	a.setupRouters()

	a.server = &http.Server{
		Addr:    ":" + a.cfg.ServerPort,
		Handler: a.router,
	}

	errChan := make(chan error)
	shutdownErrChan := make(chan error)

	go func() {
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error("Failed to start server", "error", err)
			errChan <- err
		}
	}()

	a.logger.Info("Server started on: " + a.cfg.ServerPort)

	go func() {
		shutdownErrChan <- a.waitForShutdown()
	}()

	select {
	case err := <-errChan:
		return err
	case err := <-shutdownErrChan:
		return err
	}

}

func (a *App) waitForShutdown() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	a.logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	defer func() {
		a.dbPool.Close()
		a.logger.Info("Database connection pool closed")
	}()

	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("Server forced to shutdown", "error", err)
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	a.logger.Info("Server exited gracefully")
	return nil
}

package handlers

import (
	"context"
	"log"
	"log/slog"
	"medication-scheduler/internal/domain"
	myerrors "medication-scheduler/internal/errors"

	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ScheduleService interface {
	CreateSchedule(ctx context.Context, schedule *domain.Schedule) error
	GetSchedulesByUserID(ctx context.Context, userID int) ([]domain.Schedule, error)
	GetScheduleByIDs(ctx context.Context, userID, scheduleID int) (*domain.Schedule, error)
	GetNextTakings(ctx context.Context, userID int, now time.Time) ([]domain.Schedule, error)
}

type ScheduleHandler struct {
	service ScheduleService
	logger  *slog.Logger
}

func New(service ScheduleService, logger *slog.Logger) *ScheduleHandler {
	return &ScheduleHandler{service: service, logger: logger}
}

type ScheduleRequest struct {
	UserID     int    `json:"user_id"`
	Medication string `json:"medication"`
	Frequency  string `json:"frequency"`
	Duration   string `json:"duration"`
}

type ScheduleResponse struct {
	ScheduleIDs []int  `json:"schedule_ids"`
	Count       int    `json:"count,omitempty"`
	UserID      int    `json:"user_id,omitempty"`
	Message     string `json:"message,omitempty"`
}

type TakingsResponse struct {
	Medication string      `json:"medication"`
	Takings    []time.Time `json:"takings"`
}

func (h *ScheduleHandler) CreateSchedule(c *gin.Context) {
	var req ScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Binding error: %v", err)
		myerrors.HandleError(c, myerrors.ErrInvalidRequest)
		return
	}

	freq, err := time.ParseDuration(req.Frequency)
	if err != nil {
		myerrors.HandleError(c, myerrors.ErrInvalidFrequency)
		return
	}

	dur, err := time.ParseDuration(req.Duration)
	if err != nil {
		myerrors.HandleError(c, myerrors.ErrInvalidDuration)
		return
	}

	schedule := &domain.Schedule{
		UserID:     req.UserID,
		Medication: req.Medication,
		Frequency:  freq,
		Duration:   dur,
	}

	if err := h.service.CreateSchedule(c, schedule); err != nil {
		myerrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": schedule.ID})
}

func (h *ScheduleHandler) GetSchedules(c *gin.Context) {
	userID, err := strconv.Atoi(c.Query("user_id"))
	if err != nil || userID <= 0 {
		h.logger.Error("Invalid user_id", "userID", c.Query("user_id"), "error", err)
		myerrors.HandleError(c, myerrors.ErrInvalidUserID)
		return
	}

	h.logger.Info("Fetching schedules for user", "userID", userID)

	schedules, err := h.service.GetSchedulesByUserID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to fetch schedules", "userID", userID, "error", err)
		myerrors.HandleError(c, err)
		return
	}

	if len(schedules) == 0 {
		h.logger.Info("No schedules found for user", "userID", userID)
		c.JSON(http.StatusOK, ScheduleResponse{
			ScheduleIDs: []int{},
		})
		return
	}

	scheduleIDs := make([]int, 0, len(schedules))
	for _, schedule := range schedules {
		scheduleIDs = append(scheduleIDs, schedule.ID)
	}

	response := ScheduleResponse{
		ScheduleIDs: scheduleIDs,
	}

	h.logger.Info("Successfully fetched schedules", "userID", userID, "count", len(scheduleIDs))
	c.JSON(http.StatusOK, response)
}

func (h *ScheduleHandler) GetExactSchedule(c *gin.Context) {
	userID, err := strconv.Atoi(c.Query("user_id"))
	if err != nil || userID <= 0 {
		myerrors.HandleError(c, myerrors.ErrInvalidUserID)
		return
	}

	scheduleID, err := strconv.Atoi(c.Query("schedule_id"))
	if err != nil || scheduleID <= 0 {
		myerrors.HandleError(c, myerrors.ErrInvalidScheduleID)
		return
	}

	schedule, err := h.service.GetScheduleByIDs(c.Request.Context(), userID, scheduleID)
	if err != nil {
		myerrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, schedule)
}

func (h *ScheduleHandler) GetNextTakings(c *gin.Context) {
	userID, err := strconv.Atoi(c.Query("user_id"))
	if err != nil || userID <= 0 {
		myerrors.HandleError(c, myerrors.ErrInvalidUserID)
		return
	}

	now := time.Now().UTC()
	schedules, err := h.service.GetNextTakings(c.Request.Context(), userID, now)
	if err != nil {
		myerrors.HandleError(c, err)
		return
	}

	response := make([]TakingsResponse, 0, len(schedules))
	for _, s := range schedules {
		response = append(response, TakingsResponse{
			Medication: s.Medication,
			Takings:    s.Takings,
		})
	}

	c.JSON(http.StatusOK, response)
}

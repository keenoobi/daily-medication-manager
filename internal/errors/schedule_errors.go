package myerrors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	ErrInvalidUserID     = errors.New("user ID must be positive")
	ErrInvalidScheduleID = errors.New("schedule ID must be positive")
	ErrInvalidMedication = errors.New("medication cannot be empty")
	ErrInvalidTimeRange  = errors.New("wrong start or end time")
	ErrInvalidTimeWindow = errors.New("medication can only be taken between 9:00 and 22:00")
	ErrScheduleNotFound  = errors.New("schedule not found")
	ErrForbidden         = errors.New("schedule does not belong to the user")
	ErrInvalidRequest    = errors.New("invalid data in request")
)

func Wrap(err error, context string) error {
	return fmt.Errorf("%s: %w", context, err)
}

func HandleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidUserID),
		errors.Is(err, ErrInvalidScheduleID),
		errors.Is(err, ErrInvalidMedication),
		errors.Is(err, ErrInvalidTimeRange),
		errors.Is(err, ErrInvalidRequest),
		errors.Is(err, ErrInvalidTimeWindow):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, ErrScheduleNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

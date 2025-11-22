package httputil

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RespondWithError sends an error JSON response
func RespondWithError(c *gin.Context, statusCode int, message string, err error) {
	errorCode := getErrorCode(statusCode)

	response := ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    errorCode,
			Message: message,
		},
		Timestamp: time.Now().UTC(),
	}

	// Only include detailed error info in debug mode (never in production)
	if gin.Mode() != gin.ReleaseMode && err != nil {
		response.Error.Details = err.Error()
	}

	c.JSON(statusCode, response)
}

// RespondWithValidationError sends a validation error response
func RespondWithValidationError(c *gin.Context, errors map[string]string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    ErrCodeValidationError,
			Message: "Validation failed",
			Details: errors,
		},
		Timestamp: time.Now().UTC(),
	})
}

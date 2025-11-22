package httputil

import (
	"time"

	"github.com/gin-gonic/gin"
)

// RespondWithSuccess sends a successful JSON response
/*func RespondWithSuccess(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, SuccessResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().UTC(),
	})
}*/

// RespondWithSuccessMessage sends a successful JSON response with a message
func RespondWithSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, SuccessResponse{
		Success:   true,
		Data:      data,
		Message:   message,
		Timestamp: time.Now().UTC(),
	})
}

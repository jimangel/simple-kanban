package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// ErrorHandler middleware handles errors consistently
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle any errors that occurred during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			log.Printf("Request error: %v", err)

			// Determine status code
			status := c.Writer.Status()
			if status == http.StatusOK {
				status = http.StatusInternalServerError
			}

			// Create error response
			c.JSON(status, ErrorResponse{
				Error:   http.StatusText(status),
				Message: err.Error(),
			})
		}
	}
}

// HandleError is a helper function to handle errors in handlers
func HandleError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}
package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ErrorHandler creates a middleware for handling errors and panics
func ErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// Log the panic
		if recovered != nil {
			// In a real application, you would use a proper logger here
			debug.PrintStack()
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "An internal server error occurred",
		})
	})
}

// NotFoundHandler handles 404 errors
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "The requested resource was not found",
		})
	}
}

// MethodNotAllowedHandler handles 405 errors
func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, ErrorResponse{
			Error:   "method_not_allowed",
			Message: "The requested method is not allowed for this resource",
		})
	}
}
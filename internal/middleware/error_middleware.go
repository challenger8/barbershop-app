// internal/middleware/error_middleware.go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Code    string                 `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// AppError represents a custom application error
type AppError struct {
	StatusCode int
	Code       string
	Message    string
	Details    map[string]interface{}
	Err        error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// Common error constructors
func NewBadRequestError(message string, details map[string]interface{}) *AppError {
	return &AppError{
		StatusCode: http.StatusBadRequest,
		Code:       "BAD_REQUEST",
		Message:    message,
		Details:    details,
	}
}

func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusUnauthorized,
		Code:       "UNAUTHORIZED",
		Message:    message,
	}
}

func NewForbiddenError(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusForbidden,
		Code:       "FORBIDDEN",
		Message:    message,
	}
}

func NewNotFoundError(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusNotFound,
		Code:       "NOT_FOUND",
		Message:    message,
	}
}

func NewConflictError(message string, details map[string]interface{}) *AppError {
	return &AppError{
		StatusCode: http.StatusConflict,
		Code:       "CONFLICT",
		Message:    message,
		Details:    details,
	}
}

func NewInternalServerError(message string, err error) *AppError {
	return &AppError{
		StatusCode: http.StatusInternalServerError,
		Code:       "INTERNAL_SERVER_ERROR",
		Message:    message,
		Err:        err,
	}
}

func NewValidationError(message string, details map[string]interface{}) *AppError {
	return &AppError{
		StatusCode: http.StatusUnprocessableEntity,
		Code:       "VALIDATION_ERROR",
		Message:    message,
		Details:    details,
	}
}

// ErrorHandler is a middleware that handles errors
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// Check if it's an AppError
			if appErr, ok := err.Err.(*AppError); ok {
				response := ErrorResponse{
					Error:   http.StatusText(appErr.StatusCode),
					Message: appErr.Message,
					Code:    appErr.Code,
					Details: appErr.Details,
				}

				// Log internal server errors
				if appErr.StatusCode >= 500 {
					c.Set("error", appErr.Err)
				}

				c.JSON(appErr.StatusCode, response)
				return
			}

			// Handle generic errors
			response := ErrorResponse{
				Error:   "Internal Server Error",
				Message: "An unexpected error occurred",
				Code:    "INTERNAL_SERVER_ERROR",
			}

			// In development, include error details
			if gin.Mode() == gin.DebugMode {
				response.Details = map[string]interface{}{
					"error": err.Error(),
				}
			}

			c.Set("error", err.Err)
			c.JSON(http.StatusInternalServerError, response)
			return
		}
	}
}

// RecoveryHandler handles panics and converts them to errors
func RecoveryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				c.Set("panic", err)

				response := ErrorResponse{
					Error:   "Internal Server Error",
					Message: "An unexpected error occurred",
					Code:    "INTERNAL_SERVER_ERROR",
				}

				// In development, include panic details
				if gin.Mode() == gin.DebugMode {
					response.Details = map[string]interface{}{
						"panic": err,
					}
				}

				c.AbortWithStatusJSON(http.StatusInternalServerError, response)
			}
		}()

		c.Next()
	}
}

// AbortWithError is a helper to abort with an AppError
func AbortWithError(c *gin.Context, err *AppError) {
	response := ErrorResponse{
		Error:   http.StatusText(err.StatusCode),
		Message: err.Message,
		Code:    err.Code,
		Details: err.Details,
	}

	if err.StatusCode >= 500 && err.Err != nil {
		c.Set("error", err.Err)
	}

	c.AbortWithStatusJSON(err.StatusCode, response)
}

// RespondWithError is a helper to respond with an AppError directly
func RespondWithError(c *gin.Context, err *AppError) {
	response := ErrorResponse{
		Error:   http.StatusText(err.StatusCode),
		Message: err.Message,
		Code:    err.Code,
		Details: err.Details,
	}

	if err.StatusCode >= 500 && err.Err != nil {
		c.Set("error", err.Err)
	}

	c.JSON(err.StatusCode, response)
}

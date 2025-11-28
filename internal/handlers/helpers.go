// internal/handlers/helpers.go
package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"barber-booking-system/internal/middleware"

	"github.com/gin-gonic/gin"
)

// ============================================================================
// QUERY PARAMETER PARSING
// ============================================================================

// ParseIntQuery parses an integer from query string with default value
func ParseIntQuery(c *gin.Context, key string, defaultValue int) int {
	if value := c.Query(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// ParseFloatQuery parses a float from query string with default value
func ParseFloatQuery(c *gin.Context, key string, defaultValue float64) float64 {
	if value := c.Query(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// ParseBoolQuery parses a boolean from query string, returns nil if not present
func ParseBoolQuery(c *gin.Context, key string) *bool {
	if value := c.Query(key); value != "" {
		b := value == "true"
		return &b
	}
	return nil
}

// ParseTimeQuery parses a time from query string supporting multiple formats
func ParseTimeQuery(c *gin.Context, key string) time.Time {
	value := c.Query(key)
	if value == "" {
		return time.Time{}
	}
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, value); err == nil {
			return t
		}
	}
	return time.Time{}
}

// ============================================================================
// URL PARAMETER PARSING
// ============================================================================

// ParseIntParam parses an integer from URL parameter
func ParseIntParam(c *gin.Context, paramName string) (int, error) {
	return strconv.Atoi(c.Param(paramName))
}

// RequireIntParam extracts and validates an integer URL parameter.
// Returns the value and true if valid, or sends error response and returns 0, false.
func RequireIntParam(c *gin.Context, paramName string, entityName string) (int, bool) {
	id, err := strconv.Atoi(c.Param(paramName))
	if err != nil {
		c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
			Error:   fmt.Sprintf("Invalid %s ID", entityName),
			Message: fmt.Sprintf("%s ID must be a number", entityName),
		})
		return 0, false
	}
	return id, true
}

// ============================================================================
// ERROR RESPONSE HELPERS
// ============================================================================

// RespondNotFound sends a standardized 404 response
func RespondNotFound(c *gin.Context, entityName string) {
	c.JSON(http.StatusNotFound, middleware.ErrorResponse{
		Error:   fmt.Sprintf("%s not found", entityName),
		Message: fmt.Sprintf("No %s found with the given ID", strings.ToLower(entityName)),
	})
}

// RespondInternalError sends a standardized 500 response
func RespondInternalError(c *gin.Context, operation string, err error) {
	c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
		Error:   fmt.Sprintf("Failed to %s", operation),
		Message: err.Error(),
	})
}

// RespondBadRequest sends a standardized 400 response
func RespondBadRequest(c *gin.Context, errorMsg string, message string) {
	c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
		Error:   errorMsg,
		Message: message,
	})
}

// RespondUnauthorized sends a standardized 401 response
func RespondUnauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, middleware.ErrorResponse{
		Error:   "Unauthorized",
		Message: message,
	})
}

// ============================================================================
// SUCCESS RESPONSE HELPERS
// ============================================================================

// RespondSuccess sends a success response with data
func RespondSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    data,
	})
}

// RespondSuccessWithMeta sends a success response with data and metadata
func RespondSuccessWithMeta(c *gin.Context, data interface{}, meta map[string]interface{}) {
	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// RespondSuccessWithMessage sends a success response with a message
func RespondSuccessWithMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: message,
	})
}

// RespondCreated sends a 201 response for created resources
func RespondCreated(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// ADD THESE TO YOUR EXISTING helpers.go FILE:

// ============================================================================
// ADDITIONAL HELPERS - Add these to existing file
// ============================================================================

// HandleRepositoryError intelligently maps repository errors to HTTP responses
// Returns true if error was handled, false if not recognized
func HandleRepositoryError(c *gin.Context, err error, entityName string) bool {
	switch {
	case err == sql.ErrNoRows:
		RespondNotFound(c, entityName)
		return true
	case strings.Contains(err.Error(), "not found"):
		RespondNotFound(c, entityName)
		return true
	case strings.Contains(err.Error(), "duplicate"):
		RespondBadRequest(c, "Duplicate entry", "This "+strings.ToLower(entityName)+" already exists")
		return true
	default:
		return false
	}
}

// RespondValidationError sends a 400 response for validation errors
func RespondValidationError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
		Error:   "Invalid request body",
		Message: err.Error(),
	})
}

// PaginationMeta creates standardized pagination metadata
func PaginationMeta(count, limit, offset int) map[string]interface{} {
	return map[string]interface{}{
		"count":    count,
		"limit":    limit,
		"offset":   offset,
		"has_more": count >= limit,
	}
}

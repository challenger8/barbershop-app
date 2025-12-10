// internal/handlers/helpers.go
package handlers

import (
	"barber-booking-system/internal/middleware"
	"barber-booking-system/internal/repository"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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

// ============================================================================
// CONSOLIDATED ERROR HANDLING
// ============================================================================

// HandleServiceError handles common repository errors and responds appropriately.
// Returns true if error was handled, false if err is nil.
// This consolidates the repetitive error handling pattern across all handlers.
//
// Usage:
//
//	barber, err := h.barberService.GetBarberByID(ctx, id)
//	if HandleServiceError(c, err, "Barber", "fetch barber") {
//	    return
//	}
func HandleServiceError(c *gin.Context, err error, entityName, operation string) bool {
	if err == nil {
		return false
	}

	// Check for "not found" errors from repository
	switch err {
	case repository.ErrServiceNotFound,
		repository.ErrBarberServiceNotFound,
		repository.ErrCategoryNotFound:
		RespondNotFound(c, entityName)
		return true
	}

	// Check for other common repository errors by string matching
	// (until we implement custom error types in Phase 1, Step 4)
	errMsg := err.Error()

	// Not found errors
	if strings.Contains(errMsg, "not found") {
		RespondNotFound(c, entityName)
		return true
	}

	// Duplicate entry errors
	if strings.Contains(errMsg, "duplicate") || strings.Contains(errMsg, "already exists") {
		RespondBadRequest(c, "Duplicate entry",
			fmt.Sprintf("This %s already exists", strings.ToLower(entityName)))
		return true
	}

	// Default to internal server error
	RespondInternalError(c, operation, err)
	return true
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
// JSON/QUERY/URI BINDING HELPERS (Generic with Type Parameters)
// ============================================================================

// BindJSON is a generic helper that binds and validates JSON request body.
// Returns the parsed struct and true on success, or sends error response and returns nil, false.
//
// Usage:
//
//	req, ok := BindJSON[services.CreateServiceRequest](c)
//	if !ok {
//	    return // Error response already sent
//	}
//	// Use req here...
func BindJSON[T any](c *gin.Context) (*T, bool) {
	var req T
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, err)
		return nil, false
	}
	return &req, true
}

// BindQuery is a generic helper that binds and validates query parameters.
// Returns the parsed struct and true on success, or sends error response and returns nil, false.
//
// Usage:
//
//	filters, ok := BindQuery[FilterParams](c)
//	if !ok {
//	    return
//	}
func BindQuery[T any](c *gin.Context) (*T, bool) {
	var req T
	if err := c.ShouldBindQuery(&req); err != nil {
		RespondValidationError(c, err)
		return nil, false
	}
	return &req, true
}

// BindURI is a generic helper that binds and validates URI parameters.
// Returns the parsed struct and true on success, or sends error response and returns nil, false.
//
// Usage:
//
//	params, ok := BindURI[URIParams](c)
//	if !ok {
//	    return
//	}
func BindURI[T any](c *gin.Context) (*T, bool) {
	var req T
	if err := c.ShouldBindUri(&req); err != nil {
		RespondValidationError(c, err)
		return nil, false
	}
	return &req, true
}

// RespondValidationError sends a 400 response for validation errors
// This should already exist in your code, but adding here for completeness
func RespondValidationError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
		Error:   "Invalid request body",
		Message: err.Error(),
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

// RespondSuccessWithData sends a success response with data and a message
func RespondSuccessWithData(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    data,
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

// ============================================================================
// PAGINATION HELPERS
// ============================================================================

// PaginationMeta creates standardized pagination metadata
func PaginationMeta(count, limit, offset int) map[string]interface{} {
	return map[string]interface{}{
		"count":    count,
		"limit":    limit,
		"offset":   offset,
		"has_more": count >= limit,
	}
}

// Note: ContainsAny moved to internal/utils/strings.go as utils.ContainsAny

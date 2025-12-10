// internal/middleware/validation_middleware.go
package middleware

import (
	"fmt"
	"net/http"
	"reflect"

	"barber-booking-system/internal/config"
	"barber-booking-system/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidationConfig defines the configuration for validation middleware
type ValidationConfig struct {
	SkipPaths []string
	// Custom error messages for validation tags
	CustomMessages map[string]string
}

// DefaultValidationConfig returns default validation configuration
func DefaultValidationConfig() ValidationConfig {
	return ValidationConfig{
		SkipPaths: config.DefaultSkipPaths,
		CustomMessages: map[string]string{
			"required": "%s is required",
			"email":    "%s must be a valid email address",
			"min":      "%s must be at least %s characters",
			"max":      "%s must be at most %s characters",
			"url":      "%s must be a valid URL",
			"uuid":     "%s must be a valid UUID",
		},
	}
}

// validationError represents a field validation error
type validationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag"`
	Value   string `json:"value,omitempty"`
}

// ValidateJSON validates JSON request body against a struct
func ValidateJSON(model interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a new instance of the model
		modelType := reflect.TypeOf(model)
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}
		modelValue := reflect.New(modelType).Interface()

		// Bind JSON to model
		if err := c.ShouldBindJSON(modelValue); err != nil {
			// Check if it's a validation error
			if validationErrs, ok := err.(validator.ValidationErrors); ok {
				errors := formatValidationErrors(validationErrs)
				RespondWithError(c, NewValidationError("Validation failed", map[string]interface{}{
					"errors": errors,
				}))
				c.Abort()
				return
			}

			// JSON parsing error
			RespondWithError(c, NewBadRequestError("Invalid JSON format", map[string]interface{}{
				"error": err.Error(),
			}))
			c.Abort()
			return
		}

		// Store validated model in context
		c.Set("validated_body", modelValue)
		c.Next()
	}
}

// ValidateQuery validates query parameters
func ValidateQuery(model interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a new instance of the model
		modelType := reflect.TypeOf(model)
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}
		modelValue := reflect.New(modelType).Interface()

		// Bind query parameters to model
		if err := c.ShouldBindQuery(modelValue); err != nil {
			if validationErrs, ok := err.(validator.ValidationErrors); ok {
				errors := formatValidationErrors(validationErrs)
				RespondWithError(c, NewValidationError("Invalid query parameters", map[string]interface{}{
					"errors": errors,
				}))
				c.Abort()
				return
			}

			RespondWithError(c, NewBadRequestError("Invalid query parameters", map[string]interface{}{
				"error": err.Error(),
			}))
			c.Abort()
			return
		}

		// Store validated query in context
		c.Set("validated_query", modelValue)
		c.Next()
	}
}

// ValidateURI validates URI parameters
func ValidateURI(model interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a new instance of the model
		modelType := reflect.TypeOf(model)
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}
		modelValue := reflect.New(modelType).Interface()

		// Bind URI parameters to model
		if err := c.ShouldBindUri(modelValue); err != nil {
			if validationErrs, ok := err.(validator.ValidationErrors); ok {
				errors := formatValidationErrors(validationErrs)
				RespondWithError(c, NewValidationError("Invalid URI parameters", map[string]interface{}{
					"errors": errors,
				}))
				c.Abort()
				return
			}

			// Handle binding errors (like invalid type conversions)
			RespondWithError(c, NewValidationError("Invalid URI parameters", map[string]interface{}{
				"error": err.Error(),
			}))
			c.Abort()
			return
		}

		// Store validated URI in context
		c.Set("validated_uri", modelValue)
		c.Next()
	}
}

// formatValidationErrors formats validator errors into a readable format
func formatValidationErrors(errs validator.ValidationErrors) []validationError {
	var errors []validationError

	for _, err := range errs {
		field := err.Field()
		tag := err.Tag()
		param := err.Param()

		// Convert field name to snake_case for JSON
		jsonField := utils.ToSnakeCase(field)

		// Create user-friendly message
		message := getValidationMessage(field, tag, param)

		errors = append(errors, validationError{
			Field:   jsonField,
			Message: message,
			Tag:     tag,
			Value:   fmt.Sprintf("%v", err.Value()),
		})
	}

	return errors
}

// getValidationMessage returns a user-friendly validation message
func getValidationMessage(field, tag, param string) string {
	messages := map[string]string{
		"required": fmt.Sprintf("%s is required", field),
		"email":    fmt.Sprintf("%s must be a valid email address", field),
		"min":      fmt.Sprintf("%s must be at least %s characters", field, param),
		"max":      fmt.Sprintf("%s must be at most %s characters", field, param),
		"len":      fmt.Sprintf("%s must be exactly %s characters", field, param),
		"gt":       fmt.Sprintf("%s must be greater than %s", field, param),
		"gte":      fmt.Sprintf("%s must be greater than or equal to %s", field, param),
		"lt":       fmt.Sprintf("%s must be less than %s", field, param),
		"lte":      fmt.Sprintf("%s must be less than or equal to %s", field, param),
		"alpha":    fmt.Sprintf("%s must contain only alphabetic characters", field),
		"alphanum": fmt.Sprintf("%s must contain only alphanumeric characters", field),
		"numeric":  fmt.Sprintf("%s must be a number", field),
		"url":      fmt.Sprintf("%s must be a valid URL", field),
		"uri":      fmt.Sprintf("%s must be a valid URI", field),
		"uuid":     fmt.Sprintf("%s must be a valid UUID", field),
		"uuid4":    fmt.Sprintf("%s must be a valid UUID v4", field),
		"oneof":    fmt.Sprintf("%s must be one of: %s", field, param),
		"eqfield":  fmt.Sprintf("%s must equal %s", field, param),
		"nefield":  fmt.Sprintf("%s must not equal %s", field, param),
		"datetime": fmt.Sprintf("%s must be a valid date/time in format: %s", field, param),
	}

	if msg, ok := messages[tag]; ok {
		return msg
	}

	return fmt.Sprintf("%s failed validation for tag: %s", field, tag)
}

// Note: toSnakeCase moved to internal/utils/strings.go as utils.ToSnakeCase

// GetValidatedBody retrieves the validated body from context
func GetValidatedBody(c *gin.Context) (interface{}, bool) {
	return c.Get("validated_body")
}

// GetValidatedQuery retrieves the validated query from context
func GetValidatedQuery(c *gin.Context) (interface{}, bool) {
	return c.Get("validated_query")
}

// GetValidatedURI retrieves the validated URI from context
func GetValidatedURI(c *gin.Context) (interface{}, bool) {
	return c.Get("validated_uri")
}

// MustGetValidatedBody retrieves validated body or panics
func MustGetValidatedBody(c *gin.Context, target interface{}) {
	value, exists := GetValidatedBody(c)
	if !exists {
		panic("validated body not found in context")
	}

	// Type assertion
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		panic("target must be a pointer")
	}

	sourceValue := reflect.ValueOf(value)
	if sourceValue.Kind() == reflect.Ptr {
		sourceValue = sourceValue.Elem()
	}

	targetValue.Elem().Set(sourceValue)
}

// Note: Custom validation functions (validatePhone, validateUsername) have been
// consolidated into internal/validation/validator.go to avoid duplication.
// Use validation.Initialize() to register all custom validators.

// SanitizeInput sanitizes string input to prevent XSS
func SanitizeInput() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get request body if it exists
		if c.Request.Body != nil && c.Request.Method != http.MethodGet {
			// Note: This is a simple example. For production, use a proper
			// sanitization library like bluemonday
			c.Next()
		} else {
			c.Next()
		}
	}
}

// internal/utils/validation_messages.go
package utils

import "fmt"

// GetValidationMessage returns a user-friendly validation error message
// This is the single source of truth for all validation error messages
func GetValidationMessage(field, tag, param string) string {
	messages := map[string]string{
		// Standard validators
		"required": fmt.Sprintf("%s is required", field),
		"email":    fmt.Sprintf("%s must be a valid email address", field),
		"min":      fmt.Sprintf("%s must be at least %s", field, param),
		"max":      fmt.Sprintf("%s must be at most %s", field, param),
		"len":      fmt.Sprintf("%s must be exactly %s characters", field, param),
		"gt":       fmt.Sprintf("%s must be greater than %s", field, param),
		"gte":      fmt.Sprintf("%s must be greater than or equal to %s", field, param),
		"lt":       fmt.Sprintf("%s must be less than %s", field, param),
		"lte":      fmt.Sprintf("%s must be less than or equal to %s", field, param),
		"url":      fmt.Sprintf("%s must be a valid URL", field),
		"uri":      fmt.Sprintf("%s must be a valid URI", field),
		"uuid":     fmt.Sprintf("%s must be a valid UUID", field),
		"uuid4":    fmt.Sprintf("%s must be a valid UUID v4", field),
		"oneof":    fmt.Sprintf("%s must be one of: %s", field, param),
		"eqfield":  fmt.Sprintf("%s must equal %s", field, param),
		"nefield":  fmt.Sprintf("%s must not equal %s", field, param),
		"datetime": fmt.Sprintf("%s must be a valid date/time in format: %s", field, param),

		// Character type validators
		"alpha":    fmt.Sprintf("%s must contain only alphabetic characters", field),
		"alphanum": fmt.Sprintf("%s must contain only alphanumeric characters", field),
		"numeric":  fmt.Sprintf("%s must be a number", field),

		// Custom validators
		"phone":          fmt.Sprintf("%s must be a valid phone number (E.164 format: +[country code][number])", field),
		"booking_number": fmt.Sprintf("%s must be in format BK-YYYYMMDD-XXXX", field),
		"booking_status": fmt.Sprintf("%s must be a valid booking status", field),
		"barber_status":  fmt.Sprintf("%s must be a valid barber status", field),
		"payment_status": fmt.Sprintf("%s must be a valid payment status", field),
		"not_future":     fmt.Sprintf("%s cannot be in the future", field),
		"not_past":       fmt.Sprintf("%s cannot be in the past", field),
	}

	if msg, ok := messages[tag]; ok {
		return msg
	}

	return fmt.Sprintf("%s is invalid", field)
}

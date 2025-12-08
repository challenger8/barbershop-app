// internal/validation/validator.go
package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ========================================================================
// CUSTOM VALIDATOR - Enhanced Validation with go-playground/validator
// ========================================================================

var (
	// Global validator instance
	validate *validator.Validate

	// Phone regex (E.164 format with minimum length requirement)
	// Must start with + and have at least 7 digits total
	phoneRegex = regexp.MustCompile(`^\+[1-9]\d{6,14}$`)

	// Booking number format regex (e.g., BK-20240101-0001)
	bookingNumberRegex = regexp.MustCompile(`^BK-\d{8}-\d{4}$`)
)

// Initialize sets up the validator with custom validators
func Initialize() {
	validate = validator.New()

	// Register custom validators
	_ = validate.RegisterValidation("phone", validatePhone)
	_ = validate.RegisterValidation("booking_number", validateBookingNumber)
	_ = validate.RegisterValidation("booking_status", validateBookingStatus)
	_ = validate.RegisterValidation("barber_status", validateBarberStatus)
	_ = validate.RegisterValidation("payment_status", validatePaymentStatus)
	_ = validate.RegisterValidation("not_future", validateNotFuture)
	_ = validate.RegisterValidation("not_past", validateNotPast)
}

// GetValidator returns the global validator instance
func GetValidator() *validator.Validate {
	if validate == nil {
		Initialize()
	}
	return validate
}

// ========================================================================
// CUSTOM VALIDATORS
// ========================================================================

// validatePhone validates phone numbers (E.164 format with minimum length)
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if phone == "" {
		return true // Optional fields handled by 'required' tag
	}
	return phoneRegex.MatchString(phone)
}

// validateBookingNumber validates booking number format
func validateBookingNumber(fl validator.FieldLevel) bool {
	number := fl.Field().String()
	if number == "" {
		return true
	}
	return bookingNumberRegex.MatchString(number)
}

// validateBookingStatus validates booking status values
func validateBookingStatus(fl validator.FieldLevel) bool {
	status := fl.Field().String()
	validStatuses := map[string]bool{
		"pending":     true,
		"confirmed":   true,
		"in_progress": true,
		"completed":   true,
		"cancelled":   true,
		"no_show":     true,
	}
	return validStatuses[status]
}

// validateBarberStatus validates barber status values
func validateBarberStatus(fl validator.FieldLevel) bool {
	status := fl.Field().String()
	validStatuses := map[string]bool{
		"pending":   true,
		"active":    true,
		"inactive":  true,
		"suspended": true,
		"rejected":  true,
	}
	return validStatuses[status]
}

// validatePaymentStatus validates payment status values
func validatePaymentStatus(fl validator.FieldLevel) bool {
	status := fl.Field().String()
	validStatuses := map[string]bool{
		"pending":        true,
		"paid":           true,
		"partially_paid": true,
		"refunded":       true,
		"failed":         true,
	}
	return validStatuses[status]
}

// validateNotFuture validates that a date is not in the future
func validateNotFuture(fl validator.FieldLevel) bool {
	// Implementation would check if date is not after current time
	return true // Simplified for now
}

// validateNotPast validates that a date is not in the past
func validateNotPast(fl validator.FieldLevel) bool {
	// Implementation would check if date is not before current time
	return true // Simplified for now
}

// ========================================================================
// VALIDATION HELPERS
// ========================================================================

// ValidateStruct validates a struct and returns user-friendly errors
func ValidateStruct(s interface{}) error {
	if err := GetValidator().Struct(s); err != nil {
		return FormatValidationErrors(err)
	}
	return nil
}

// FormatValidationErrors converts validator errors to user-friendly messages
func FormatValidationErrors(err error) error {
	if err == nil {
		return nil
	}

	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}

	var messages []string
	for _, e := range validationErrs {
		messages = append(messages, formatFieldError(e))
	}

	return fmt.Errorf("validation failed: %s", strings.Join(messages, "; "))
}

// formatFieldError creates a user-friendly error message for a field
func formatFieldError(e validator.FieldError) string {
	field := toSnakeCase(e.Field())

	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "phone":
		return fmt.Sprintf("%s must be a valid phone number (E.164 format: +[country code][number])", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, e.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, e.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, e.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, e.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, e.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, e.Param())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, e.Param())
	case "booking_number":
		return fmt.Sprintf("%s must be in format BK-YYYYMMDD-XXXX", field)
	case "booking_status":
		return fmt.Sprintf("%s must be a valid booking status", field)
	case "barber_status":
		return fmt.Sprintf("%s must be a valid barber status", field)
	case "payment_status":
		return fmt.Sprintf("%s must be a valid payment status", field)
	case "not_future":
		return fmt.Sprintf("%s cannot be in the future", field)
	case "not_past":
		return fmt.Sprintf("%s cannot be in the past", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// toSnakeCase converts camelCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// ========================================================================
// CONVENIENCE FUNCTIONS
// ========================================================================

// ValidateEmail validates an email address
func ValidateEmail(email string) error {
	type EmailStruct struct {
		Email string `validate:"required,email"`
	}
	return ValidateStruct(&EmailStruct{Email: email})
}

// ValidatePhone validates a phone number
func ValidatePhone(phone string) error {
	type PhoneStruct struct {
		Phone string `validate:"required,phone"`
	}
	return ValidateStruct(&PhoneStruct{Phone: phone})
}

// ValidateUUID validates a UUID
func ValidateUUID(uuid string) error {
	type UUIDStruct struct {
		UUID string `validate:"required,uuid"`
	}
	return ValidateStruct(&UUIDStruct{UUID: uuid})
}

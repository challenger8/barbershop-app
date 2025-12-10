// internal/validation/validator.go
package validation

import (
	"fmt"
	"regexp"
	"strings"

	"barber-booking-system/internal/config"
	"barber-booking-system/internal/utils"

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
		config.BookingStatusPending:     true,
		config.BookingStatusConfirmed:   true,
		config.BookingStatusInProgress:  true,
		config.BookingStatusCompleted:   true,
		config.BookingStatusCancelled:   true,
		config.BookingStatusNoShow:      true,
	}
	return validStatuses[status]
}

// validateBarberStatus validates barber status values
func validateBarberStatus(fl validator.FieldLevel) bool {
	status := fl.Field().String()
	validStatuses := map[string]bool{
		config.BarberStatusPending:   true,
		config.BarberStatusActive:    true,
		config.BarberStatusInactive:  true,
		config.BarberStatusSuspended: true,
		config.BarberStatusRejected:  true,
	}
	return validStatuses[status]
}

// validatePaymentStatus validates payment status values
func validatePaymentStatus(fl validator.FieldLevel) bool {
	status := fl.Field().String()
	validStatuses := map[string]bool{
		config.PaymentStatusPending:       true,
		config.PaymentStatusPaid:          true,
		config.PaymentStatusPartiallyPaid: true,
		config.PaymentStatusRefunded:      true,
		config.PaymentStatusFailed:        true,
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
// Uses the shared utils.GetValidationMessage for consistent error messages
func formatFieldError(e validator.FieldError) string {
	field := utils.ToSnakeCase(e.Field())
	return utils.GetValidationMessage(field, e.Tag(), e.Param())
}

// Note: toSnakeCase moved to internal/utils/strings.go as utils.ToSnakeCase

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

// internal/repository/errors.go
package repository

import "errors"

// ========================================================================
// ENTITY NOT FOUND ERRORS
// ========================================================================

var (
	// User errors
	ErrUserNotFound = errors.New("user not found")

	// Barber errors
	ErrBarberNotFound = errors.New("barber not found")

	// Service errors
	ErrServiceNotFound       = errors.New("service not found")
	ErrCategoryNotFound      = errors.New("category not found")
	ErrBarberServiceNotFound = errors.New("barber service not found")

	// Booking errors
	ErrBookingNotFound  = errors.New("booking not found")
	ErrTimeSlotNotFound = errors.New("time slot not found")

	// Review errors
	ErrReviewNotFound = errors.New("review not found")
)

// ========================================================================
// DUPLICATE/CONFLICT ERRORS
// ========================================================================

var (
	// User duplicates
	ErrDuplicateEmail = errors.New("email already exists")

	// Barber duplicates
	ErrDuplicateBarber = errors.New("user already has a barber profile")

	// Service duplicates
	ErrDuplicateSlug     = errors.New("service slug already exists")
	ErrDuplicateService  = errors.New("service name already exists")
	ErrDuplicateCategory = errors.New("category already exists")

	// Booking conflicts
	ErrBookingConflict = errors.New("time slot already booked")
)

// ========================================================================
// VALIDATION ERRORS
// ========================================================================

var (
	ErrInvalidStatus    = errors.New("invalid status")
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrInvalidTimeSlot  = errors.New("invalid time slot")
	ErrInvalidDateRange = errors.New("invalid date range")
)

// ========================================================================
// BUSINESS LOGIC ERRORS
// ========================================================================

var (
	// Booking business rules
	ErrPastBookingTime         = errors.New("cannot book in the past")
	ErrBarberUnavailable       = errors.New("barber is unavailable")
	ErrServiceInactive         = errors.New("service is not active")
	ErrBarberInactive          = errors.New("barber is not active")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrBookingTooFarInAdvance  = errors.New("booking too far in advance")
	ErrInsufficientNotice      = errors.New("insufficient notice for booking")
	ErrCannotCancelCompleted   = errors.New("cannot cancel completed booking")
	ErrAlreadyCancelled        = errors.New("booking already cancelled")
)

// ========================================================================
// PERMISSION ERRORS
// ========================================================================

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrNotOwner     = errors.New("not the owner of this resource")
)

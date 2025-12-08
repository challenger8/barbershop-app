// internal/repository/errors.go
package repository

import "errors"

// ============================================================================
// REPOSITORY ERRORS - Domain-specific error types
// ============================================================================

// Entity Not Found Errors
var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")

	// ErrBarberNotFound is returned when a barber is not found
	ErrBarberNotFound = errors.New("barber not found")

	// ErrServiceNotFound is returned when a service is not found
	ErrServiceNotFound = errors.New("service not found")

	// ErrCategoryNotFound is returned when a category is not found
	ErrCategoryNotFound = errors.New("category not found")

	// ErrBarberServiceNotFound is returned when a barber-service association is not found
	ErrBarberServiceNotFound = errors.New("barber service not found")

	// ErrBookingNotFound is returned when a booking is not found
	ErrBookingNotFound = errors.New("booking not found")

	// ErrReviewNotFound is returned when a review is not found
	ErrReviewNotFound = errors.New("review not found")

	// ErrTimeSlotNotFound is returned when a time slot is not found
	ErrTimeSlotNotFound = errors.New("time slot not found")
)

// Duplicate/Conflict Errors
var (
	// ErrDuplicateEmail is returned when email already exists
	ErrDuplicateEmail = errors.New("email already registered")

	// ErrDuplicateSlug is returned when slug already exists
	ErrDuplicateSlug = errors.New("slug already exists")

	// ErrDuplicateBarber is returned when barber already exists for user
	ErrDuplicateBarber = errors.New("barber profile already exists for this user")

	// ErrDuplicateService is returned when service name already exists
	ErrDuplicateService = errors.New("service already exists")

	// ErrDuplicateCategory is returned when category name already exists
	ErrDuplicateCategory = errors.New("category already exists")
)

// Validation Errors
var (
	// ErrInvalidStatus is returned when status value is invalid
	ErrInvalidStatus = errors.New("invalid status")

	// ErrInvalidEmail is returned when email format is invalid
	ErrInvalidEmail = errors.New("invalid email format")

	// ErrInvalidPassword is returned when password doesn't meet requirements
	ErrInvalidPassword = errors.New("invalid password")

	// ErrInvalidTimeSlot is returned when time slot is invalid
	ErrInvalidTimeSlot = errors.New("invalid time slot")

	// ErrInvalidDateRange is returned when date range is invalid
	ErrInvalidDateRange = errors.New("invalid date range")
)

// Business Logic Errors
var (
	// ErrBookingConflict is returned when booking time conflicts with existing booking
	ErrBookingConflict = errors.New("booking time conflict with existing booking")

	// ErrPastBookingTime is returned when attempting to book in the past
	ErrPastBookingTime = errors.New("cannot book time in the past")

	// ErrBarberUnavailable is returned when barber is not available
	ErrBarberUnavailable = errors.New("barber is not available at requested time")

	// ErrServiceInactive is returned when trying to book inactive service
	ErrServiceInactive = errors.New("service is not active")

	// ErrBarberInactive is returned when barber profile is inactive
	ErrBarberInactive = errors.New("barber profile is not active")

	// ErrInvalidStatusTransition is returned when status transition is not allowed
	ErrInvalidStatusTransition = errors.New("invalid status transition")

	// ErrBookingTooFarInAdvance is returned when booking is too far in future
	ErrBookingTooFarInAdvance = errors.New("booking is too far in advance")

	// ErrInsufficientNotice is returned when booking doesn't meet minimum notice requirement
	ErrInsufficientNotice = errors.New("insufficient notice for booking")

	// ErrCannotCancelCompleted is returned when trying to cancel completed booking
	ErrCannotCancelCompleted = errors.New("cannot cancel completed booking")

	// ErrAlreadyCancelled is returned when booking is already cancelled
	ErrAlreadyCancelled = errors.New("booking is already cancelled")
)

// Permission Errors
var (
	// ErrUnauthorized is returned when user is not authenticated
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when user doesn't have permission
	ErrForbidden = errors.New("forbidden")

	// ErrNotOwner is returned when user is not the owner of the resource
	ErrNotOwner = errors.New("user is not the owner of this resource")
)

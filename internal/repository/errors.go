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

	// Notification errors
	ErrNotificationNotFound = errors.New("notification not found")
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

	// Review duplicates
	ErrDuplicateReview     = errors.New("review already exists for this booking")
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

	// Review validation
	ErrInvalidRating       = errors.New("rating must be between 1 and 5")
	ErrInvalidModeration   = errors.New("invalid moderation status")
	ErrBookingNotCompleted = errors.New("can only review completed bookings")
	ErrCannotModifyReview  = errors.New("review cannot be modified")

	// Notification validation
	ErrInvalidNotificationType   = errors.New("invalid notification type")
	ErrInvalidNotificationStatus = errors.New("invalid notification status")
	ErrNotificationExpired       = errors.New("notification has expired")
	ErrNotificationAlreadySent   = errors.New("notification has already been sent")
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

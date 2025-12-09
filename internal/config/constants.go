// internal/config/constants.go
package config

import "time"

// ========================================================================
// PAGINATION CONSTANTS
// ========================================================================

const (
	// DefaultPageLimit is the default number of items per page
	DefaultPageLimit = 20

	// MaxPageLimit is the maximum number of items per page
	MaxPageLimit = 100

	// MinPageLimit is the minimum number of items per page
	MinPageLimit = 1

	// BarberServicesPageLimit is the default limit for barber services
	BarberServicesPageLimit = 50
)

// ========================================================================
// CACHE TTL CONSTANTS
// ========================================================================

const (
	// CacheTTLShort is for frequently changing data (5 minutes)
	CacheTTLShort = 5 * time.Minute

	// CacheTTLMedium is for moderately stable data (15 minutes)
	CacheTTLMedium = 15 * time.Minute

	// CacheTTLLong is for stable data (1 hour)
	CacheTTLLong = 1 * time.Hour

	// CacheTTLBarber is TTL for barber data
	CacheTTLBarber = CacheTTLMedium

	// CacheTTLService is TTL for service data
	CacheTTLService = CacheTTLLong

	// CacheTTLCategory is TTL for category data
	CacheTTLCategory = CacheTTLLong
)

// ========================================================================
// BOOKING CONSTANTS
// ========================================================================

const (
	// MinBookingDurationMinutes is the minimum booking duration in minutes
	MinBookingDurationMinutes = 15

	// MaxBookingDurationMinutes is the maximum booking duration in minutes
	MaxBookingDurationMinutes = 480 // 8 hours

	// DefaultBookingDurationMinutes is the default booking duration
	DefaultBookingDurationMinutes = 60 // 1 hour

	// MinAdvanceBookingHours is the minimum notice required for booking
	MinAdvanceBookingHours = 2

	// MaxAdvanceBookingDays is the maximum days in advance to book
	MaxAdvanceBookingDays = 90

	// DefaultAdvanceBookingDays is the default advance booking window
	DefaultAdvanceBookingDays = 30

	// BookingBufferMinutes is the buffer time between bookings
	BookingBufferMinutes = 15
)

// ========================================================================
// BUSINESS HOURS CONSTANTS
// ========================================================================

const (
	// BusinessHoursStart is the default business opening time (9 AM)
	BusinessHoursStart = 9

	// BusinessHoursEnd is the default business closing time (6 PM)
	BusinessHoursEnd = 18

	// TimeSlotIntervalMinutes is the default time slot interval
	TimeSlotIntervalMinutes = 30

	// MaxDailyBookings is the maximum bookings per barber per day
	MaxDailyBookings = 16
)

// ========================================================================
// RATING & REVIEW CONSTANTS
// ========================================================================

const (
	// MinRating is the minimum rating value (1-5 scale)
	MinRating = 1.0

	// MaxRating is the maximum rating value (1-5 scale)
	MaxRating = 5.0

	// DefaultRating is the default rating for new barbers
	DefaultRating = 0.0

	// MinReviewLength is the minimum review text length
	MinReviewLength = 10

	// MaxReviewLength is the maximum review text length
	MaxReviewLength = 1000
)

// ========================================================================
// SEARCH & FILTER CONSTANTS
// ========================================================================

const (
	// MinSearchQueryLength is the minimum search query length
	MinSearchQueryLength = 2

	// MaxSearchQueryLength is the maximum search query length
	MaxSearchQueryLength = 100

	// SearchRelevanceThreshold is the minimum relevance score for search results
	SearchRelevanceThreshold = 0.3

	// MaxSearchResults is the maximum number of search results to return
	MaxSearchResults = 100
)

// ========================================================================
// AUTHENTICATION CONSTANTS
// ========================================================================

const (
	// MaxFailedLoginAttempts is the maximum failed login attempts before lockout
	MaxFailedLoginAttempts = 5

	// AccountLockDuration is how long an account stays locked
	AccountLockDuration = 30 * time.Minute

	// MinPasswordLength is the minimum password length
	MinPasswordLength = 8

	// MaxPasswordLength is the maximum password length
	MaxPasswordLength = 128

	// TokenExpirationTime is the default JWT token expiration
	TokenExpirationTime = 24 * time.Hour

	// RefreshTokenExpirationTime is the refresh token expiration
	RefreshTokenExpirationTime = 7 * 24 * time.Hour
)

// ========================================================================
// FILE UPLOAD CONSTANTS
// ========================================================================

const (
	// MaxImageSizeBytes is the maximum image upload size (10 MB)
	MaxImageSizeBytes = 10 * 1024 * 1024

	// MaxDocumentSizeBytes is the maximum document upload size (50 MB)
	MaxDocumentSizeBytes = 50 * 1024 * 1024

	// MaxGalleryImages is the maximum number of gallery images
	MaxGalleryImages = 20

	// ImageQuality is the JPEG quality for image compression (0-100)
	ImageQuality = 85

	// ThumbnailWidth is the width of thumbnail images in pixels
	ThumbnailWidth = 300

	// ThumbnailHeight is the height of thumbnail images in pixels
	ThumbnailHeight = 300
)

// ========================================================================
// PRICING CONSTANTS
// ========================================================================

const (
	// DefaultCurrency is the default currency code
	DefaultCurrency = "USD"

	// MinServicePrice is the minimum service price
	MinServicePrice = 0.0

	// MaxServicePrice is the maximum service price
	MaxServicePrice = 10000.0

	// DefaultTaxRate is the default tax rate (percentage)
	DefaultTaxRate = 0.0

	// DefaultCommissionRate is the default commission rate for barbers
	DefaultCommissionRate = 15.0 // 15%
)

// ========================================================================
// NOTIFICATION CONSTANTS
// ========================================================================

const (
	// BookingReminderHoursBefore is hours before booking to send reminder
	BookingReminderHoursBefore = 24

	// ReviewRequestHoursAfter is hours after completed booking to request review
	ReviewRequestHoursAfter = 24

	// MaxNotificationRetries is maximum retry attempts for failed notifications
	MaxNotificationRetries = 3

	// NotificationRetryDelay is delay between notification retry attempts
	NotificationRetryDelay = 5 * time.Minute
)

// ========================================================================
// PERFORMANCE CONSTANTS
// ========================================================================

const (
	// SlowQueryThresholdMs is the threshold for slow query logging (milliseconds)
	SlowQueryThresholdMs = 1000

	// MaxConcurrentRequests is the maximum concurrent requests per user
	MaxConcurrentRequests = 10

	// DatabasePoolSize is the connection pool size
	DatabasePoolSize = 25

	// DatabaseIdleConnections is the number of idle connections
	DatabaseIdleConnections = 5

	// DatabaseConnectionLifetime is the connection lifetime
	DatabaseConnectionLifetime = 5 * time.Minute
)

// ========================================================================
// DEFAULT STATUS VALUES
// ========================================================================

const (
	// User statuses
	UserStatusActive    = "active"
	UserStatusInactive  = "inactive"
	UserStatusSuspended = "suspended"
	UserStatusDeleted   = "deleted"

	// Barber statuses
	BarberStatusPending   = "pending"
	BarberStatusActive    = "active"
	BarberStatusInactive  = "inactive"
	BarberStatusSuspended = "suspended"
	BarberStatusRejected  = "rejected"

	// Booking statuses
	BookingStatusPending    = "pending"
	BookingStatusConfirmed  = "confirmed"
	BookingStatusInProgress = "in_progress"
	BookingStatusCompleted  = "completed"
	BookingStatusCancelled  = "cancelled"
	BookingStatusNoShow     = "no_show"

	// Payment statuses
	PaymentStatusPending   = "pending"
	PaymentStatusPaid      = "paid"
	PaymentStatusFailed    = "failed"
	PaymentStatusRefunded  = "refunded"
	PaymentStatusCancelled = "cancelled"
)

// ========================================================================
// USER TYPES
// ========================================================================

const (
	UserTypeCustomer = "customer"
	UserTypeBarber   = "barber"
	UserTypeAdmin    = "admin"
)

// ========================================================================
// REVIEW STATUS VALUES
// ========================================================================

const (
	// Review moderation statuses
	ReviewModerationPending  = "pending"
	ReviewModerationApproved = "approved"
	ReviewModerationRejected = "rejected"
	ReviewModerationFlagged  = "flagged"
)

// ========================================================================
// NOTIFICATION STATUS VALUES
// ========================================================================

const (
	// Notification delivery statuses
	NotificationStatusPending   = "pending"
	NotificationStatusSent      = "sent"
	NotificationStatusDelivered = "delivered"
	NotificationStatusRead      = "read"
	NotificationStatusFailed    = "failed"

	// Notification priority levels
	NotificationPriorityLow    = "low"
	NotificationPriorityNormal = "normal"
	NotificationPriorityHigh   = "high"
	NotificationPriorityUrgent = "urgent"

	// Notification types
	NotificationTypeBookingConfirmation = "booking_confirmation"
	NotificationTypeBookingReminder     = "booking_reminder"
	NotificationTypeBookingCancelled    = "booking_cancelled"
	NotificationTypeBookingRescheduled  = "booking_rescheduled"
	NotificationTypeBookingCompleted    = "booking_completed"
	NotificationTypeReviewRequest       = "review_request"
	NotificationTypeReviewResponse      = "review_response"
	NotificationTypePaymentReceived     = "payment_received"
	NotificationTypePaymentFailed       = "payment_failed"
	NotificationTypeAccountWelcome      = "account_welcome"
	NotificationTypeAccountVerification = "account_verification"
	NotificationTypePasswordReset       = "password_reset"
	NotificationTypePromotion           = "promotion"
	NotificationTypeSystemAlert         = "system_alert"

	// Notification channels
	NotificationChannelApp   = "app"
	NotificationChannelEmail = "email"
	NotificationChannelSMS   = "sms"
	NotificationChannelPush  = "push"
)

// ========================================================================
// RELATED ENTITY TYPES
// ========================================================================

const (
	EntityTypeBooking = "booking"
	EntityTypePayment = "payment"
	EntityTypeReview  = "review"
)

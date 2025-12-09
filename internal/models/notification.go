package models

import "time"

// Notification represents system notifications
type Notification struct {
	ID      int    `json:"id" db:"id"`
	UserID  int    `json:"user_id" db:"user_id"`
	Title   string `json:"title" db:"title"`
	Message string `json:"message" db:"message"`
	Type    string `json:"type" db:"type"` // booking_confirmation, booking_reminder, booking_cancelled, etc.

	// Delivery channels and status
	Channels StringArray `json:"channels" db:"channels"` // app, email, sms, push
	Status   string      `json:"status" db:"status"`     // pending, sent, delivered, read, failed

	// Delivery tracking
	SentAt      *time.Time `json:"sent_at" db:"sent_at"`
	DeliveredAt *time.Time `json:"delivered_at" db:"delivered_at"`
	ReadAt      *time.Time `json:"read_at" db:"read_at"`

	// Related entities
	RelatedEntityType *string `json:"related_entity_type" db:"related_entity_type"` // booking, payment, review
	RelatedEntityID   *int    `json:"related_entity_id" db:"related_entity_id"`

	// Notification data and settings
	Data     JSONMap `json:"data" db:"data"`         // Additional notification data
	Priority string  `json:"priority" db:"priority"` // low, normal, high, urgent

	// Scheduling
	ScheduledFor *time.Time `json:"scheduled_for" db:"scheduled_for"`
	ExpiresAt    *time.Time `json:"expires_at" db:"expires_at"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// Relations
	User *User `json:"user,omitempty"`
}

// Note: Helper methods for User, Booking, and Review models have been moved to their
// respective files (user.go, booking.go, review.go) for better code organization.

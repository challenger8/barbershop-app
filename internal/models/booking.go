package models

import (
	"barber-booking-system/internal/config"
	"errors"
	"time"
)

// Booking represents a booking/appointment
type Booking struct {
	ID            int    `json:"id" db:"id"`
	UUID          string `json:"uuid" db:"uuid"`
	BookingNumber string `json:"booking_number" db:"booking_number"` // Human-readable reference

	// Relationships
	CustomerID *int `json:"customer_id" db:"customer_id"` // Nullable for guest bookings
	BarberID   int  `json:"barber_id" db:"barber_id"`
	TimeSlotID int  `json:"time_slot_id" db:"time_slot_id"`

	// Service information
	ServiceName              string  `json:"service_name" db:"service_name"`
	ServiceCategory          *string `json:"service_category" db:"service_category"`
	EstimatedDurationMinutes int     `json:"estimated_duration_minutes" db:"estimated_duration_minutes"`

	// Customer information (for guest bookings)
	CustomerName  *string `json:"customer_name" db:"customer_name"`
	CustomerEmail *string `json:"customer_email" db:"customer_email"`
	CustomerPhone *string `json:"customer_phone" db:"customer_phone"`

	// Booking status
	Status string `json:"status" db:"status"` // pending, confirmed, in_progress, completed, cancelled_by_customer, cancelled_by_barber, no_show

	// Pricing breakdown
	ServicePrice   float64 `json:"service_price" db:"service_price"`
	TotalPrice     float64 `json:"total_price" db:"total_price"`
	DiscountAmount float64 `json:"discount_amount" db:"discount_amount"`
	TaxAmount      float64 `json:"tax_amount" db:"tax_amount"`
	TipAmount      float64 `json:"tip_amount" db:"tip_amount"`
	Currency       string  `json:"currency" db:"currency"`

	// Payment information
	PaymentStatus    string     `json:"payment_status" db:"payment_status"` // pending, paid, partially_paid, refunded, failed
	PaymentMethod    *string    `json:"payment_method" db:"payment_method"`
	PaymentReference *string    `json:"payment_reference" db:"payment_reference"`
	PaidAt           *time.Time `json:"paid_at" db:"paid_at"`

	// Booking details
	Notes           *string `json:"notes" db:"notes"`
	SpecialRequests *string `json:"special_requests" db:"special_requests"`
	InternalNotes   *string `json:"internal_notes" db:"internal_notes"` // For barber use only

	// Communication tracking
	ConfirmationMethod *string    `json:"confirmation_method" db:"confirmation_method"` // email, sms, phone, app
	ConfirmationSentAt *time.Time `json:"confirmation_sent_at" db:"confirmation_sent_at"`
	ReminderSentAt     *time.Time `json:"reminder_sent_at" db:"reminder_sent_at"`

	// Timing
	ScheduledStartTime time.Time  `json:"scheduled_start_time" db:"scheduled_start_time"`
	ScheduledEndTime   time.Time  `json:"scheduled_end_time" db:"scheduled_end_time"`
	ActualStartTime    *time.Time `json:"actual_start_time" db:"actual_start_time"`
	ActualEndTime      *time.Time `json:"actual_end_time" db:"actual_end_time"`

	// Cancellation information
	CancelledAt        *time.Time `json:"cancelled_at" db:"cancelled_at"`
	CancelledBy        *int       `json:"cancelled_by" db:"cancelled_by"`
	CancellationReason *string    `json:"cancellation_reason" db:"cancellation_reason"`
	CancellationFee    float64    `json:"cancellation_fee" db:"cancellation_fee"`

	// Source and attribution
	BookingSource  string  `json:"booking_source" db:"booking_source"` // mobile_app, web_app, phone, walk_in, admin
	ReferralSource *string `json:"referral_source" db:"referral_source"`
	UTMCampaign    *string `json:"utm_campaign" db:"utm_campaign"`

	// Audit fields
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Relations (populated when needed)
	Customer *User     `json:"customer,omitempty"`
	Barber   *Barber   `json:"barber,omitempty"`
	TimeSlot *TimeSlot `json:"time_slot,omitempty"`
	Review   *Review   `json:"review,omitempty"`
}

// BookingHistory represents audit trail for booking changes
type BookingHistory struct {
	ID           int       `json:"id" db:"id"`
	BookingID    int       `json:"booking_id" db:"booking_id"`
	ChangedBy    *int      `json:"changed_by" db:"changed_by"`
	ChangeType   string    `json:"change_type" db:"change_type"` // created, status_changed, rescheduled, cancelled
	OldValues    JSONMap   `json:"old_values" db:"old_values"`
	NewValues    JSONMap   `json:"new_values" db:"new_values"`
	ChangeReason *string   `json:"change_reason" db:"change_reason"`
	IPAddress    *string   `json:"ip_address" db:"ip_address"`
	UserAgent    *string   `json:"user_agent" db:"user_agent"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// ========================================================================
// BOOKING HELPER METHODS
// ========================================================================

// IsPending returns true if the booking is pending
func (b *Booking) IsPending() bool {
	return b.Status == config.BookingStatusPending
}

// IsConfirmed returns true if the booking is confirmed
func (b *Booking) IsConfirmed() bool {
	return b.Status == config.BookingStatusConfirmed
}

// IsCompleted returns true if the booking is completed
func (b *Booking) IsCompleted() bool {
	return b.Status == config.BookingStatusCompleted
}

// IsCancelled returns true if the booking was cancelled
func (b *Booking) IsCancelled() bool {
	return b.Status == config.BookingStatusCancelledByCustomer || b.Status == config.BookingStatusCancelledByBarber
}

// CanBeCancelled returns true if the booking can still be cancelled
func (b *Booking) CanBeCancelled() bool {
	return b.Status == config.BookingStatusPending || b.Status == config.BookingStatusConfirmed
}

// GetCustomerInfo returns customer name, email, and phone
func (b *Booking) GetCustomerInfo() (string, string, string) {
	name := ""
	email := ""
	phone := ""

	if b.CustomerName != nil {
		name = *b.CustomerName
	}
	if b.CustomerEmail != nil {
		email = *b.CustomerEmail
	}
	if b.CustomerPhone != nil {
		phone = *b.CustomerPhone
	}

	return name, email, phone
}

// GetDuration returns the scheduled duration of the booking
func (b *Booking) GetDuration() time.Duration {
	return b.ScheduledEndTime.Sub(b.ScheduledStartTime)
}

// IsUpcoming returns true if the booking is scheduled for the future
func (b *Booking) IsUpcoming() bool {
	return b.ScheduledStartTime.After(time.Now()) && (b.Status == config.BookingStatusPending || b.Status == config.BookingStatusConfirmed)
}

// Validate validates booking fields
func (b *Booking) Validate() error {
	if b.BarberID <= 0 {
		return errors.New("valid barber ID is required")
	}
	if b.TimeSlotID <= 0 {
		return errors.New("valid time slot ID is required")
	}
	if b.ServiceName == "" {
		return errors.New("service name is required")
	}
	if b.TotalPrice < 0 {
		return errors.New("total price cannot be negative")
	}
	return nil
}

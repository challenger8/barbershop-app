package models

import (
	"errors"
	"time"
)

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

// Helper methods for User model
func (u *User) IsCustomer() bool {
	return u.UserType == "customer"
}

func (u *User) IsBarber() bool {
	return u.UserType == "barber"
}

func (u *User) IsAdmin() bool {
	return u.UserType == "admin"
}

func (u *User) IsActive() bool {
	return u.Status == "active"
}

func (u *User) GetFullName() string {
	return u.Name
}

func (u *User) GetDisplayLocation() string {
	if u.City != nil && u.State != nil {
		return *u.City + ", " + *u.State
	} else if u.City != nil {
		return *u.City
	}
	return ""
}

// Helper methods for Booking model
func (b *Booking) IsPending() bool {
	return b.Status == "pending"
}

func (b *Booking) IsConfirmed() bool {
	return b.Status == "confirmed"
}

func (b *Booking) IsCompleted() bool {
	return b.Status == "completed"
}

func (b *Booking) IsCancelled() bool {
	return b.Status == "cancelled_by_customer" || b.Status == "cancelled_by_barber"
}

func (b *Booking) CanBeCancelled() bool {
	return b.Status == "pending" || b.Status == "confirmed"
}

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

func (b *Booking) GetDuration() time.Duration {
	return b.ScheduledEndTime.Sub(b.ScheduledStartTime)
}

func (b *Booking) IsUpcoming() bool {
	return b.ScheduledStartTime.After(time.Now()) && (b.Status == "pending" || b.Status == "confirmed")
}

// Helper methods for Review model
func (r *Review) GetAverageRating() float64 {
	ratings := []int{r.OverallRating}
	count := 1

	if r.ServiceQualityRating != nil {
		ratings = append(ratings, *r.ServiceQualityRating)
		count++
	}
	if r.PunctualityRating != nil {
		ratings = append(ratings, *r.PunctualityRating)
		count++
	}
	if r.CleanlinessRating != nil {
		ratings = append(ratings, *r.CleanlinessRating)
		count++
	}
	if r.ValueForMoneyRating != nil {
		ratings = append(ratings, *r.ValueForMoneyRating)
		count++
	}
	if r.ProfessionalismRating != nil {
		ratings = append(ratings, *r.ProfessionalismRating)
		count++
	}

	sum := 0
	for _, rating := range ratings {
		sum += rating
	}

	return float64(sum) / float64(count)
}

func (r *Review) IsPositive() bool {
	return r.OverallRating >= 4
}

func (r *Review) GetHelpfulnessRatio() float64 {
	if r.TotalVotes == 0 {
		return 0
	}
	return float64(r.HelpfulVotes) / float64(r.TotalVotes)
}

// Validation methods
func (u *User) Validate() error {
	if u.Name == "" {
		return errors.New("name is required")
	}
	if u.Email == "" {
		return errors.New("email is required")
	}
	if u.UserType == "" {
		return errors.New("user type is required")
	}
	return nil
}

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

func (r *Review) Validate() error {
	if r.BookingID <= 0 {
		return errors.New("valid booking ID is required")
	}
	if r.BarberID <= 0 {
		return errors.New("valid barber ID is required")
	}
	if r.OverallRating < 1 || r.OverallRating > 5 {
		return errors.New("overall rating must be between 1 and 5")
	}
	return nil
}

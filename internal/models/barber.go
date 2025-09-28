package models

import "time"

// Barber represents a barber's business profile
type Barber struct {
	ID        int     `json:"id" db:"id"`
	UserID    int     `json:"user_id" db:"user_id"` // Links to User table
	UUID      string  `json:"uuid" db:"uuid"`
	UserName  *string `json:"user_name" db:"user_name"`
	UserEmail *string `json:"user_email" db:"user_email"`
	// Business information
	ShopName                   string  `json:"shop_name" db:"shop_name"`
	BusinessName               *string `json:"business_name" db:"business_name"`
	BusinessRegistrationNumber *string `json:"business_registration_number" db:"business_registration_number"`
	TaxID                      *string `json:"tax_id" db:"tax_id"`

	// Location and contact
	Address      string   `json:"address" db:"address"`
	AddressLine2 *string  `json:"address_line_2" db:"address_line_2"`
	City         string   `json:"city" db:"city"`
	State        string   `json:"state" db:"state"`
	Country      string   `json:"country" db:"country"`
	PostalCode   string   `json:"postal_code" db:"postal_code"`
	Latitude     *float64 `json:"latitude" db:"latitude"`
	Longitude    *float64 `json:"longitude" db:"longitude"`

	Phone         *string `json:"phone" db:"phone"`
	BusinessEmail *string `json:"business_email" db:"business_email"`
	WebsiteURL    *string `json:"website_url" db:"website_url"`

	// Business details
	Description     *string     `json:"description" db:"description"`
	YearsExperience *int        `json:"years_experience" db:"years_experience"`
	Specialties     StringArray `json:"specialties" db:"specialties"`
	Certifications  StringArray `json:"certifications" db:"certifications"`
	LanguagesSpoken StringArray `json:"languages_spoken" db:"languages_spoken"`

	// Media and branding
	ProfileImageURL *string     `json:"profile_image_url" db:"profile_image_url"`
	CoverImageURL   *string     `json:"cover_image_url" db:"cover_image_url"`
	GalleryImages   StringArray `json:"gallery_images" db:"gallery_images"`

	// Working hours (flexible JSON structure)
	WorkingHours JSONMap `json:"working_hours" db:"working_hours"`

	// Business metrics
	Rating        float64 `json:"rating" db:"rating"`
	TotalReviews  int     `json:"total_reviews" db:"total_reviews"`
	TotalBookings int     `json:"total_bookings" db:"total_bookings"`

	// Performance metrics
	ResponseTimeMinutes int     `json:"response_time_minutes" db:"response_time_minutes"`
	AcceptanceRate      float64 `json:"acceptance_rate" db:"acceptance_rate"`
	CancellationRate    float64 `json:"cancellation_rate" db:"cancellation_rate"`

	// Status and verification
	Status            string     `json:"status" db:"status"` // pending, active, inactive, suspended, rejected
	IsVerified        bool       `json:"is_verified" db:"is_verified"`
	VerificationDate  *time.Time `json:"verification_date" db:"verification_date"`
	VerificationNotes *string    `json:"verification_notes" db:"verification_notes"`

	// Business settings
	AdvanceBookingDays    int  `json:"advance_booking_days" db:"advance_booking_days"`
	MinBookingNoticeHours int  `json:"min_booking_notice_hours" db:"min_booking_notice_hours"`
	AutoAcceptBookings    bool `json:"auto_accept_bookings" db:"auto_accept_bookings"`
	InstantBookingEnabled bool `json:"instant_booking_enabled" db:"instant_booking_enabled"`

	// Financial information
	CommissionRate float64 `json:"commission_rate" db:"commission_rate"`
	PayoutMethod   string  `json:"payout_method" db:"payout_method"`
	PayoutDetails  JSONMap `json:"payout_details" db:"payout_details"`

	// Audit fields
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	LastActiveAt time.Time  `json:"last_active_at" db:"last_active_at"`
	DeletedAt    *time.Time `json:"deleted_at" db:"deleted_at"`

	// Relations (populated when needed)
	User     *User           `json:"user,omitempty"`
	Services []BarberService `json:"services,omitempty"`
}

// BarberAvailability represents detailed availability patterns
type BarberAvailability struct {
	ID        int       `json:"id" db:"id"`
	BarberID  int       `json:"barber_id" db:"barber_id"`
	Date      time.Time `json:"date" db:"date"`
	DayOfWeek int       `json:"day_of_week" db:"day_of_week"` // 0=Sunday, 6=Saturday
	StartTime time.Time `json:"start_time" db:"start_time"`
	EndTime   time.Time `json:"end_time" db:"end_time"`

	// Availability type
	AvailabilityType string `json:"availability_type" db:"availability_type"` // available, busy, break, blocked

	// Recurring settings
	IsRecurring      bool       `json:"is_recurring" db:"is_recurring"`
	RecurringPattern *string    `json:"recurring_pattern" db:"recurring_pattern"` // weekly, monthly, none
	RecurringEndDate *time.Time `json:"recurring_end_date" db:"recurring_end_date"`

	// Notes and reasons
	Notes         *string `json:"notes" db:"notes"`
	BlockedReason *string `json:"blocked_reason" db:"blocked_reason"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

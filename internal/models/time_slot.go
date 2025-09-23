package models

import "time"

// TimeSlot represents available appointment slots
type TimeSlot struct {
	ID              int       `json:"id" db:"id"`
	BarberID        int       `json:"barber_id" db:"barber_id"`
	StartTime       time.Time `json:"start_time" db:"start_time"`
	EndTime         time.Time `json:"end_time" db:"end_time"`
	DurationMinutes int       `json:"duration_minutes" db:"duration_minutes"`

	// Availability and type
	IsAvailable bool   `json:"is_available" db:"is_available"`
	SlotType    string `json:"slot_type" db:"slot_type"` // regular, premium, last_minute, group

	// Pricing
	BasePrice          float64  `json:"base_price" db:"base_price"`
	DynamicPrice       *float64 `json:"dynamic_price" db:"dynamic_price"` // AI-optimized pricing
	DiscountPercentage float64  `json:"discount_percentage" db:"discount_percentage"`

	// Service constraints
	ServiceID             *int `json:"service_id" db:"service_id"` // Specific service for this slot
	MaxCustomers          int  `json:"max_customers" db:"max_customers"`
	MinAdvanceNoticeHours int  `json:"min_advance_notice_hours" db:"min_advance_notice_hours"`

	// Metadata
	Notes               *string `json:"notes" db:"notes"`
	SpecialRequirements JSONMap `json:"special_requirements" db:"special_requirements"`

	// Audit fields
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy *int      `json:"created_by" db:"created_by"`

	// Relations
	Barber  *Barber        `json:"barber,omitempty"`
	Service *BarberService `json:"service,omitempty"`
}

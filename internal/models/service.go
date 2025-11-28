// internal/models/service.go
package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// StringArray handles PostgreSQL text arrays
type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, s)
}

// JSONMap handles JSONB fields in PostgreSQL
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, j)
}

// Service represents a service in the global catalog (master catalog)
type Service struct {
	ID   int    `json:"id" db:"id"`
	UUID string `json:"uuid" db:"uuid"`
	Name string `json:"name" db:"name"`
	Slug string `json:"slug" db:"slug"` // URL-friendly name

	// Service details
	ShortDescription    string  `json:"short_description" db:"short_description"`
	DetailedDescription *string `json:"detailed_description" db:"detailed_description"`
	CategoryID          int     `json:"category_id" db:"category_id"`
	// Service characteristics
	ServiceType        string `json:"service_type" db:"service_type"`                 // haircut, styling, treatment, grooming
	Complexity         int    `json:"complexity" db:"complexity"`                     // 1-5 scale (1=easy, 5=expert)
	SkillLevelRequired string `json:"skill_level_required" db:"skill_level_required"` // beginner, intermediate, advanced, expert

	// Default duration (barbers can override)
	DefaultDurationMin int  `json:"default_duration_min" db:"default_duration_min"` // 30 minutes
	DefaultDurationMax *int `json:"default_duration_max" db:"default_duration_max"` // 45 minutes

	// Suggested pricing (barbers set their own prices)
	SuggestedPriceMin *float64 `json:"suggested_price_min" db:"suggested_price_min"` // $20.00
	SuggestedPriceMax *float64 `json:"suggested_price_max" db:"suggested_price_max"` // $40.00
	Currency          string   `json:"currency" db:"currency"`                       // USD, EUR, etc.

	// Target demographics
	TargetGender string      `json:"target_gender" db:"target_gender"` // male, female, all, non_binary
	TargetAgeMin *int        `json:"target_age_min" db:"target_age_min"`
	TargetAgeMax *int        `json:"target_age_max" db:"target_age_max"`
	HairTypes    StringArray `json:"hair_types" db:"hair_types"` // straight, wavy, curly, coily, all

	// Service requirements
	RequiresConsultation   bool        `json:"requires_consultation" db:"requires_consultation"`
	RequiredTools          StringArray `json:"required_tools" db:"required_tools"`       // ["scissors", "comb", "clippers"]
	RequiredProducts       StringArray `json:"required_products" db:"required_products"` // ["shampoo", "conditioner"]
	RequiredCertifications StringArray `json:"required_certifications" db:"required_certifications"`

	// Health and safety
	AllergenWarnings    StringArray `json:"allergen_warnings" db:"allergen_warnings"`   // ["nuts", "fragrance"]
	HealthPrecautions   StringArray `json:"health_precautions" db:"health_precautions"` // ["pregnancy", "skin_sensitivity"]
	RequiresHealthCheck bool        `json:"requires_health_check" db:"requires_health_check"`

	// Media
	ImageURL      *string     `json:"image_url" db:"image_url"`
	GalleryImages StringArray `json:"gallery_images" db:"gallery_images"`
	VideoURL      *string     `json:"video_url" db:"video_url"`

	// SEO and search
	Tags            StringArray `json:"tags" db:"tags"`                       // ["trendy", "classic", "quick"]
	SearchKeywords  StringArray `json:"search_keywords" db:"search_keywords"` // ["fade", "buzz cut", "trim"]
	MetaDescription *string     `json:"meta_description" db:"meta_description"`

	// Service variations and add-ons
	HasVariations bool `json:"has_variations" db:"has_variations"` // Can have length variations
	AllowsAddOns  bool `json:"allows_add_ons" db:"allows_add_ons"` // Can add beard trim, etc.

	// Global statistics (updated periodically)
	GlobalPopularityScore float64 `json:"global_popularity_score" db:"global_popularity_score"` // 0-100
	TotalGlobalBookings   int     `json:"total_global_bookings" db:"total_global_bookings"`
	AverageGlobalRating   float64 `json:"average_global_rating" db:"average_global_rating"` // 0-5
	TotalGlobalReviews    int     `json:"total_global_reviews" db:"total_global_reviews"`

	// Status and approval
	IsActive      bool    `json:"is_active" db:"is_active"`
	IsApproved    bool    `json:"is_approved" db:"is_approved"` // Platform approval required
	ApprovalNotes *string `json:"approval_notes" db:"approval_notes"`

	// Audit fields
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy      *int      `json:"created_by" db:"created_by"` // Admin user ID
	LastModifiedBy *int      `json:"last_modified_by" db:"last_modified_by"`

	// Version control
	Version   int     `json:"version" db:"version"`
	ChangeLog JSONMap `json:"change_log" db:"change_log"`

	// Relations (populated when needed via joins)
	CategoryName *string          `json:"category_name,omitempty" db:"category_name"` // Populated from JOIN
	Category     *ServiceCategory `json:"category,omitempty"`
}

// BarberService represents the junction table - how a barber offers a specific service
type BarberService struct {
	ID        int `json:"id" db:"id"`
	BarberID  int `json:"barber_id" db:"barber_id"`   // FK to barbers table
	ServiceID int `json:"service_id" db:"service_id"` // FK to services table

	// Barber's customization of the service
	CustomName        *string `json:"custom_name" db:"custom_name"`               // Override service name
	CustomDescription *string `json:"custom_description" db:"custom_description"` // Barber's custom description

	// Barber's pricing
	Price              float64    `json:"price" db:"price"`                   // Required: barber's price
	MaxPrice           *float64   `json:"max_price" db:"max_price"`           // For variable/surge pricing
	Currency           string     `json:"currency" db:"currency"`             // USD, EUR, etc.
	DiscountPrice      *float64   `json:"discount_price" db:"discount_price"` // Special offer price
	DiscountValidUntil *time.Time `json:"discount_valid_until" db:"discount_valid_until"`

	// Barber's timing
	EstimatedDurationMin int  `json:"estimated_duration_min" db:"estimated_duration_min"` // 30 minutes
	EstimatedDurationMax *int `json:"estimated_duration_max" db:"estimated_duration_max"` // 45 minutes
	BufferTimeMinutes    int  `json:"buffer_time_minutes" db:"buffer_time_minutes"`       // Break between services

	// Barber's booking rules
	AdvanceNoticeHours    int         `json:"advance_notice_hours" db:"advance_notice_hours"`         // 2 hours notice
	MaxAdvanceBookingDays *int        `json:"max_advance_booking_days" db:"max_advance_booking_days"` // 30 days max
	AvailableDays         StringArray `json:"available_days" db:"available_days"`                     // ["monday", "tuesday"]
	AvailableTimeSlots    JSONMap     `json:"available_time_slots" db:"available_time_slots"`         // Custom time restrictions

	// Barber's service-specific requirements
	RequiresConsultation   *bool   `json:"requires_consultation" db:"requires_consultation"`       // Override global setting
	ConsultationDuration   *int    `json:"consultation_duration" db:"consultation_duration"`       // 15 minutes
	PreServiceInstructions *string `json:"pre_service_instructions" db:"pre_service_instructions"` // "Wash hair first"
	PostServiceCare        *string `json:"post_service_care" db:"post_service_care"`               // "Avoid washing for 24h"

	// Age restrictions
	MinCustomerAge *int `json:"min_customer_age" db:"min_customer_age"` // 16 years old
	MaxCustomerAge *int `json:"max_customer_age" db:"max_customer_age"` // No upper limit usually

	// Seasonal availability
	IsSeasonal         bool `json:"is_seasonal" db:"is_seasonal"`
	SeasonalStartMonth *int `json:"seasonal_start_month" db:"seasonal_start_month"` // 3 (March)
	SeasonalEndMonth   *int `json:"seasonal_end_month" db:"seasonal_end_month"`     // 10 (October)

	// Barber's portfolio for this service
	PortfolioImages   StringArray `json:"portfolio_images" db:"portfolio_images"`       // Barber's work examples
	BeforeAfterImages StringArray `json:"before_after_images" db:"before_after_images"` // Before/after photos

	// Performance metrics for this barber's service
	TotalBookings        int     `json:"total_bookings" db:"total_bookings"`               // 156 bookings
	TotalRevenue         float64 `json:"total_revenue" db:"total_revenue"`                 // $3,900.00
	AverageRating        float64 `json:"average_rating" db:"average_rating"`               // 4.7 stars
	TotalReviews         int     `json:"total_reviews" db:"total_reviews"`                 // 89 reviews
	CancellationRate     float64 `json:"cancellation_rate" db:"cancellation_rate"`         // 5.2%
	CustomerSatisfaction float64 `json:"customer_satisfaction" db:"customer_satisfaction"` // 94.5%
	RepeatCustomerRate   float64 `json:"repeat_customer_rate" db:"repeat_customer_rate"`   // 67.3%

	// Recent performance (for ML and trending)
	BookingsLast30Days int     `json:"bookings_last_30_days" db:"bookings_last_30_days"` // 12 bookings
	RevenueLast30Days  float64 `json:"revenue_last_30_days" db:"revenue_last_30_days"`   // $300.00
	PopularityScore    float64 `json:"popularity_score" db:"popularity_score"`           // ML-calculated score
	DemandLevel        float64 `json:"demand_level" db:"demand_level"`                   // Current demand (0-1)

	// Marketing and promotions
	IsPromotional      bool       `json:"is_promotional" db:"is_promotional"`     // Special offer
	PromotionalText    *string    `json:"promotional_text" db:"promotional_text"` // "20% off this week!"
	PromotionStartDate *time.Time `json:"promotion_start_date" db:"promotion_start_date"`
	PromotionEndDate   *time.Time `json:"promotion_end_date" db:"promotion_end_date"`
	IsFeatured         bool       `json:"is_featured" db:"is_featured"` // Highlight this service

	// Display and ordering
	DisplayOrder int     `json:"display_order" db:"display_order"` // Order in barber's list
	ServiceNote  *string `json:"service_note" db:"service_note"`   // Special note for customers

	// Status
	IsActive     bool       `json:"is_active" db:"is_active"`         // Barber currently offers this
	PausedReason *string    `json:"paused_reason" db:"paused_reason"` // Why service is paused
	PausedUntil  *time.Time `json:"paused_until" db:"paused_until"`   // Temporary pause

	// Audit fields
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Relations (populated via joins when needed)
	Service *Service `json:"service,omitempty"` // The service from catalog
	Barber  *Barber  `json:"barber,omitempty"`  // The barber offering it
}

// ServiceCategory represents service categories
type ServiceCategory struct {
	ID          int     `json:"id" db:"id"`
	Name        string  `json:"name" db:"name"`               // "Haircuts"
	Slug        string  `json:"slug" db:"slug"`               // "haircuts"
	Description *string `json:"description" db:"description"` // "Professional hair cutting services"

	// Hierarchy support
	ParentCategoryID *int   `json:"parent_category_id" db:"parent_category_id"` // For subcategories
	Level            int    `json:"level" db:"level"`                           // 1=main, 2=sub, etc.
	CategoryPath     string `json:"category_path" db:"category_path"`           // "haircuts/mens"

	// Display attributes
	IconURL   *string `json:"icon_url" db:"icon_url"`     // Category icon
	ColorHex  *string `json:"color_hex" db:"color_hex"`   // #FF6B35
	ImageURL  *string `json:"image_url" db:"image_url"`   // Category banner
	SortOrder int     `json:"sort_order" db:"sort_order"` // Display order

	// Status
	IsActive   bool `json:"is_active" db:"is_active"`
	IsFeatured bool `json:"is_featured" db:"is_featured"` // Show prominently

	// SEO
	MetaTitle       *string     `json:"meta_title" db:"meta_title"`
	MetaDescription *string     `json:"meta_description" db:"meta_description"`
	Keywords        StringArray `json:"keywords" db:"keywords"`

	// Statistics (updated periodically)
	ServiceCount    int     `json:"service_count" db:"service_count"`       // Services in category
	BarberCount     int     `json:"barber_count" db:"barber_count"`         // Barbers offering services
	AveragePrice    float64 `json:"average_price" db:"average_price"`       // Average price in category
	PopularityScore float64 `json:"popularity_score" db:"popularity_score"` // Category popularity

	// Audit fields
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Relations
	ParentCategory *ServiceCategory  `json:"parent_category,omitempty"`
	SubCategories  []ServiceCategory `json:"sub_categories,omitempty"`
	Services       []Service         `json:"services,omitempty"`
}

// Helper methods for Service model
func (s *Service) GetComplexityLabel() string {
	switch s.Complexity {
	case 1:
		return "Very Easy"
	case 2:
		return "Easy"
	case 3:
		return "Moderate"
	case 4:
		return "Difficult"
	case 5:
		return "Expert Level"
	default:
		return "Unknown"
	}
}

func (s *Service) IsCompatibleWithHairType(hairType string) bool {
	for _, compatibleType := range s.HairTypes {
		if compatibleType == hairType || compatibleType == "all" {
			return true
		}
	}
	return false
}

func (s *Service) GetEstimatedDuration() int {
	if s.DefaultDurationMax != nil {
		return (s.DefaultDurationMin + *s.DefaultDurationMax) / 2
	}
	return s.DefaultDurationMin
}

func (s *Service) GetSuggestedPriceRange() string {
	if s.SuggestedPriceMin != nil && s.SuggestedPriceMax != nil {
		return fmt.Sprintf("$%.2f - $%.2f", *s.SuggestedPriceMin, *s.SuggestedPriceMax)
	} else if s.SuggestedPriceMin != nil {
		return fmt.Sprintf("From $%.2f", *s.SuggestedPriceMin)
	}
	return "Price varies"
}

// Helper methods for BarberService model
func (bs *BarberService) GetFinalPrice(demandMultiplier float64, isLoyalCustomer bool) float64 {
	price := bs.Price

	// Apply current discount if valid
	if bs.DiscountPrice != nil && bs.DiscountValidUntil != nil && time.Now().Before(*bs.DiscountValidUntil) {
		price = *bs.DiscountPrice
	}

	// Apply demand-based pricing (ML integration)
	if demandMultiplier > 1.0 && bs.MaxPrice != nil {
		adjustedPrice := price * demandMultiplier
		if adjustedPrice > *bs.MaxPrice {
			price = *bs.MaxPrice
		} else {
			price = adjustedPrice
		}
	}

	// Apply loyalty discount (5% for loyal customers)
	if isLoyalCustomer {
		price = price * 0.95
	}

	return price
}

func (bs *BarberService) GetEstimatedDuration() int {
	if bs.EstimatedDurationMax != nil {
		return (bs.EstimatedDurationMin + *bs.EstimatedDurationMax) / 2
	}
	return bs.EstimatedDurationMin
}

func (bs *BarberService) GetTotalServiceTime() int {
	return bs.GetEstimatedDuration() + bs.BufferTimeMinutes
}

func (bs *BarberService) IsAvailableOn(day string) bool {
	for _, availableDay := range bs.AvailableDays {
		if availableDay == day {
			return true
		}
	}
	return false
}

func (bs *BarberService) RequiresAdvanceBooking(requestedTime time.Time) bool {
	requiredNotice := time.Duration(bs.AdvanceNoticeHours) * time.Hour
	return time.Now().Add(requiredNotice).After(requestedTime)
}

func (bs *BarberService) IsWithinSeasonalPeriod() bool {
	if !bs.IsSeasonal || bs.SeasonalStartMonth == nil || bs.SeasonalEndMonth == nil {
		return true // Always available if not seasonal
	}

	currentMonth := int(time.Now().Month())
	start := *bs.SeasonalStartMonth
	end := *bs.SeasonalEndMonth

	if start <= end {
		return currentMonth >= start && currentMonth <= end
	} else {
		// Crosses year boundary (e.g., Nov-Feb)
		return currentMonth >= start || currentMonth <= end
	}
}

func (bs *BarberService) GetPerformanceScore() float64 {
	// Calculate a composite performance score (0-100)
	ratingScore := (bs.AverageRating / 5.0) * 40                // 40 points max for rating
	satisfactionScore := (bs.CustomerSatisfaction / 100.0) * 30 // 30 points max for satisfaction
	repeatScore := (bs.RepeatCustomerRate / 100.0) * 20         // 20 points max for repeat customers
	cancelScore := (1.0 - (bs.CancellationRate / 100.0)) * 10   // 10 points max for low cancellation

	return ratingScore + satisfactionScore + repeatScore + cancelScore
}

func (bs *BarberService) GetDisplayName() string {
	if bs.CustomName != nil && *bs.CustomName != "" {
		return *bs.CustomName
	}
	if bs.Service != nil {
		return bs.Service.Name
	}
	return "Service" // Fallback
}

func (bs *BarberService) HasActivePromotion() bool {
	if !bs.IsPromotional || bs.PromotionStartDate == nil || bs.PromotionEndDate == nil {
		return false
	}

	now := time.Now()
	return now.After(*bs.PromotionStartDate) && now.Before(*bs.PromotionEndDate)
}

func (bs *BarberService) CalculateRevenuePotential() float64 {
	// Estimate monthly revenue potential based on performance
	avgBookingsPerMonth := float64(bs.BookingsLast30Days)
	if avgBookingsPerMonth == 0 {
		avgBookingsPerMonth = float64(bs.TotalBookings) / 12.0 // Rough estimate
	}

	avgPrice := bs.Price
	if bs.DiscountPrice != nil && bs.HasActivePromotion() {
		avgPrice = *bs.DiscountPrice
	}

	return avgBookingsPerMonth * avgPrice
}

// Helper methods for ServiceCategory
func (sc *ServiceCategory) GetFullPath() string {
	return sc.CategoryPath
}

func (sc *ServiceCategory) IsSubCategory() bool {
	return sc.ParentCategoryID != nil && sc.Level > 1
}

func (sc *ServiceCategory) GetDisplayColor() string {
	if sc.ColorHex != nil {
		return *sc.ColorHex
	}
	return "#6B7280" // Default gray
}

// Validation methods
func (s *Service) Validate() error {
	if s.Name == "" {
		return errors.New("service name is required")
	}
	if s.CategoryID <= 0 {
		return errors.New("valid category ID is required")
	}
	if s.DefaultDurationMin <= 0 {
		return errors.New("default duration must be positive")
	}
	if s.Complexity < 1 || s.Complexity > 5 {
		return errors.New("complexity must be between 1 and 5")
	}
	return nil
}

func (bs *BarberService) Validate() error {
	if bs.BarberID <= 0 {
		return errors.New("valid barber ID is required")
	}
	if bs.ServiceID <= 0 {
		return errors.New("valid service ID is required")
	}
	if bs.Price <= 0 {
		return errors.New("price must be positive")
	}
	if bs.EstimatedDurationMin <= 0 {
		return errors.New("estimated duration must be positive")
	}
	if bs.AdvanceNoticeHours < 0 {
		return errors.New("advance notice hours cannot be negative")
	}
	return nil
}

package models

import "time"

// Review represents customer reviews and ratings
type Review struct {
	ID         int  `json:"id" db:"id"`
	BookingID  int  `json:"booking_id" db:"booking_id"` // One review per booking
	CustomerID *int `json:"customer_id" db:"customer_id"`
	BarberID   int  `json:"barber_id" db:"barber_id"`

	// Overall and detailed ratings (1-5 scale)
	OverallRating         int  `json:"overall_rating" db:"overall_rating"`
	ServiceQualityRating  *int `json:"service_quality_rating" db:"service_quality_rating"`
	PunctualityRating     *int `json:"punctuality_rating" db:"punctuality_rating"`
	CleanlinessRating     *int `json:"cleanliness_rating" db:"cleanliness_rating"`
	ValueForMoneyRating   *int `json:"value_for_money_rating" db:"value_for_money_rating"`
	ProfessionalismRating *int `json:"professionalism_rating" db:"professionalism_rating"`

	// Review content
	Title   *string `json:"title" db:"title"`
	Comment *string `json:"comment" db:"comment"`
	Pros    *string `json:"pros" db:"pros"`
	Cons    *string `json:"cons" db:"cons"`

	// Additional feedback
	WouldRecommend    *bool `json:"would_recommend" db:"would_recommend"`
	WouldBookAgain    *bool `json:"would_book_again" db:"would_book_again"`
	ServiceAsExpected *bool `json:"service_as_expected" db:"service_as_expected"`
	DurationAccurate  *bool `json:"duration_accurate" db:"duration_accurate"`

	// Media attachments
	Images StringArray `json:"images" db:"images"`

	// Verification and moderation
	IsVerified       bool       `json:"is_verified" db:"is_verified"`
	IsPublished      bool       `json:"is_published" db:"is_published"`
	ModerationStatus string     `json:"moderation_status" db:"moderation_status"` // pending, approved, rejected, flagged
	ModerationNotes  *string    `json:"moderation_notes" db:"moderation_notes"`
	ModeratedBy      *int       `json:"moderated_by" db:"moderated_by"`
	ModeratedAt      *time.Time `json:"moderated_at" db:"moderated_at"`

	// Community interaction
	HelpfulVotes int `json:"helpful_votes" db:"helpful_votes"`
	TotalVotes   int `json:"total_votes" db:"total_votes"`

	// Barber response
	BarberResponse   *string    `json:"barber_response" db:"barber_response"`
	BarberResponseAt *time.Time `json:"barber_response_at" db:"barber_response_at"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Relations
	Customer *User    `json:"customer,omitempty"`
	Barber   *Barber  `json:"barber,omitempty"`
	Booking  *Booking `json:"booking,omitempty"`
}

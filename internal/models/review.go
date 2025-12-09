package models

import (
	"errors"
	"time"
)

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

// ========================================================================
// REVIEW HELPER METHODS
// ========================================================================

// GetAverageRating calculates the average of all ratings
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

// IsPositive returns true if the review has a positive overall rating
func (r *Review) IsPositive() bool {
	return r.OverallRating >= 4
}

// GetHelpfulnessRatio returns the ratio of helpful votes to total votes
func (r *Review) GetHelpfulnessRatio() float64 {
	if r.TotalVotes == 0 {
		return 0
	}
	return float64(r.HelpfulVotes) / float64(r.TotalVotes)
}

// Validate validates review fields
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

// tests/unit/models/review_test.go
package models

import (
	"testing"

	"barber-booking-system/internal/models"
)

// ========================================================================
// REVIEW MODEL UNIT TESTS
// ========================================================================

func TestReview_GetAverageRating_OnlyOverall(t *testing.T) {
	review := &models.Review{
		OverallRating: 5,
	}

	avg := review.GetAverageRating()
	if avg != 5.0 {
		t.Errorf("Expected average rating 5.0, got %f", avg)
	}
}

func TestReview_GetAverageRating_WithAllRatings(t *testing.T) {
	serviceQuality := 4
	punctuality := 5
	cleanliness := 4
	valueForMoney := 3
	professionalism := 5

	review := &models.Review{
		OverallRating:         5,
		ServiceQualityRating:  &serviceQuality,
		PunctualityRating:     &punctuality,
		CleanlinessRating:     &cleanliness,
		ValueForMoneyRating:   &valueForMoney,
		ProfessionalismRating: &professionalism,
	}

	avg := review.GetAverageRating()
	// (5 + 4 + 5 + 4 + 3 + 5) / 6 = 26 / 6 = 4.333...
	expected := 26.0 / 6.0
	if avg != expected {
		t.Errorf("Expected average rating %f, got %f", expected, avg)
	}
}

func TestReview_GetAverageRating_PartialRatings(t *testing.T) {
	serviceQuality := 4
	punctuality := 5

	review := &models.Review{
		OverallRating:        5,
		ServiceQualityRating: &serviceQuality,
		PunctualityRating:    &punctuality,
	}

	avg := review.GetAverageRating()
	// (5 + 4 + 5) / 3 = 14 / 3 = 4.666...
	expected := 14.0 / 3.0
	if avg != expected {
		t.Errorf("Expected average rating %f, got %f", expected, avg)
	}
}

func TestReview_IsPositive_HighRating(t *testing.T) {
	testCases := []struct {
		rating   int
		expected bool
	}{
		{5, true},
		{4, true},
		{3, false},
		{2, false},
		{1, false},
	}

	for _, tc := range testCases {
		review := &models.Review{OverallRating: tc.rating}
		if review.IsPositive() != tc.expected {
			t.Errorf("Rating %d: expected IsPositive=%v, got %v", tc.rating, tc.expected, review.IsPositive())
		}
	}
}

func TestReview_GetHelpfulnessRatio_NoVotes(t *testing.T) {
	review := &models.Review{
		HelpfulVotes: 0,
		TotalVotes:   0,
	}

	ratio := review.GetHelpfulnessRatio()
	if ratio != 0 {
		t.Errorf("Expected ratio 0 with no votes, got %f", ratio)
	}
}

func TestReview_GetHelpfulnessRatio_AllHelpful(t *testing.T) {
	review := &models.Review{
		HelpfulVotes: 10,
		TotalVotes:   10,
	}

	ratio := review.GetHelpfulnessRatio()
	if ratio != 1.0 {
		t.Errorf("Expected ratio 1.0, got %f", ratio)
	}
}

func TestReview_GetHelpfulnessRatio_Mixed(t *testing.T) {
	review := &models.Review{
		HelpfulVotes: 7,
		TotalVotes:   10,
	}

	ratio := review.GetHelpfulnessRatio()
	if ratio != 0.7 {
		t.Errorf("Expected ratio 0.7, got %f", ratio)
	}
}

func TestReview_Validate_Success(t *testing.T) {
	review := &models.Review{
		BookingID:     1,
		BarberID:      1,
		OverallRating: 5,
	}

	err := review.Validate()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestReview_Validate_MissingBookingID(t *testing.T) {
	review := &models.Review{
		BookingID:     0,
		BarberID:      1,
		OverallRating: 5,
	}

	err := review.Validate()
	if err == nil {
		t.Error("Expected error for missing booking ID")
	}
}

func TestReview_Validate_MissingBarberID(t *testing.T) {
	review := &models.Review{
		BookingID:     1,
		BarberID:      0,
		OverallRating: 5,
	}

	err := review.Validate()
	if err == nil {
		t.Error("Expected error for missing barber ID")
	}
}

func TestReview_Validate_InvalidRating_TooLow(t *testing.T) {
	review := &models.Review{
		BookingID:     1,
		BarberID:      1,
		OverallRating: 0,
	}

	err := review.Validate()
	if err == nil {
		t.Error("Expected error for rating below 1")
	}
}

func TestReview_Validate_InvalidRating_TooHigh(t *testing.T) {
	review := &models.Review{
		BookingID:     1,
		BarberID:      1,
		OverallRating: 6,
	}

	err := review.Validate()
	if err == nil {
		t.Error("Expected error for rating above 5")
	}
}

func TestReview_Validate_ValidRatings(t *testing.T) {
	for rating := 1; rating <= 5; rating++ {
		review := &models.Review{
			BookingID:     1,
			BarberID:      1,
			OverallRating: rating,
		}

		err := review.Validate()
		if err != nil {
			t.Errorf("Rating %d should be valid, got error: %v", rating, err)
		}
	}
}

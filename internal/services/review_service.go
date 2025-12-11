// internal/services/review_service.go
package services

import (
	"context"
	"fmt"

	"barber-booking-system/internal/cache"
	"barber-booking-system/internal/config"
	"barber-booking-system/internal/logger"
	"barber-booking-system/internal/models"
	"barber-booking-system/internal/repository"
)

// ========================================================================
// REVIEW SERVICE - Business Logic Layer for Reviews
// ========================================================================

// ReviewService handles review business logic
type ReviewService struct {
	repo        *repository.ReviewRepository
	bookingRepo *repository.BookingRepository
	barberRepo  *repository.BarberRepository
	cache       *cache.CacheService
}

// NewReviewService creates a new review service
func NewReviewService(
	repo *repository.ReviewRepository,
	bookingRepo *repository.BookingRepository,
	barberRepo *repository.BarberRepository,
	cache *cache.CacheService,
) *ReviewService {
	return &ReviewService{
		repo:        repo,
		bookingRepo: bookingRepo,
		barberRepo:  barberRepo,
		cache:       cache,
	}
}

// ========================================================================
// REQUEST/RESPONSE STRUCTS
// ========================================================================

// CreateReviewRequest represents a request to create a review
type CreateReviewRequest struct {
	BookingID int `json:"booking_id" binding:"required"`

	// Ratings (required: overall, optional: detailed)
	OverallRating         int  `json:"overall_rating" binding:"required,min=1,max=5"`
	ServiceQualityRating  *int `json:"service_quality_rating" binding:"omitempty,min=1,max=5"`
	PunctualityRating     *int `json:"punctuality_rating" binding:"omitempty,min=1,max=5"`
	CleanlinessRating     *int `json:"cleanliness_rating" binding:"omitempty,min=1,max=5"`
	ValueForMoneyRating   *int `json:"value_for_money_rating" binding:"omitempty,min=1,max=5"`
	ProfessionalismRating *int `json:"professionalism_rating" binding:"omitempty,min=1,max=5"`

	// Content
	Title   *string `json:"title" binding:"omitempty,max=200"`
	Comment *string `json:"comment" binding:"omitempty,max=2000"`
	Pros    *string `json:"pros" binding:"omitempty,max=500"`
	Cons    *string `json:"cons" binding:"omitempty,max=500"`

	// Feedback
	WouldRecommend    *bool `json:"would_recommend"`
	WouldBookAgain    *bool `json:"would_book_again"`
	ServiceAsExpected *bool `json:"service_as_expected"`
	DurationAccurate  *bool `json:"duration_accurate"`

	// Media
	Images []string `json:"images"`
}

// UpdateReviewRequest represents a request to update a review
type UpdateReviewRequest struct {
	OverallRating         *int `json:"overall_rating" binding:"omitempty,min=1,max=5"`
	ServiceQualityRating  *int `json:"service_quality_rating" binding:"omitempty,min=1,max=5"`
	PunctualityRating     *int `json:"punctuality_rating" binding:"omitempty,min=1,max=5"`
	CleanlinessRating     *int `json:"cleanliness_rating" binding:"omitempty,min=1,max=5"`
	ValueForMoneyRating   *int `json:"value_for_money_rating" binding:"omitempty,min=1,max=5"`
	ProfessionalismRating *int `json:"professionalism_rating" binding:"omitempty,min=1,max=5"`

	Title   *string `json:"title"`
	Comment *string `json:"comment"`
	Pros    *string `json:"pros"`
	Cons    *string `json:"cons"`

	WouldRecommend    *bool `json:"would_recommend"`
	WouldBookAgain    *bool `json:"would_book_again"`
	ServiceAsExpected *bool `json:"service_as_expected"`
	DurationAccurate  *bool `json:"duration_accurate"`

	Images []string `json:"images"`
}

// ModerateReviewRequest represents a moderation action
type ModerateReviewRequest struct {
	Status string  `json:"status" binding:"required,oneof=pending approved rejected flagged"`
	Notes  *string `json:"notes"`
}

// BarberResponseRequest represents a barber's response to a review
type BarberResponseRequest struct {
	Response string `json:"response" binding:"required,min=10,max=1000"`
}

// VoteReviewRequest represents a helpfulness vote
type VoteReviewRequest struct {
	IsHelpful bool `json:"is_helpful"`
}

// ReviewResponse wraps review with additional computed fields
type ReviewResponse struct {
	*models.Review
	AverageRating    float64 `json:"average_rating"`
	HelpfulnessRatio float64 `json:"helpfulness_ratio"`
	IsPositive       bool    `json:"is_positive"`
	CanEdit          bool    `json:"can_edit"`
	CanRespond       bool    `json:"can_respond"`
	CustomerName     string  `json:"customer_name,omitempty"`
	BarberName       string  `json:"barber_name,omitempty"`
}

// ReviewStatsResponse wraps stats with additional info
type ReviewStatsResponse struct {
	*repository.ReviewStats
	RatingDistribution map[int]int `json:"rating_distribution"`
}

// ========================================================================
// HELPER FUNCTIONS
// ========================================================================

// validateRating validates a rating value
func validateRating(rating int) error {
	if rating < int(config.MinRating) || rating > int(config.MaxRating) {
		return fmt.Errorf("rating must be between %d and %d", int(config.MinRating), int(config.MaxRating))
	}
	return nil
}

// validateComment validates comment length
func validateComment(comment *string) error {
	if comment != nil && len(*comment) > 0 {
		if len(*comment) < config.MinReviewLength {
			return fmt.Errorf("comment must be at least %d characters", config.MinReviewLength)
		}
		if len(*comment) > config.MaxReviewLength {
			return fmt.Errorf("comment cannot exceed %d characters", config.MaxReviewLength)
		}
	}
	return nil
}

// toReviewResponse converts a review to a response with computed fields
func (s *ReviewService) toReviewResponse(review *models.Review, userID *int) *ReviewResponse {
	response := &ReviewResponse{
		Review:           review,
		AverageRating:    review.GetAverageRating(),
		HelpfulnessRatio: review.GetHelpfulnessRatio(),
		IsPositive:       review.IsPositive(),
		CanEdit:          false,
		CanRespond:       false,
	}

	// Check if user can edit (only the review author within 24 hours)
	if userID != nil && review.CustomerID != nil && *userID == *review.CustomerID {
		response.CanEdit = review.ModerationStatus == config.ReviewModerationPending
	}

	return response
}

// ========================================================================
// CREATE REVIEW
// ========================================================================

// CreateReview creates a new review for a completed booking
// CreateReview creates a new review for a completed booking
// Uses a transaction to ensure review creation and barber stats update are atomic
// CreateReview creates a new review for a completed booking
func (s *ReviewService) CreateReview(ctx context.Context, req CreateReviewRequest, customerID int) (*ReviewResponse, error) {
	log := logger.FromContext(ctx)

	log.Debug("Creating review").
		Int("booking_id", req.BookingID).
		Int("customer_id", customerID).
		Int("rating", req.OverallRating).
		Send()

	// Step 1: Validate rating
	if err := validateRating(req.OverallRating); err != nil {
		log.Warn("Rating validation failed").
			Int("rating", req.OverallRating).
			Err(err).
			Send()
		return nil, err
	}

	// Step 2: Validate comment if provided
	if err := validateComment(req.Comment); err != nil {
		log.Warn("Comment validation failed").
			Err(err).
			Send()
		return nil, err
	}

	// Step 3: Get and validate booking
	booking, err := s.bookingRepo.FindByID(ctx, req.BookingID)
	if err != nil {
		log.Warn("Booking not found for review").
			Int("booking_id", req.BookingID).
			Err(err).
			Send()
		return nil, fmt.Errorf("booking not found: %w", err)
	}

	// Step 4: Verify booking is completed
	if booking.Status != config.BookingStatusCompleted {
		log.Warn("Cannot review non-completed booking").
			Int("booking_id", req.BookingID).
			Str("status", booking.Status).
			Send()
		return nil, repository.ErrBookingNotCompleted
	}

	// Step 5: Verify customer owns the booking
	if booking.CustomerID == nil || *booking.CustomerID != customerID {
		log.Warn("Customer does not own booking").
			Int("booking_id", req.BookingID).
			Int("customer_id", customerID).
			Send()
		return nil, fmt.Errorf("you can only review your own bookings")
	}

	// Step 6: Check for existing review
	exists, err := s.repo.ExistsByBookingID(ctx, req.BookingID)
	if err != nil {
		return nil, err
	}
	if exists {
		log.Warn("Review already exists for booking").
			Int("booking_id", req.BookingID).
			Send()
		return nil, repository.ErrDuplicateReview
	}

	// Step 7: Build review model
	review := &models.Review{
		BookingID:  req.BookingID,
		CustomerID: &customerID,
		BarberID:   booking.BarberID,

		OverallRating:         req.OverallRating,
		ServiceQualityRating:  req.ServiceQualityRating,
		PunctualityRating:     req.PunctualityRating,
		CleanlinessRating:     req.CleanlinessRating,
		ValueForMoneyRating:   req.ValueForMoneyRating,
		ProfessionalismRating: req.ProfessionalismRating,

		Title:   req.Title,
		Comment: req.Comment,
		Pros:    req.Pros,
		Cons:    req.Cons,

		WouldRecommend:    req.WouldRecommend,
		WouldBookAgain:    req.WouldBookAgain,
		ServiceAsExpected: req.ServiceAsExpected,
		DurationAccurate:  req.DurationAccurate,

		Images: req.Images,

		IsVerified:       true,
		IsPublished:      false,
		ModerationStatus: config.ReviewModerationPending,
	}

	// Step 8: Save review
	if err := s.repo.Create(ctx, review); err != nil {
		log.Error(err).
			Int("booking_id", req.BookingID).
			Msg("Failed to create review")
		return nil, err
	}

	// Step 9: Invalidate barber cache
	if s.cache != nil {
		_ = s.cache.InvalidateBarber(ctx, booking.BarberID)
	}

	log.Info("Review created successfully").
		Int("review_id", review.ID).
		Int("booking_id", req.BookingID).
		Int("barber_id", booking.BarberID).
		Int("rating", req.OverallRating).
		Send()

	return s.toReviewResponse(review, &customerID), nil
}

// saveReviewWithStatsUpdate saves review and updates barber stats atomically
func (s *ReviewService) saveReviewWithStatsUpdate(ctx context.Context, review *models.Review) error {
	// Start transaction
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Ensure rollback on error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Create review within transaction
	if err = s.repo.CreateTx(ctx, tx, review); err != nil {
		return err
	}

	// Update barber rating stats within same transaction
	// Note: Stats only count published/approved reviews, so this ensures
	// the barber's denormalized stats stay in sync
	if err = s.barberRepo.UpdateRatingStatsTx(ctx, tx, review.BarberID); err != nil {
		return fmt.Errorf("failed to update barber stats: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ========================================================================
// READ OPERATIONS
// ========================================================================

// GetReviewByID retrieves a review by ID
func (s *ReviewService) GetReviewByID(ctx context.Context, id int, userID *int) (*ReviewResponse, error) {
	review, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toReviewResponse(review, userID), nil
}

// GetReviewByBookingID retrieves a review by booking ID
func (s *ReviewService) GetReviewByBookingID(ctx context.Context, bookingID int, userID *int) (*ReviewResponse, error) {
	review, err := s.repo.FindByBookingID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	return s.toReviewResponse(review, userID), nil
}

// GetBarberReviews retrieves all published reviews for a barber
func (s *ReviewService) GetBarberReviews(ctx context.Context, barberID int, filters repository.ReviewFilters) ([]ReviewResponse, error) {
	// Only show published and approved reviews to the public
	isPublished := true
	filters.IsPublished = &isPublished
	filters.ModerationStatus = config.ReviewModerationApproved

	reviews, err := s.repo.FindByBarberID(ctx, barberID, filters)
	if err != nil {
		return nil, err
	}

	responses := make([]ReviewResponse, len(reviews))
	for i, review := range reviews {
		responses[i] = *s.toReviewResponse(&review, nil)
	}
	return responses, nil
}

// GetCustomerReviews retrieves all reviews by a customer
func (s *ReviewService) GetCustomerReviews(ctx context.Context, customerID int, filters repository.ReviewFilters) ([]ReviewResponse, error) {
	reviews, err := s.repo.FindByCustomerID(ctx, customerID, filters)
	if err != nil {
		return nil, err
	}

	responses := make([]ReviewResponse, len(reviews))
	for i, review := range reviews {
		responses[i] = *s.toReviewResponse(&review, &customerID)
	}
	return responses, nil
}

// GetPendingReviews retrieves reviews pending moderation (admin only)
func (s *ReviewService) GetPendingReviews(ctx context.Context, filters repository.ReviewFilters) ([]ReviewResponse, error) {
	filters.ModerationStatus = config.ReviewModerationPending

	reviews, err := s.repo.FindAll(ctx, filters)
	if err != nil {
		return nil, err
	}

	responses := make([]ReviewResponse, len(reviews))
	for i, review := range reviews {
		responses[i] = *s.toReviewResponse(&review, nil)
	}
	return responses, nil
}

// ========================================================================
// UPDATE OPERATIONS
// ========================================================================

// UpdateReview updates a review (only by author and before moderation)
func (s *ReviewService) UpdateReview(ctx context.Context, id int, req UpdateReviewRequest, customerID int) (*ReviewResponse, error) {
	log := logger.FromContext(ctx)

	log.Debug("Updating review").
		Int("review_id", id).
		Int("customer_id", customerID).
		Send()

	// Get existing review
	review, err := s.repo.FindByID(ctx, id)
	if err != nil {
		log.Warn("Review not found").
			Int("review_id", id).
			Err(err).
			Send()
		return nil, err
	}

	// Verify ownership
	if review.CustomerID == nil || *review.CustomerID != customerID {
		log.Warn("Unauthorized review update attempt").
			Int("review_id", id).
			Int("customer_id", customerID).
			Send()
		return nil, fmt.Errorf("you can only edit your own reviews")
	}

	// Check if review can be edited
	if review.ModerationStatus != config.ReviewModerationPending {
		log.Warn("Cannot edit moderated review").
			Int("review_id", id).
			Str("moderation_status", review.ModerationStatus).
			Send()
		return nil, repository.ErrCannotModifyReview
	}

	// Update fields if provided
	if req.OverallRating != nil {
		if err := validateRating(*req.OverallRating); err != nil {
			return nil, err
		}
		review.OverallRating = *req.OverallRating
	}
	if req.ServiceQualityRating != nil {
		review.ServiceQualityRating = req.ServiceQualityRating
	}
	if req.PunctualityRating != nil {
		review.PunctualityRating = req.PunctualityRating
	}
	if req.CleanlinessRating != nil {
		review.CleanlinessRating = req.CleanlinessRating
	}
	if req.ValueForMoneyRating != nil {
		review.ValueForMoneyRating = req.ValueForMoneyRating
	}
	if req.ProfessionalismRating != nil {
		review.ProfessionalismRating = req.ProfessionalismRating
	}
	if req.Title != nil {
		review.Title = req.Title
	}
	if req.Comment != nil {
		if err := validateComment(req.Comment); err != nil {
			return nil, err
		}
		review.Comment = req.Comment
	}
	if req.Pros != nil {
		review.Pros = req.Pros
	}
	if req.Cons != nil {
		review.Cons = req.Cons
	}
	if req.WouldRecommend != nil {
		review.WouldRecommend = req.WouldRecommend
	}
	if req.WouldBookAgain != nil {
		review.WouldBookAgain = req.WouldBookAgain
	}
	if req.ServiceAsExpected != nil {
		review.ServiceAsExpected = req.ServiceAsExpected
	}
	if req.DurationAccurate != nil {
		review.DurationAccurate = req.DurationAccurate
	}
	if req.Images != nil {
		review.Images = req.Images
	}

	// Save changes
	if err := s.repo.Update(ctx, review); err != nil {
		log.Error(err).
			Int("review_id", id).
			Msg("Failed to update review")
		return nil, err
	}

	log.Info("Review updated successfully").
		Int("review_id", id).
		Int("barber_id", review.BarberID).
		Send()

	return s.toReviewResponse(review, &customerID), nil
}

// ModerateReview updates the moderation status (admin only)
func (s *ReviewService) ModerateReview(ctx context.Context, id int, req ModerateReviewRequest, moderatorID int) (*ReviewResponse, error) {
	log := logger.FromContext(ctx)

	log.Info("Moderating review").
		Int("review_id", id).
		Int("moderator_id", moderatorID).
		Str("new_status", req.Status).
		Send()

	// Get existing review
	review, err := s.repo.FindByID(ctx, id)
	if err != nil {
		log.Warn("Review not found for moderation").
			Int("review_id", id).
			Err(err).
			Send()
		return nil, err
	}

	oldStatus := review.ModerationStatus

	// Update moderation status
	if err := s.repo.UpdateModerationStatus(ctx, id, req.Status, moderatorID, req.Notes); err != nil {
		log.Error(err).
			Int("review_id", id).
			Msg("Failed to moderate review")
		return nil, err
	}

	// Invalidate barber cache if review is now published
	if req.Status == config.ReviewModerationApproved && s.cache != nil {
		_ = s.cache.InvalidateBarber(ctx, review.BarberID)
	}

	log.Info("Review moderated successfully").
		Int("review_id", id).
		Str("old_status", oldStatus).
		Str("new_status", req.Status).
		Int("barber_id", review.BarberID).
		Send()

	// Fetch updated review
	return s.GetReviewByID(ctx, id, nil)
}

// AddBarberResponse allows a barber to respond to a review
func (s *ReviewService) AddBarberResponse(ctx context.Context, id int, req BarberResponseRequest, barberUserID int) (*ReviewResponse, error) {
	// Get existing review
	review, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify the barber owns this review (need to check barber profile)
	barber, err := s.barberRepo.FindByUserID(ctx, barberUserID)
	if err != nil {
		return nil, fmt.Errorf("barber profile not found")
	}

	if barber.ID != review.BarberID {
		return nil, fmt.Errorf("you can only respond to your own reviews")
	}

	// Check if already responded
	if review.BarberResponse != nil && *review.BarberResponse != "" {
		return nil, fmt.Errorf("you have already responded to this review")
	}

	// Add response
	if err := s.repo.AddBarberResponse(ctx, id, req.Response); err != nil {
		return nil, err
	}

	// Fetch updated review
	return s.GetReviewByID(ctx, id, nil)
}

// VoteReview allows users to vote on review helpfulness
func (s *ReviewService) VoteReview(ctx context.Context, id int, req VoteReviewRequest) error {
	// Increment vote counter
	return s.repo.IncrementHelpfulVotes(ctx, id, req.IsHelpful)
}

// ========================================================================
// DELETE OPERATIONS
// ========================================================================

// DeleteReview soft deletes a review
func (s *ReviewService) DeleteReview(ctx context.Context, id int, userID int, isAdmin bool) error {
	log := logger.FromContext(ctx)

	log.Info("Deleting review").
		Int("review_id", id).
		Int("user_id", userID).
		Bool("is_admin", isAdmin).
		Send()

	// Get review
	review, err := s.repo.FindByID(ctx, id)
	if err != nil {
		log.Warn("Review not found for deletion").
			Int("review_id", id).
			Err(err).
			Send()
		return err
	}

	// Check authorization
	if !isAdmin {
		if review.CustomerID == nil || *review.CustomerID != userID {
			log.Warn("Unauthorized review deletion attempt").
				Int("review_id", id).
				Int("user_id", userID).
				Send()
			return fmt.Errorf("you can only delete your own reviews")
		}
	}

	// Soft delete
	if err := s.repo.Delete(ctx, id); err != nil {
		log.Error(err).
			Int("review_id", id).
			Msg("Failed to delete review")
		return err
	}

	// Invalidate barber cache
	if s.cache != nil {
		_ = s.cache.InvalidateBarber(ctx, review.BarberID)
	}

	log.Info("Review deleted successfully").
		Int("review_id", id).
		Int("barber_id", review.BarberID).
		Send()

	return nil
}

// ========================================================================
// STATISTICS
// ========================================================================

// GetBarberReviewStats retrieves review statistics for a barber
func (s *ReviewService) GetBarberReviewStats(ctx context.Context, barberID int) (*ReviewStatsResponse, error) {
	stats, err := s.repo.GetBarberStats(ctx, barberID)
	if err != nil {
		return nil, err
	}

	response := &ReviewStatsResponse{
		ReviewStats: stats,
		RatingDistribution: map[int]int{
			5: stats.FiveStarCount,
			4: stats.FourStarCount,
			3: stats.ThreeStarCount,
			2: stats.TwoStarCount,
			1: stats.OneStarCount,
		},
	}

	return response, nil
}

// CanReviewBooking checks if a customer can review a specific booking
func (s *ReviewService) CanReviewBooking(ctx context.Context, bookingID int, customerID int) (bool, string, error) {
	// Get booking
	booking, err := s.bookingRepo.FindByID(ctx, bookingID)
	if err != nil {
		return false, "Booking not found", err
	}

	// Check if customer owns booking
	if booking.CustomerID == nil || *booking.CustomerID != customerID {
		return false, "You can only review your own bookings", nil
	}

	// Check if booking is completed
	if booking.Status != config.BookingStatusCompleted {
		return false, "You can only review completed bookings", nil
	}

	// Check if review already exists
	exists, err := s.repo.ExistsByBookingID(ctx, bookingID)
	if err != nil {
		return false, "Error checking existing review", err
	}
	if exists {
		return false, "You have already reviewed this booking", nil
	}

	return true, "", nil
}

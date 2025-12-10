// internal/handlers/review_handler.go
package handlers

import (
	"net/http"

	"barber-booking-system/internal/middleware"
	"barber-booking-system/internal/repository"
	"barber-booking-system/internal/services"
	"barber-booking-system/internal/utils"

	"github.com/gin-gonic/gin"
)

// ========================================================================
// REVIEW HANDLER - HTTP Request Handlers for Reviews
// ========================================================================

// ReviewHandler handles review-related HTTP requests
type ReviewHandler struct {
	reviewService *services.ReviewService
}

// NewReviewHandler creates a new review handler
func NewReviewHandler(reviewService *services.ReviewService) *ReviewHandler {
	return &ReviewHandler{
		reviewService: reviewService,
	}
}

// ========================================================================
// HELPER FUNCTIONS
// ========================================================================

// buildReviewFilters builds ReviewFilters from query parameters
func buildReviewFilters(c *gin.Context) repository.ReviewFilters {
	filters := repository.ReviewFilters{
		ModerationStatus: c.Query("moderation_status"),
		Search:           c.Query("search"),
		SortBy:           c.Query("sort_by"),
		Order:            c.Query("order"),
		Limit:            ParseIntQuery(c, "limit", 50),
		Offset:           ParseIntQuery(c, "offset", 0),
		MinRating:        ParseIntQuery(c, "min_rating", 0),
		MaxRating:        ParseIntQuery(c, "max_rating", 0),
		CreatedFrom:      ParseTimeQuery(c, "created_from"),
		CreatedTo:        ParseTimeQuery(c, "created_to"),
	}

	// Boolean filters
	if isPublished := ParseBoolQuery(c, "is_published"); isPublished != nil {
		filters.IsPublished = isPublished
	}
	if isVerified := ParseBoolQuery(c, "is_verified"); isVerified != nil {
		filters.IsVerified = isVerified
	}
	if hasComment := ParseBoolQuery(c, "has_comment"); hasComment != nil {
		filters.HasComment = hasComment
	}
	if hasImages := ParseBoolQuery(c, "has_images"); hasImages != nil {
		filters.HasImages = hasImages
	}
	if hasResponse := ParseBoolQuery(c, "has_response"); hasResponse != nil {
		filters.HasResponse = hasResponse
	}
	if wouldRecommend := ParseBoolQuery(c, "would_recommend"); wouldRecommend != nil {
		filters.WouldRecommend = wouldRecommend
	}

	return filters
}

// ========================================================================
// CREATE REVIEW
// ========================================================================

// CreateReview godoc
// @Summary Create a new review
// @Description Create a review for a completed booking
// @Tags reviews
// @Accept json
// @Produce json
// @Param review body services.CreateReviewRequest true "Review data"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 409 {object} middleware.ErrorResponse "Review already exists"
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/reviews [post]
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	req, ok := BindJSON[services.CreateReviewRequest](c)
	if !ok {
		return
	}

	// Get authenticated user
	userID, exists := middleware.GetUserID(c)
	if !exists {
		RespondUnauthorized(c, "You must be logged in to create a review")
		return
	}

	// Create review
	review, err := h.reviewService.CreateReview(c.Request.Context(), *req, userID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch err {
		case repository.ErrBookingNotCompleted:
			statusCode = http.StatusBadRequest
		case repository.ErrReviewAlreadyExists, repository.ErrDuplicateReview:
			statusCode = http.StatusConflict
		default:
			if utils.ContainsAny(err.Error(), []string{"not found", "required", "must be", "only review"}) {
				statusCode = http.StatusBadRequest
			}
		}

		c.JSON(statusCode, middleware.ErrorResponse{
			Error:   "Failed to create review",
			Message: err.Error(),
		})
		return
	}

	RespondCreated(c, review, "Review created successfully")
}

// ========================================================================
// GET REVIEW BY ID
// ========================================================================

// GetReview godoc
// @Summary Get review by ID
// @Description Get detailed information about a specific review
// @Tags reviews
// @Accept json
// @Produce json
// @Param id path int true "Review ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/reviews/{id} [get]
func (h *ReviewHandler) GetReview(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "review")
	if !ok {
		return
	}

	// Get user ID if authenticated (optional)
	var userID *int
	if uid, exists := middleware.GetUserID(c); exists {
		userID = &uid
	}

	review, err := h.reviewService.GetReviewByID(c.Request.Context(), id, userID)
	if err != nil {
		if err == repository.ErrReviewNotFound {
			RespondNotFound(c, "Review")
			return
		}
		RespondInternalError(c, "fetch review", err)
		return
	}

	RespondSuccess(c, review)
}

// GetReviewByBooking godoc
// @Summary Get review by booking ID
// @Description Get the review for a specific booking
// @Tags reviews
// @Accept json
// @Produce json
// @Param booking_id path int true "Booking ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/reviews/booking/{booking_id} [get]
func (h *ReviewHandler) GetReviewByBooking(c *gin.Context) {
	bookingID, ok := RequireIntParam(c, "booking_id", "booking")
	if !ok {
		return
	}

	var userID *int
	if uid, exists := middleware.GetUserID(c); exists {
		userID = &uid
	}

	review, err := h.reviewService.GetReviewByBookingID(c.Request.Context(), bookingID, userID)
	if err != nil {
		if err == repository.ErrReviewNotFound {
			RespondNotFound(c, "Review")
			return
		}
		RespondInternalError(c, "fetch review", err)
		return
	}

	RespondSuccess(c, review)
}

// ========================================================================
// GET BARBER REVIEWS
// ========================================================================

// GetBarberReviews godoc
// @Summary Get barber's reviews
// @Description Get all published reviews for a specific barber
// @Tags reviews
// @Accept json
// @Produce json
// @Param id path int true "Barber ID"
// @Param min_rating query int false "Filter by minimum rating"
// @Param max_rating query int false "Filter by maximum rating"
// @Param has_comment query bool false "Filter reviews with comments"
// @Param has_images query bool false "Filter reviews with images"
// @Param sort_by query string false "Sort by field" default(created_at)
// @Param order query string false "Sort order (ASC/DESC)" default(DESC)
// @Param limit query int false "Limit results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/barbers/{id}/reviews [get]
func (h *ReviewHandler) GetBarberReviews(c *gin.Context) {
	barberID, ok := RequireIntParam(c, "id", "barber")
	if !ok {
		return
	}

	filters := buildReviewFilters(c)

	reviews, err := h.reviewService.GetBarberReviews(c.Request.Context(), barberID, filters)
	if err != nil {
		RespondInternalError(c, "fetch reviews", err)
		return
	}

	RespondSuccessWithMeta(c, reviews, map[string]interface{}{
		"barber_id": barberID,
		"count":     len(reviews),
		"limit":     filters.Limit,
		"offset":    filters.Offset,
	})
}

// GetBarberReviewStats godoc
// @Summary Get barber's review statistics
// @Description Get review statistics (ratings distribution, average) for a barber
// @Tags reviews
// @Accept json
// @Produce json
// @Param id path int true "Barber ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/barbers/{id}/reviews/stats [get]
func (h *ReviewHandler) GetBarberReviewStats(c *gin.Context) {
	barberID, ok := RequireIntParam(c, "id", "barber")
	if !ok {
		return
	}

	stats, err := h.reviewService.GetBarberReviewStats(c.Request.Context(), barberID)
	if err != nil {
		RespondInternalError(c, "fetch review stats", err)
		return
	}

	RespondSuccessWithMeta(c, stats, map[string]interface{}{
		"barber_id": barberID,
	})
}

// ========================================================================
// GET MY REVIEWS (Customer)
// ========================================================================

// GetMyReviews godoc
// @Summary Get my reviews
// @Description Get all reviews created by the authenticated customer
// @Tags reviews
// @Accept json
// @Produce json
// @Param moderation_status query string false "Filter by moderation status"
// @Param sort_by query string false "Sort by field" default(created_at)
// @Param order query string false "Sort order (ASC/DESC)" default(DESC)
// @Param limit query int false "Limit results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/reviews/me [get]
func (h *ReviewHandler) GetMyReviews(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		RespondUnauthorized(c, "You must be logged in to view your reviews")
		return
	}

	filters := buildReviewFilters(c)

	reviews, err := h.reviewService.GetCustomerReviews(c.Request.Context(), userID, filters)
	if err != nil {
		RespondInternalError(c, "fetch reviews", err)
		return
	}

	RespondSuccessWithMeta(c, reviews, PaginationMeta(len(reviews), filters.Limit, filters.Offset))
}

// ========================================================================
// UPDATE REVIEW
// ========================================================================

// UpdateReview godoc
// @Summary Update a review
// @Description Update a review (only by author and before moderation)
// @Tags reviews
// @Accept json
// @Produce json
// @Param id path int true "Review ID"
// @Param review body services.UpdateReviewRequest true "Updated review data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 403 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/reviews/{id} [put]
func (h *ReviewHandler) UpdateReview(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "review")
	if !ok {
		return
	}

	req, ok := BindJSON[services.UpdateReviewRequest](c)
	if !ok {
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		RespondUnauthorized(c, "You must be logged in to update a review")
		return
	}

	review, err := h.reviewService.UpdateReview(c.Request.Context(), id, *req, userID)
	if err != nil {
		if err == repository.ErrReviewNotFound {
			RespondNotFound(c, "Review")
			return
		}
		if err == repository.ErrCannotModifyReview {
			c.JSON(http.StatusForbidden, middleware.ErrorResponse{
				Error:   "Cannot modify review",
				Message: "This review has already been moderated and cannot be edited",
			})
			return
		}
		if utils.ContainsAny(err.Error(), []string{"only edit your own"}) {
			c.JSON(http.StatusForbidden, middleware.ErrorResponse{
				Error:   "Forbidden",
				Message: err.Error(),
			})
			return
		}
		RespondInternalError(c, "update review", err)
		return
	}

	RespondSuccessWithData(c, review, "Review updated successfully")
}

// ========================================================================
// MODERATE REVIEW (Admin)
// ========================================================================

// ModerateReview godoc
// @Summary Moderate a review
// @Description Update the moderation status of a review (admin only)
// @Tags reviews
// @Accept json
// @Produce json
// @Param id path int true "Review ID"
// @Param moderation body services.ModerateReviewRequest true "Moderation action"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 403 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/reviews/{id}/moderate [patch]
func (h *ReviewHandler) ModerateReview(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "review")
	if !ok {
		return
	}

	req, ok := BindJSON[services.ModerateReviewRequest](c)
	if !ok {
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		RespondUnauthorized(c, "You must be logged in to moderate reviews")
		return
	}

	// Note: In a real app, you'd check if user is admin here
	// userType, _ := middleware.GetUserType(c)
	// if userType != "admin" { ... }

	review, err := h.reviewService.ModerateReview(c.Request.Context(), id, *req, userID)
	if err != nil {
		if err == repository.ErrReviewNotFound {
			RespondNotFound(c, "Review")
			return
		}
		if err == repository.ErrInvalidModeration {
			c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
				Error:   "Invalid moderation status",
				Message: err.Error(),
			})
			return
		}
		RespondInternalError(c, "moderate review", err)
		return
	}

	RespondSuccessWithData(c, review, "Review moderation updated successfully")
}

// GetPendingReviews godoc
// @Summary Get pending reviews
// @Description Get all reviews pending moderation (admin only)
// @Tags reviews
// @Accept json
// @Produce json
// @Param limit query int false "Limit results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 403 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/reviews/pending [get]
func (h *ReviewHandler) GetPendingReviews(c *gin.Context) {
	filters := buildReviewFilters(c)

	reviews, err := h.reviewService.GetPendingReviews(c.Request.Context(), filters)
	if err != nil {
		RespondInternalError(c, "fetch pending reviews", err)
		return
	}

	RespondSuccessWithMeta(c, reviews, PaginationMeta(len(reviews), filters.Limit, filters.Offset))
}

// ========================================================================
// BARBER RESPONSE
// ========================================================================

// AddBarberResponse godoc
// @Summary Add barber response to a review
// @Description Allow a barber to respond to a review on their profile
// @Tags reviews
// @Accept json
// @Produce json
// @Param id path int true "Review ID"
// @Param response body services.BarberResponseRequest true "Response content"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 403 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/reviews/{id}/response [post]
func (h *ReviewHandler) AddBarberResponse(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "review")
	if !ok {
		return
	}

	req, ok := BindJSON[services.BarberResponseRequest](c)
	if !ok {
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		RespondUnauthorized(c, "You must be logged in to respond to reviews")
		return
	}

	review, err := h.reviewService.AddBarberResponse(c.Request.Context(), id, *req, userID)
	if err != nil {
		if err == repository.ErrReviewNotFound {
			RespondNotFound(c, "Review")
			return
		}
		if utils.ContainsAny(err.Error(), []string{"only respond to your own", "already responded"}) {
			c.JSON(http.StatusForbidden, middleware.ErrorResponse{
				Error:   "Cannot respond",
				Message: err.Error(),
			})
			return
		}
		RespondInternalError(c, "add barber response", err)
		return
	}

	RespondSuccessWithData(c, review, "Response added successfully")
}

// ========================================================================
// VOTE ON REVIEW
// ========================================================================

// VoteReview godoc
// @Summary Vote on review helpfulness
// @Description Vote whether a review was helpful or not
// @Tags reviews
// @Accept json
// @Produce json
// @Param id path int true "Review ID"
// @Param vote body services.VoteReviewRequest true "Vote data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/reviews/{id}/vote [post]
func (h *ReviewHandler) VoteReview(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "review")
	if !ok {
		return
	}

	req, ok := BindJSON[services.VoteReviewRequest](c)
	if !ok {
		return
	}

	err := h.reviewService.VoteReview(c.Request.Context(), id, *req)
	if err != nil {
		if err == repository.ErrReviewNotFound {
			RespondNotFound(c, "Review")
			return
		}
		RespondInternalError(c, "vote on review", err)
		return
	}

	RespondSuccessWithMessage(c, "Vote recorded successfully")
}

// ========================================================================
// DELETE REVIEW
// ========================================================================

// DeleteReview godoc
// @Summary Delete a review
// @Description Delete a review (soft delete by author or admin)
// @Tags reviews
// @Accept json
// @Produce json
// @Param id path int true "Review ID"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 403 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/reviews/{id} [delete]
func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "review")
	if !ok {
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		RespondUnauthorized(c, "You must be logged in to delete a review")
		return
	}

	// Note: Check if admin in a real app
	isAdmin := false // middleware.IsAdmin(c)

	err := h.reviewService.DeleteReview(c.Request.Context(), id, userID, isAdmin)
	if err != nil {
		if err == repository.ErrReviewNotFound {
			RespondNotFound(c, "Review")
			return
		}
		if utils.ContainsAny(err.Error(), []string{"only delete your own"}) {
			c.JSON(http.StatusForbidden, middleware.ErrorResponse{
				Error:   "Forbidden",
				Message: err.Error(),
			})
			return
		}
		RespondInternalError(c, "delete review", err)
		return
	}

	RespondSuccessWithMessage(c, "Review deleted successfully")
}

// ========================================================================
// CHECK IF CAN REVIEW
// ========================================================================

// CanReviewBooking godoc
// @Summary Check if booking can be reviewed
// @Description Check if the authenticated user can review a specific booking
// @Tags reviews
// @Accept json
// @Produce json
// @Param booking_id path int true "Booking ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/reviews/can-review/{booking_id} [get]
func (h *ReviewHandler) CanReviewBooking(c *gin.Context) {
	bookingID, ok := RequireIntParam(c, "booking_id", "booking")
	if !ok {
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		RespondUnauthorized(c, "You must be logged in to check review eligibility")
		return
	}

	canReview, reason, err := h.reviewService.CanReviewBooking(c.Request.Context(), bookingID, userID)
	if err != nil {
		// Check if it's a "not found" error
		if err == repository.ErrBookingNotFound || utils.ContainsAny(err.Error(), []string{"not found"}) {
			RespondNotFound(c, "Booking")
			return
		}
		RespondInternalError(c, "check review eligibility", err)
		return
	}

	RespondSuccess(c, map[string]interface{}{
		"can_review": canReview,
		"reason":     reason,
		"booking_id": bookingID,
	})
}

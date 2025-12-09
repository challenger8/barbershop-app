// internal/handlers/swagger_models.go
// This file contains Swagger model definitions for API documentation.
// NOTE: SuccessResponse and PaginatedResponse are defined in response.go
package handlers

import "time"

// ========================================================================
// REVIEW SWAGGER MODELS
// ========================================================================

// ReviewCreateRequest represents a request to create a review
// @Description Request body for creating a new review
type ReviewCreateRequest struct {
	BookingID             int    `json:"booking_id" example:"123" binding:"required"`
	OverallRating         int    `json:"overall_rating" example:"5" binding:"required,min=1,max=5"`
	ServiceQualityRating  *int   `json:"service_quality_rating,omitempty" example:"5"`
	PunctualityRating     *int   `json:"punctuality_rating,omitempty" example:"4"`
	CleanlinessRating     *int   `json:"cleanliness_rating,omitempty" example:"5"`
	ValueForMoneyRating   *int   `json:"value_for_money_rating,omitempty" example:"4"`
	ProfessionalismRating *int   `json:"professionalism_rating,omitempty" example:"5"`
	Title                 string `json:"title,omitempty" example:"Great haircut!"`
	Comment               string `json:"comment,omitempty" example:"The barber was very professional and skilled."`
	Pros                  string `json:"pros,omitempty" example:"Friendly staff, clean environment"`
	Cons                  string `json:"cons,omitempty" example:"Waiting time was a bit long"`
	WouldRecommend        *bool  `json:"would_recommend,omitempty" example:"true"`
	WouldBookAgain        *bool  `json:"would_book_again,omitempty" example:"true"`
	ServiceAsExpected     *bool  `json:"service_as_expected,omitempty" example:"true"`
	DurationAccurate      *bool  `json:"duration_accurate,omitempty" example:"true"`
	Images                []string `json:"images,omitempty"`
}

// ReviewUpdateRequest represents a request to update a review
// @Description Request body for updating an existing review
type ReviewUpdateRequest struct {
	OverallRating         *int   `json:"overall_rating,omitempty" example:"4"`
	ServiceQualityRating  *int   `json:"service_quality_rating,omitempty" example:"4"`
	PunctualityRating     *int   `json:"punctuality_rating,omitempty" example:"4"`
	CleanlinessRating     *int   `json:"cleanliness_rating,omitempty" example:"5"`
	ValueForMoneyRating   *int   `json:"value_for_money_rating,omitempty" example:"4"`
	ProfessionalismRating *int   `json:"professionalism_rating,omitempty" example:"5"`
	Title                 string `json:"title,omitempty" example:"Updated review title"`
	Comment               string `json:"comment,omitempty" example:"Updated comment text"`
	Pros                  string `json:"pros,omitempty" example:"Updated pros"`
	Cons                  string `json:"cons,omitempty" example:"Updated cons"`
	WouldRecommend        *bool  `json:"would_recommend,omitempty" example:"true"`
	WouldBookAgain        *bool  `json:"would_book_again,omitempty" example:"true"`
	ServiceAsExpected     *bool  `json:"service_as_expected,omitempty" example:"true"`
	DurationAccurate      *bool  `json:"duration_accurate,omitempty" example:"true"`
	Images                []string `json:"images,omitempty"`
}

// ReviewModerateRequest represents a moderation action request
// @Description Request body for moderating a review (admin only)
type ReviewModerateRequest struct {
	Status string `json:"status" example:"approved" binding:"required,oneof=pending approved rejected flagged"`
	Notes  string `json:"notes,omitempty" example:"Review meets community guidelines"`
}

// ReviewBarberResponseRequest represents a barber's response to a review
// @Description Request body for barber responding to a review
type ReviewBarberResponseRequest struct {
	Response string `json:"response" example:"Thank you for your feedback!" binding:"required,min=10,max=1000"`
}

// ReviewVoteRequest represents a helpfulness vote
// @Description Request body for voting on review helpfulness
type ReviewVoteRequest struct {
	IsHelpful bool `json:"is_helpful" example:"true"`
}

// ReviewResponse represents a review in API responses
// @Description Review data returned in API responses
type ReviewResponse struct {
	ID                    int        `json:"id" example:"1"`
	BookingID             int        `json:"booking_id" example:"123"`
	CustomerID            *int       `json:"customer_id" example:"456"`
	BarberID              int        `json:"barber_id" example:"789"`
	OverallRating         int        `json:"overall_rating" example:"5"`
	ServiceQualityRating  *int       `json:"service_quality_rating,omitempty" example:"5"`
	PunctualityRating     *int       `json:"punctuality_rating,omitempty" example:"4"`
	CleanlinessRating     *int       `json:"cleanliness_rating,omitempty" example:"5"`
	ValueForMoneyRating   *int       `json:"value_for_money_rating,omitempty" example:"4"`
	ProfessionalismRating *int       `json:"professionalism_rating,omitempty" example:"5"`
	Title                 *string    `json:"title,omitempty" example:"Great haircut!"`
	Comment               *string    `json:"comment,omitempty" example:"The barber was very professional."`
	Pros                  *string    `json:"pros,omitempty" example:"Friendly staff"`
	Cons                  *string    `json:"cons,omitempty" example:"Long wait"`
	WouldRecommend        *bool      `json:"would_recommend,omitempty" example:"true"`
	WouldBookAgain        *bool      `json:"would_book_again,omitempty" example:"true"`
	ServiceAsExpected     *bool      `json:"service_as_expected,omitempty" example:"true"`
	DurationAccurate      *bool      `json:"duration_accurate,omitempty" example:"true"`
	Images                []string   `json:"images,omitempty"`
	IsVerified            bool       `json:"is_verified" example:"true"`
	IsPublished           bool       `json:"is_published" example:"true"`
	ModerationStatus      string     `json:"moderation_status" example:"approved"`
	ModerationNotes       *string    `json:"moderation_notes,omitempty"`
	ModeratedBy           *int       `json:"moderated_by,omitempty"`
	ModeratedAt           *time.Time `json:"moderated_at,omitempty"`
	HelpfulVotes          int        `json:"helpful_votes" example:"15"`
	TotalVotes            int        `json:"total_votes" example:"20"`
	BarberResponse        *string    `json:"barber_response,omitempty" example:"Thank you for your feedback!"`
	BarberResponseAt      *time.Time `json:"barber_response_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
	AverageRating         float64    `json:"average_rating" example:"4.5"`
	HelpfulnessRatio      float64    `json:"helpfulness_ratio" example:"0.75"`
	IsPositive            bool       `json:"is_positive" example:"true"`
	CanEdit               bool       `json:"can_edit" example:"false"`
	CanRespond            bool       `json:"can_respond" example:"true"`
}

// ReviewStatsResponse represents review statistics for a barber
// @Description Statistics about reviews for a barber
type ReviewStatsResponse struct {
	TotalReviews       int            `json:"total_reviews" example:"150"`
	AverageRating      float64        `json:"average_rating" example:"4.7"`
	FiveStarCount      int            `json:"five_star_count" example:"100"`
	FourStarCount      int            `json:"four_star_count" example:"35"`
	ThreeStarCount     int            `json:"three_star_count" example:"10"`
	TwoStarCount       int            `json:"two_star_count" example:"3"`
	OneStarCount       int            `json:"one_star_count" example:"2"`
	RecommendPercent   float64        `json:"recommend_percent" example:"95.5"`
	RatingDistribution map[int]int    `json:"rating_distribution"`
}

// CanReviewResponse represents the response for checking review eligibility
// @Description Response indicating if a booking can be reviewed
type CanReviewResponse struct {
	CanReview bool   `json:"can_review" example:"true"`
	Reason    string `json:"reason,omitempty" example:""`
	BookingID int    `json:"booking_id" example:"123"`
}

// ========================================================================
// NOTIFICATION SWAGGER MODELS
// ========================================================================

// NotificationCreateRequest represents a request to create a notification
// @Description Request body for creating a notification (admin only)
type NotificationCreateRequest struct {
	UserID            int                    `json:"user_id" example:"123" binding:"required"`
	Title             string                 `json:"title" example:"Booking Confirmed" binding:"required,min=1,max=200"`
	Message           string                 `json:"message" example:"Your booking has been confirmed." binding:"required,min=1,max=2000"`
	Type              string                 `json:"type" example:"booking_confirmation" binding:"required"`
	Priority          string                 `json:"priority,omitempty" example:"normal"`
	Channels          []string               `json:"channels,omitempty" example:"app,email"`
	RelatedEntityType *string                `json:"related_entity_type,omitempty" example:"booking"`
	RelatedEntityID   *int                   `json:"related_entity_id,omitempty" example:"456"`
	Data              map[string]interface{} `json:"data,omitempty"`
	ScheduledFor      *time.Time             `json:"scheduled_for,omitempty"`
	ExpiresAt         *time.Time             `json:"expires_at,omitempty"`
}

// NotificationBookingRequest represents a booking notification request
// @Description Request body for sending booking-related notifications
type NotificationBookingRequest struct {
	BookingID        int    `json:"booking_id" example:"123" binding:"required"`
	NotificationType string `json:"notification_type" example:"booking_confirmation" binding:"required"`
	CustomMessage    string `json:"custom_message,omitempty" example:"Your appointment is confirmed!"`
}

// NotificationResponse represents a notification in API responses
// @Description Notification data returned in API responses
type NotificationResponse struct {
	ID                int                    `json:"id" example:"1"`
	UserID            int                    `json:"user_id" example:"123"`
	Title             string                 `json:"title" example:"Booking Confirmed"`
	Message           string                 `json:"message" example:"Your booking #BK123 has been confirmed."`
	Type              string                 `json:"type" example:"booking_confirmation"`
	Channels          []string               `json:"channels" example:"app,email"`
	Status            string                 `json:"status" example:"delivered"`
	SentAt            *time.Time             `json:"sent_at,omitempty"`
	DeliveredAt       *time.Time             `json:"delivered_at,omitempty"`
	ReadAt            *time.Time             `json:"read_at,omitempty"`
	RelatedEntityType *string                `json:"related_entity_type,omitempty" example:"booking"`
	RelatedEntityID   *int                   `json:"related_entity_id,omitempty" example:"456"`
	Data              map[string]interface{} `json:"data,omitempty"`
	Priority          string                 `json:"priority" example:"normal"`
	ScheduledFor      *time.Time             `json:"scheduled_for,omitempty"`
	ExpiresAt         *time.Time             `json:"expires_at,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	IsRead            bool                   `json:"is_read" example:"false"`
	TimeAgo           string                 `json:"time_ago" example:"2 hours ago"`
	IsExpired         bool                   `json:"is_expired" example:"false"`
}

// NotificationStatsResponse represents notification statistics
// @Description Statistics about user notifications
type NotificationStatsResponse struct {
	TotalCount     int  `json:"total_count" example:"50"`
	UnreadCount    int  `json:"unread_count" example:"5"`
	PendingCount   int  `json:"pending_count" example:"2"`
	SentCount      int  `json:"sent_count" example:"10"`
	DeliveredCount int  `json:"delivered_count" example:"33"`
	FailedCount    int  `json:"failed_count" example:"0"`
	HasUnread      bool `json:"has_unread" example:"true"`
}

// UnreadCountResponse represents the unread notification count
// @Description Response containing unread notification count
type UnreadCountResponse struct {
	UnreadCount int `json:"unread_count" example:"5"`
}

// MarkAllReadResponse represents the response for marking all as read
// @Description Response after marking all notifications as read
type MarkAllReadResponse struct {
	MarkedCount int `json:"marked_count" example:"5"`
}

// ========================================================================
// NOTIFICATION TYPES AND ENUMS (for documentation)
// ========================================================================

// NotificationTypeEnum documents valid notification types
// @Description Valid notification types
// booking_confirmation - Booking has been confirmed
// booking_reminder - Reminder about upcoming booking
// booking_cancelled - Booking has been cancelled
// booking_rescheduled - Booking has been rescheduled
// booking_completed - Booking service completed
// review_request - Request to review a completed booking
// review_response - Barber responded to your review
// payment_received - Payment has been received
// payment_failed - Payment failed
// account_welcome - Welcome to the platform
// account_verification - Verify your account
// password_reset - Password reset request
// promotion - Promotional notification
// system_alert - System alert
type NotificationTypeEnum string

// NotificationPriorityEnum documents valid priority levels
// @Description Valid notification priority levels
// low - Low priority
// normal - Normal priority (default)
// high - High priority
// urgent - Urgent priority
type NotificationPriorityEnum string

// NotificationStatusEnum documents valid notification statuses
// @Description Valid notification delivery statuses
// pending - Notification is queued
// sent - Notification has been sent
// delivered - Notification has been delivered
// read - Notification has been read
// failed - Notification delivery failed
type NotificationStatusEnum string

// ReviewModerationStatusEnum documents valid moderation statuses
// @Description Valid review moderation statuses
// pending - Awaiting moderation
// approved - Review approved and visible
// rejected - Review rejected
// flagged - Review flagged for further review
type ReviewModerationStatusEnum string

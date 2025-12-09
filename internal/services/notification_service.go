// internal/services/notification_service.go
package services

import (
	"context"
	"fmt"
	"time"

	"barber-booking-system/internal/config"
	"barber-booking-system/internal/models"
	"barber-booking-system/internal/repository"
)

// ========================================================================
// NOTIFICATION SERVICE - Business Logic Layer for Notifications
// ========================================================================

// NotificationService handles notification business logic
type NotificationService struct {
	repo        *repository.NotificationRepository
	userRepo    *repository.UserRepository
	bookingRepo *repository.BookingRepository
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	repo *repository.NotificationRepository,
	userRepo *repository.UserRepository,
	bookingRepo *repository.BookingRepository,
) *NotificationService {
	return &NotificationService{
		repo:        repo,
		userRepo:    userRepo,
		bookingRepo: bookingRepo,
	}
}

// ========================================================================
// REQUEST/RESPONSE STRUCTS
// ========================================================================

// CreateNotificationRequest represents a request to create a notification
type CreateNotificationRequest struct {
	UserID   int    `json:"user_id" binding:"required"`
	Title    string `json:"title" binding:"required,min=1,max=200"`
	Message  string `json:"message" binding:"required,min=1,max=2000"`
	Type     string `json:"type" binding:"required"`
	Priority string `json:"priority" binding:"omitempty,oneof=low normal high urgent"`

	// Optional fields
	Channels          []string               `json:"channels"`
	RelatedEntityType *string                `json:"related_entity_type"`
	RelatedEntityID   *int                   `json:"related_entity_id"`
	Data              map[string]interface{} `json:"data"`
	ScheduledFor      *time.Time             `json:"scheduled_for"`
	ExpiresAt         *time.Time             `json:"expires_at"`
}

// SendBookingNotificationRequest for booking-related notifications
type SendBookingNotificationRequest struct {
	BookingID        int    `json:"booking_id" binding:"required"`
	NotificationType string `json:"notification_type" binding:"required"`
	CustomMessage    string `json:"custom_message"`
}

// NotificationResponse wraps notification with computed fields
type NotificationResponse struct {
	*models.Notification
	IsRead    bool   `json:"is_read"`
	TimeAgo   string `json:"time_ago"`
	IsExpired bool   `json:"is_expired"`
}

// NotificationStatsResponse wraps stats with additional info
type NotificationStatsResponse struct {
	*repository.NotificationStats
	HasUnread bool `json:"has_unread"`
}

// ========================================================================
// HELPER FUNCTIONS
// ========================================================================

// toNotificationResponse converts a notification to a response
func (s *NotificationService) toNotificationResponse(notification *models.Notification) *NotificationResponse {
	response := &NotificationResponse{
		Notification: notification,
		IsRead:       notification.ReadAt != nil,
		IsExpired:    notification.ExpiresAt != nil && notification.ExpiresAt.Before(time.Now()),
	}

	// Calculate time ago
	duration := time.Since(notification.CreatedAt)
	switch {
	case duration < time.Minute:
		response.TimeAgo = "just now"
	case duration < time.Hour:
		response.TimeAgo = fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	case duration < 24*time.Hour:
		response.TimeAgo = fmt.Sprintf("%d hours ago", int(duration.Hours()))
	case duration < 7*24*time.Hour:
		response.TimeAgo = fmt.Sprintf("%d days ago", int(duration.Hours()/24))
	default:
		response.TimeAgo = notification.CreatedAt.Format("Jan 2, 2006")
	}

	return response
}

// getDefaultChannels returns default notification channels based on type
func getDefaultChannels(notifType string) []string {
	switch notifType {
	case "booking_confirmation", "booking_cancelled", "booking_rescheduled":
		return []string{"app", "email"}
	case "booking_reminder":
		return []string{"app", "push"}
	case "review_request":
		return []string{"app", "email"}
	case "payment_received", "payment_failed":
		return []string{"app", "email"}
	case "account_welcome", "account_verification", "password_reset":
		return []string{"email"}
	case "system_alert":
		return []string{"app"}
	default:
		return []string{"app"}
	}
}

// ========================================================================
// CREATE OPERATIONS
// ========================================================================

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(ctx context.Context, req CreateNotificationRequest) (*NotificationResponse, error) {
	// Validate notification type
	if !repository.IsValidNotificationType(req.Type) {
		return nil, repository.ErrInvalidNotificationType
	}

	// Set default channels if not provided
	channels := req.Channels
	if len(channels) == 0 {
		channels = getDefaultChannels(req.Type)
	}

	// Set default priority
	priority := req.Priority
	if priority == "" {
		priority = "normal"
	}

	// Build notification model
	notification := &models.Notification{
		UserID:   req.UserID,
		Title:    req.Title,
		Message:  req.Message,
		Type:     req.Type,
		Channels: channels,
		Status:   "pending",
		Priority: priority,
		Data:     req.Data,
	}

	// Set optional fields
	if req.RelatedEntityType != nil {
		notification.RelatedEntityType = req.RelatedEntityType
	}
	if req.RelatedEntityID != nil {
		notification.RelatedEntityID = req.RelatedEntityID
	}
	if req.ScheduledFor != nil {
		notification.ScheduledFor = req.ScheduledFor
	}
	if req.ExpiresAt != nil {
		notification.ExpiresAt = req.ExpiresAt
	}

	// Save notification
	if err := s.repo.Create(ctx, notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return s.toNotificationResponse(notification), nil
}

// ========================================================================
// BOOKING NOTIFICATION HELPERS
// ========================================================================

// SendBookingConfirmation sends a booking confirmation notification
func (s *NotificationService) SendBookingConfirmation(ctx context.Context, bookingID int) error {
	booking, err := s.bookingRepo.FindByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.CustomerID == nil {
		return nil // No notification for guest bookings without user
	}

	title := "Booking Confirmed"
	message := fmt.Sprintf("Your booking %s has been confirmed for %s",
		booking.BookingNumber,
		booking.ScheduledStartTime.Format("Monday, January 2 at 3:04 PM"))

	entityType := "booking"
	req := CreateNotificationRequest{
		UserID:            *booking.CustomerID,
		Title:             title,
		Message:           message,
		Type:              "booking_confirmation",
		Priority:          "normal",
		RelatedEntityType: &entityType,
		RelatedEntityID:   &bookingID,
		Data: map[string]interface{}{
			"booking_number": booking.BookingNumber,
			"barber_id":      booking.BarberID,
			"scheduled_time": booking.ScheduledStartTime,
		},
	}

	_, err = s.CreateNotification(ctx, req)
	return err
}

// SendBookingReminder sends a booking reminder notification
func (s *NotificationService) SendBookingReminder(ctx context.Context, bookingID int) error {
	booking, err := s.bookingRepo.FindByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.CustomerID == nil {
		return nil
	}

	title := "Upcoming Appointment Reminder"
	message := fmt.Sprintf("Reminder: Your appointment is scheduled for %s",
		booking.ScheduledStartTime.Format("Monday, January 2 at 3:04 PM"))

	entityType := "booking"
	req := CreateNotificationRequest{
		UserID:            *booking.CustomerID,
		Title:             title,
		Message:           message,
		Type:              "booking_reminder",
		Priority:          "high",
		RelatedEntityType: &entityType,
		RelatedEntityID:   &bookingID,
		Data: map[string]interface{}{
			"booking_number": booking.BookingNumber,
			"barber_id":      booking.BarberID,
		},
	}

	_, err = s.CreateNotification(ctx, req)
	return err
}

// SendBookingCancellation sends a booking cancellation notification
func (s *NotificationService) SendBookingCancellation(ctx context.Context, bookingID int, reason string) error {
	booking, err := s.bookingRepo.FindByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.CustomerID == nil {
		return nil
	}

	title := "Booking Cancelled"
	message := fmt.Sprintf("Your booking %s has been cancelled", booking.BookingNumber)
	if reason != "" {
		message += fmt.Sprintf(". Reason: %s", reason)
	}

	entityType := "booking"
	req := CreateNotificationRequest{
		UserID:            *booking.CustomerID,
		Title:             title,
		Message:           message,
		Type:              "booking_cancelled",
		Priority:          "high",
		RelatedEntityType: &entityType,
		RelatedEntityID:   &bookingID,
		Data: map[string]interface{}{
			"booking_number":      booking.BookingNumber,
			"cancellation_reason": reason,
		},
	}

	_, err = s.CreateNotification(ctx, req)
	return err
}

// SendBookingRescheduled sends a booking rescheduled notification
func (s *NotificationService) SendBookingRescheduled(ctx context.Context, bookingID int, oldTime, newTime time.Time) error {
	booking, err := s.bookingRepo.FindByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.CustomerID == nil {
		return nil
	}

	title := "Booking Rescheduled"
	message := fmt.Sprintf("Your booking %s has been rescheduled from %s to %s",
		booking.BookingNumber,
		oldTime.Format("Monday, January 2 at 3:04 PM"),
		newTime.Format("Monday, January 2 at 3:04 PM"))

	entityType := "booking"
	req := CreateNotificationRequest{
		UserID:            *booking.CustomerID,
		Title:             title,
		Message:           message,
		Type:              "booking_rescheduled",
		Priority:          "high",
		RelatedEntityType: &entityType,
		RelatedEntityID:   &bookingID,
		Data: map[string]interface{}{
			"booking_number": booking.BookingNumber,
			"old_time":       oldTime,
			"new_time":       newTime,
		},
	}

	_, err = s.CreateNotification(ctx, req)
	return err
}

// SendReviewRequest sends a request to review a completed booking
func (s *NotificationService) SendReviewRequest(ctx context.Context, bookingID int) error {
	booking, err := s.bookingRepo.FindByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.CustomerID == nil || booking.Status != config.BookingStatusCompleted {
		return nil
	}

	title := "How was your experience?"
	message := fmt.Sprintf("Please take a moment to review your recent appointment (%s). Your feedback helps us improve!", booking.ServiceName)

	// Set expiration for review request (e.g., 7 days)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	entityType := "booking"
	req := CreateNotificationRequest{
		UserID:            *booking.CustomerID,
		Title:             title,
		Message:           message,
		Type:              "review_request",
		Priority:          "normal",
		RelatedEntityType: &entityType,
		RelatedEntityID:   &bookingID,
		ExpiresAt:         &expiresAt,
		Data: map[string]interface{}{
			"booking_id":     bookingID,
			"booking_number": booking.BookingNumber,
			"service_name":   booking.ServiceName,
		},
	}

	_, err = s.CreateNotification(ctx, req)
	return err
}

// ========================================================================
// READ OPERATIONS
// ========================================================================

// GetNotificationByID retrieves a notification by ID
func (s *NotificationService) GetNotificationByID(ctx context.Context, id int, userID int) (*NotificationResponse, error) {
	notification, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify user owns the notification
	if notification.UserID != userID {
		return nil, fmt.Errorf("notification not found")
	}

	return s.toNotificationResponse(notification), nil
}

// GetUserNotifications retrieves all notifications for a user
func (s *NotificationService) GetUserNotifications(ctx context.Context, userID int, filters repository.NotificationFilters) ([]NotificationResponse, error) {
	notifications, err := s.repo.FindByUserID(ctx, userID, filters)
	if err != nil {
		return nil, err
	}

	responses := make([]NotificationResponse, len(notifications))
	for i, n := range notifications {
		responses[i] = *s.toNotificationResponse(&n)
	}
	return responses, nil
}

// GetUnreadNotifications retrieves unread notifications for a user
func (s *NotificationService) GetUnreadNotifications(ctx context.Context, userID int, limit int) ([]NotificationResponse, error) {
	if limit <= 0 {
		limit = 20
	}

	notifications, err := s.repo.GetUnreadNotifications(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	responses := make([]NotificationResponse, len(notifications))
	for i, n := range notifications {
		responses[i] = *s.toNotificationResponse(&n)
	}
	return responses, nil
}

// GetNotificationStats retrieves notification statistics for a user
func (s *NotificationService) GetNotificationStats(ctx context.Context, userID int) (*NotificationStatsResponse, error) {
	stats, err := s.repo.GetUserStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &NotificationStatsResponse{
		NotificationStats: stats,
		HasUnread:         stats.UnreadCount > 0,
	}, nil
}

// GetUnreadCount returns the unread notification count
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID int) (int, error) {
	return s.repo.GetUnreadCount(ctx, userID)
}

// ========================================================================
// UPDATE OPERATIONS
// ========================================================================

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, id int, userID int) error {
	// Verify user owns the notification
	notification, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if notification.UserID != userID {
		return fmt.Errorf("notification not found")
	}

	return s.repo.MarkAsRead(ctx, id)
}

// MarkAllAsRead marks all notifications for a user as read
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID int) (int, error) {
	return s.repo.MarkAllAsRead(ctx, userID)
}

// MarkAsDelivered marks a notification as delivered (for push notification callbacks)
func (s *NotificationService) MarkAsDelivered(ctx context.Context, id int) error {
	return s.repo.MarkAsDelivered(ctx, id)
}

// ========================================================================
// DELETE OPERATIONS
// ========================================================================

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(ctx context.Context, id int, userID int) error {
	// Verify user owns the notification
	notification, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if notification.UserID != userID {
		return fmt.Errorf("notification not found")
	}

	return s.repo.Delete(ctx, id)
}

// ========================================================================
// PROCESSING OPERATIONS (for background workers)
// ========================================================================

// GetPendingNotifications retrieves notifications ready to be sent
func (s *NotificationService) GetPendingNotifications(ctx context.Context, limit int) ([]models.Notification, error) {
	return s.repo.GetPendingNotifications(ctx, limit)
}

// ProcessNotification processes a single notification (mark as sent)
func (s *NotificationService) ProcessNotification(ctx context.Context, id int) error {
	return s.repo.MarkAsSent(ctx, id)
}

// MarkNotificationFailed marks a notification as failed
func (s *NotificationService) MarkNotificationFailed(ctx context.Context, id int, errorMsg string) error {
	return s.repo.MarkAsFailed(ctx, id, errorMsg)
}

// CleanupOldNotifications removes old read notifications
func (s *NotificationService) CleanupOldNotifications(ctx context.Context, olderThan time.Duration) (int, error) {
	return s.repo.DeleteOldNotifications(ctx, olderThan)
}

// CleanupExpiredNotifications removes expired undelivered notifications
func (s *NotificationService) CleanupExpiredNotifications(ctx context.Context) (int, error) {
	return s.repo.DeleteExpiredNotifications(ctx)
}

// ========================================================================
// BATCH OPERATIONS
// ========================================================================

// SendBulkNotification sends the same notification to multiple users
func (s *NotificationService) SendBulkNotification(ctx context.Context, userIDs []int, title, message, notifType string) error {
	notifications := make([]*models.Notification, len(userIDs))

	for i, userID := range userIDs {
		notifications[i] = &models.Notification{
			UserID:   userID,
			Title:    title,
			Message:  message,
			Type:     notifType,
			Channels: getDefaultChannels(notifType),
			Status:   "pending",
			Priority: "normal",
		}
	}

	return s.repo.CreateBatch(ctx, notifications)
}

// ScheduleBookingReminders schedules reminder notifications for upcoming bookings
func (s *NotificationService) ScheduleBookingReminders(ctx context.Context, hoursBeforeBooking int) error {
	// Get upcoming bookings within the reminder window
	reminderTime := time.Now().Add(time.Duration(hoursBeforeBooking) * time.Hour)

	filters := repository.BookingFilters{
		StartDateFrom: time.Now(),
		StartDateTo:   reminderTime,
		Statuses:      []string{config.BookingStatusConfirmed},
	}

	bookings, err := s.bookingRepo.FindAll(ctx, filters)
	if err != nil {
		return err
	}

	for _, booking := range bookings {
		// Check if reminder already sent
		existingNotifs, err := s.repo.GetByRelatedEntity(ctx, "booking", booking.ID)
		if err != nil {
			continue
		}

		reminderExists := false
		for _, n := range existingNotifs {
			if n.Type == "booking_reminder" {
				reminderExists = true
				break
			}
		}

		if !reminderExists {
			_ = s.SendBookingReminder(ctx, booking.ID)
		}
	}

	return nil
}

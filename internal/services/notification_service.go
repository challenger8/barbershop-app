// internal/services/notification_service.go
package services

import (
	"context"
	"fmt"
	"time"

	"barber-booking-system/internal/config"
	"barber-booking-system/internal/logger"
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
	case config.NotificationTypeBookingConfirmation, config.NotificationTypeBookingCancelled, config.NotificationTypeBookingRescheduled:
		return []string{config.NotificationChannelApp, config.NotificationChannelEmail}
	case config.NotificationTypeBookingReminder:
		return []string{config.NotificationChannelApp, config.NotificationChannelPush}
	case config.NotificationTypeReviewRequest:
		return []string{config.NotificationChannelApp, config.NotificationChannelEmail}
	case config.NotificationTypePaymentReceived, config.NotificationTypePaymentFailed:
		return []string{config.NotificationChannelApp, config.NotificationChannelEmail}
	case config.NotificationTypeAccountWelcome, config.NotificationTypeAccountVerification, config.NotificationTypePasswordReset:
		return []string{config.NotificationChannelEmail}
	case config.NotificationTypeSystemAlert:
		return []string{config.NotificationChannelApp}
	default:
		return []string{config.NotificationChannelApp}
	}
}

// ========================================================================
// CREATE OPERATIONS
// ========================================================================

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(ctx context.Context, req CreateNotificationRequest) (*NotificationResponse, error) {
	log := logger.FromContext(ctx)

	log.Debug("Creating notification").
		Int("user_id", req.UserID).
		Str("type", req.Type).
		Str("priority", req.Priority).
		Send()

	// Validate notification type
	if !repository.IsValidNotificationType(req.Type) {
		log.Warn("Invalid notification type").
			Str("type", req.Type).
			Send()
		return nil, repository.ErrInvalidNotificationType
	}

	// Set default priority
	if req.Priority == "" {
		req.Priority = config.NotificationPriorityNormal
	}

	// Set default channels
	if len(req.Channels) == 0 {
		req.Channels = []string{config.NotificationChannelApp}
	}

	// Build notification model
	notification := &models.Notification{
		UserID:            req.UserID,
		Title:             req.Title,
		Message:           req.Message,
		Type:              req.Type,
		Priority:          req.Priority,
		Channels:          req.Channels,
		RelatedEntityType: req.RelatedEntityType,
		RelatedEntityID:   req.RelatedEntityID,
		Data:              req.Data,
		ScheduledFor:      req.ScheduledFor,
		ExpiresAt:         req.ExpiresAt,
		Status:            config.NotificationStatusPending,
	}

	// Create in database
	if err := s.repo.Create(ctx, notification); err != nil {
		log.Error(err).
			Int("user_id", req.UserID).
			Str("type", req.Type).
			Msg("Failed to create notification")
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	log.Info("Notification created successfully").
		Int("notification_id", notification.ID).
		Int("user_id", req.UserID).
		Str("type", req.Type).
		Send()

	return s.toNotificationResponse(notification), nil
}

// ========================================================================
// BOOKING NOTIFICATION HELPERS
// ========================================================================
// SendBookingReminder sends a booking reminder notification
func (s *NotificationService) SendBookingReminder(ctx context.Context, bookingID int) error {
	booking, err := s.bookingRepo.FindByID(ctx, bookingID)
	if err != nil {
		return err
	}

	return s.sendBookingNotificationWithTemplate(
		ctx, booking, "reminder",
		[]interface{}{booking.ScheduledStartTime.Format("Monday, January 2 at 3:04 PM")},
		nil,
		nil,
	)
}

// SendBookingCancellation sends a booking cancellation notification
func (s *NotificationService) SendBookingCancellation(ctx context.Context, bookingID int, reason string) error {
	log := logger.FromContext(ctx)

	log.Debug("Sending booking cancellation notification").
		Int("booking_id", bookingID).
		Send()

	booking, err := s.bookingRepo.FindByID(ctx, bookingID)
	if err != nil {
		log.Warn("Booking not found for cancellation notification").
			Int("booking_id", bookingID).
			Err(err).
			Send()
		return err
	}

	if booking.CustomerID == nil {
		log.Debug("Skipping cancellation notification - no customer ID").
			Int("booking_id", bookingID).
			Send()
		return nil
	}

	title := "Booking Cancelled"
	message := fmt.Sprintf("Your booking %s has been cancelled", booking.BookingNumber)
	if reason != "" {
		message += fmt.Sprintf(". Reason: %s", reason)
	}

	entityType := config.EntityTypeBooking
	req := CreateNotificationRequest{
		UserID:            *booking.CustomerID,
		Title:             title,
		Message:           message,
		Type:              config.NotificationTypeBookingCancelled,
		Priority:          config.NotificationPriorityHigh,
		RelatedEntityType: &entityType,
		RelatedEntityID:   &bookingID,
		Data: map[string]interface{}{
			"booking_number":      booking.BookingNumber,
			"cancellation_reason": reason,
		},
	}

	_, err = s.CreateNotification(ctx, req)
	if err != nil {
		log.Error(err).
			Int("booking_id", bookingID).
			Msg("Failed to send booking cancellation notification")
		return err
	}

	log.Info("Booking cancellation notification sent").
		Int("booking_id", bookingID).
		Str("booking_number", booking.BookingNumber).
		Int("customer_id", *booking.CustomerID).
		Send()

	return nil
}

// SendBookingRescheduled sends a booking rescheduled notification
func (s *NotificationService) SendBookingRescheduled(ctx context.Context, bookingID int, oldTime, newTime time.Time) error {
	booking, err := s.bookingRepo.FindByID(ctx, bookingID)
	if err != nil {
		return err
	}

	return s.sendBookingNotificationWithTemplate(
		ctx, booking, "rescheduled",
		[]interface{}{
			booking.BookingNumber,
			oldTime.Format("Monday, January 2 at 3:04 PM"),
			newTime.Format("Monday, January 2 at 3:04 PM"),
		},
		map[string]interface{}{"old_time": oldTime, "new_time": newTime},
		nil,
	)
}

// SendReviewRequest sends a request to review a completed booking
func (s *NotificationService) SendReviewRequest(ctx context.Context, bookingID int) error {
	booking, err := s.bookingRepo.FindByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.Status != config.BookingStatusCompleted {
		return nil
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	return s.sendBookingNotificationWithTemplate(
		ctx, booking, "review_request",
		[]interface{}{booking.ServiceName},
		map[string]interface{}{
			"booking_id":   bookingID,
			"service_name": booking.ServiceName,
		},
		&expiresAt,
	)
}

// SendBookingConfirmation sends a booking confirmation notification
func (s *NotificationService) SendBookingConfirmation(ctx context.Context, bookingID int) error {
	log := logger.FromContext(ctx)

	log.Debug("Sending booking confirmation").
		Int("booking_id", bookingID).
		Send()

	booking, err := s.bookingRepo.FindByID(ctx, bookingID)
	if err != nil {
		log.Warn("Booking not found for confirmation").
			Int("booking_id", bookingID).
			Err(err).
			Send()
		return err
	}

	if booking.CustomerID == nil {
		log.Debug("Skipping confirmation - no customer ID").
			Int("booking_id", bookingID).
			Send()
		return nil
	}

	title := "Booking Confirmed"
	message := fmt.Sprintf("Your booking %s has been confirmed for %s",
		booking.BookingNumber,
		booking.ScheduledStartTime.Format("Monday, January 2 at 3:04 PM"))

	entityType := config.EntityTypeBooking
	req := CreateNotificationRequest{
		UserID:            *booking.CustomerID,
		Title:             title,
		Message:           message,
		Type:              config.NotificationTypeBookingConfirmation,
		Priority:          config.NotificationPriorityNormal,
		RelatedEntityType: &entityType,
		RelatedEntityID:   &bookingID,
		Data: map[string]interface{}{
			"booking_number": booking.BookingNumber,
			"barber_id":      booking.BarberID,
			"scheduled_time": booking.ScheduledStartTime,
		},
	}

	_, err = s.CreateNotification(ctx, req)
	if err != nil {
		log.Error(err).
			Int("booking_id", bookingID).
			Msg("Failed to send booking confirmation")
		return err
	}

	log.Info("Booking confirmation sent").
		Int("booking_id", bookingID).
		Str("booking_number", booking.BookingNumber).
		Int("customer_id", *booking.CustomerID).
		Send()

	return nil
}

// ========================================================================
// NOTIFICATION TEMPLATES (DRY - extracted from repeated patterns)
// ========================================================================

// BookingNotificationTemplate defines a notification template for booking events
type BookingNotificationTemplate struct {
	Title           string
	MessageTemplate string // Uses %s placeholders
	Type            string
	Priority        string
}

// bookingNotificationTemplates maps notification types to their templates
var bookingNotificationTemplates = map[string]BookingNotificationTemplate{
	"confirmation": {
		Title:           "Booking Confirmed",
		MessageTemplate: "Your booking %s has been confirmed for %s",
		Type:            config.NotificationTypeBookingConfirmation,
		Priority:        config.NotificationPriorityNormal,
	},
	"reminder": {
		Title:           "Upcoming Appointment Reminder",
		MessageTemplate: "Reminder: Your appointment is scheduled for %s",
		Type:            config.NotificationTypeBookingReminder,
		Priority:        config.NotificationPriorityHigh,
	},
	"cancellation": {
		Title:           "Booking Cancelled",
		MessageTemplate: "Your booking %s has been cancelled",
		Type:            config.NotificationTypeBookingCancelled,
		Priority:        config.NotificationPriorityHigh,
	},
	"rescheduled": {
		Title:           "Booking Rescheduled",
		MessageTemplate: "Your booking %s has been rescheduled from %s to %s",
		Type:            config.NotificationTypeBookingRescheduled,
		Priority:        config.NotificationPriorityHigh,
	},
	"review_request": {
		Title:           "How was your experience?",
		MessageTemplate: "Please take a moment to review your recent appointment (%s). Your feedback helps us improve!",
		Type:            config.NotificationTypeReviewRequest,
		Priority:        config.NotificationPriorityNormal,
	},
}

// sendBookingNotificationWithTemplate is a helper that sends booking notifications using templates
func (s *NotificationService) sendBookingNotificationWithTemplate(
	ctx context.Context,
	booking *models.Booking,
	templateKey string,
	messageArgs []interface{},
	extraData map[string]interface{},
	expiresAt *time.Time,
) error {
	if booking.CustomerID == nil {
		return nil // No notification for guest bookings
	}

	template, exists := bookingNotificationTemplates[templateKey]
	if !exists {
		return fmt.Errorf("unknown notification template: %s", templateKey)
	}

	message := fmt.Sprintf(template.MessageTemplate, messageArgs...)

	entityType := config.EntityTypeBooking
	data := map[string]interface{}{
		"booking_number": booking.BookingNumber,
		"barber_id":      booking.BarberID,
	}
	// Merge extra data
	for k, v := range extraData {
		data[k] = v
	}

	req := CreateNotificationRequest{
		UserID:            *booking.CustomerID,
		Title:             template.Title,
		Message:           message,
		Type:              template.Type,
		Priority:          template.Priority,
		RelatedEntityType: &entityType,
		RelatedEntityID:   &booking.ID,
		Data:              data,
		ExpiresAt:         expiresAt,
	}

	_, err := s.CreateNotification(ctx, req)
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
			Status:   config.NotificationStatusPending,
			Priority: config.NotificationPriorityNormal,
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
		existingNotifs, err := s.repo.GetByRelatedEntity(ctx, config.EntityTypeBooking, booking.ID)
		if err != nil {
			continue
		}

		reminderExists := false
		for _, n := range existingNotifs {
			if n.Type == config.NotificationTypeBookingReminder {
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

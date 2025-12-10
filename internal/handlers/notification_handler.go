// internal/handlers/notification_handler.go
package handlers

import (
	"net/http"

	"barber-booking-system/internal/config"
	"barber-booking-system/internal/middleware"
	"barber-booking-system/internal/repository"
	"barber-booking-system/internal/services"
	"barber-booking-system/internal/utils"

	"github.com/gin-gonic/gin"
)

// ========================================================================
// NOTIFICATION HANDLER - HTTP Request Handlers for Notifications
// ========================================================================

// NotificationHandler handles notification-related HTTP requests
type NotificationHandler struct {
	notificationService *services.NotificationService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notificationService *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// ========================================================================
// HELPER FUNCTIONS
// ========================================================================

// buildNotificationFilters builds NotificationFilters from query parameters
func buildNotificationFilters(c *gin.Context) repository.NotificationFilters {
	filters := repository.NotificationFilters{
		Type:              c.Query("type"),
		Status:            c.Query("status"),
		Priority:          c.Query("priority"),
		Channel:           c.Query("channel"),
		RelatedEntityType: c.Query("related_entity_type"),
		RelatedEntityID:   ParseIntQuery(c, "related_entity_id", 0),
		Search:            c.Query("search"),
		SortBy:            c.Query("sort_by"),
		Order:             c.Query("order"),
		Limit:             ParseIntQuery(c, "limit", 50),
		Offset:            ParseIntQuery(c, "offset", 0),
		CreatedFrom:       ParseTimeQuery(c, "created_from"),
		CreatedTo:         ParseTimeQuery(c, "created_to"),
	}

	// Boolean filters
	if isRead := ParseBoolQuery(c, "is_read"); isRead != nil {
		filters.IsRead = isRead
	}
	if isUnread := ParseBoolQuery(c, "is_unread"); isUnread != nil {
		filters.IsUnread = isUnread
	}
	if includeExpired := ParseBoolQuery(c, "include_expired"); includeExpired != nil {
		filters.IncludeExpired = *includeExpired
	}

	return filters
}

// ========================================================================
// GET MY NOTIFICATIONS
// ========================================================================

// GetMyNotifications godoc
// @Summary Get my notifications
// @Description Get all notifications for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Param type query string false "Filter by notification type"
// @Param status query string false "Filter by status"
// @Param priority query string false "Filter by priority"
// @Param is_read query bool false "Filter by read status"
// @Param is_unread query bool false "Filter by unread status"
// @Param sort_by query string false "Sort by field" default(created_at)
// @Param order query string false "Sort order (ASC/DESC)" default(DESC)
// @Param limit query int false "Limit results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/notifications [get]
func (h *NotificationHandler) GetMyNotifications(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		RespondUnauthorized(c, "You must be logged in to view notifications")
		return
	}

	filters := buildNotificationFilters(c)

	notifications, err := h.notificationService.GetUserNotifications(c.Request.Context(), userID, filters)
	if err != nil {
		RespondInternalError(c, "fetch notifications", err)
		return
	}

	RespondSuccessWithMeta(c, notifications, PaginationMeta(len(notifications), filters.Limit, filters.Offset))
}

// GetUnreadNotifications godoc
// @Summary Get unread notifications
// @Description Get unread notifications for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Param limit query int false "Limit results" default(20)
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/notifications/unread [get]
func (h *NotificationHandler) GetUnreadNotifications(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		RespondUnauthorized(c, "You must be logged in to view notifications")
		return
	}

	limit := ParseIntQuery(c, "limit", 20)

	notifications, err := h.notificationService.GetUnreadNotifications(c.Request.Context(), userID, limit)
	if err != nil {
		RespondInternalError(c, "fetch unread notifications", err)
		return
	}

	RespondSuccessWithMeta(c, notifications, map[string]interface{}{
		"count": len(notifications),
	})
}

// ========================================================================
// GET NOTIFICATION BY ID
// ========================================================================

// GetNotification godoc
// @Summary Get notification by ID
// @Description Get detailed information about a specific notification
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path int true "Notification ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/notifications/{id} [get]
func (h *NotificationHandler) GetNotification(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "notification")
	if !ok {
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		RespondUnauthorized(c, "You must be logged in to view notifications")
		return
	}

	notification, err := h.notificationService.GetNotificationByID(c.Request.Context(), id, userID)
	if err != nil {
		if err == repository.ErrNotificationNotFound || utils.ContainsAny(err.Error(), []string{"not found"}) {
			RespondNotFound(c, "Notification")
			return
		}
		RespondInternalError(c, "fetch notification", err)
		return
	}

	RespondSuccess(c, notification)
}

// ========================================================================
// GET NOTIFICATION STATS
// ========================================================================

// GetNotificationStats godoc
// @Summary Get notification statistics
// @Description Get notification statistics for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/notifications/stats [get]
func (h *NotificationHandler) GetNotificationStats(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		RespondUnauthorized(c, "You must be logged in to view notification stats")
		return
	}

	stats, err := h.notificationService.GetNotificationStats(c.Request.Context(), userID)
	if err != nil {
		RespondInternalError(c, "fetch notification stats", err)
		return
	}

	RespondSuccess(c, stats)
}

// GetUnreadCount godoc
// @Summary Get unread notification count
// @Description Get the count of unread notifications for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/notifications/unread/count [get]
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		RespondUnauthorized(c, "You must be logged in to view notification count")
		return
	}

	count, err := h.notificationService.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		RespondInternalError(c, "fetch unread count", err)
		return
	}

	RespondSuccess(c, map[string]interface{}{
		"unread_count": count,
	})
}

// ========================================================================
// MARK AS READ
// ========================================================================

// MarkAsRead godoc
// @Summary Mark notification as read
// @Description Mark a specific notification as read
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path int true "Notification ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/notifications/{id}/read [patch]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "notification")
	if !ok {
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		RespondUnauthorized(c, "You must be logged in to mark notifications as read")
		return
	}

	err := h.notificationService.MarkAsRead(c.Request.Context(), id, userID)
	if err != nil {
		if err == repository.ErrNotificationNotFound || utils.ContainsAny(err.Error(), []string{"not found"}) {
			RespondNotFound(c, "Notification")
			return
		}
		RespondInternalError(c, "mark notification as read", err)
		return
	}

	RespondSuccessWithMessage(c, "Notification marked as read")
}

// MarkAllAsRead godoc
// @Summary Mark all notifications as read
// @Description Mark all notifications for the authenticated user as read
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/notifications/read-all [patch]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		RespondUnauthorized(c, "You must be logged in to mark notifications as read")
		return
	}

	count, err := h.notificationService.MarkAllAsRead(c.Request.Context(), userID)
	if err != nil {
		RespondInternalError(c, "mark all notifications as read", err)
		return
	}

	RespondSuccessWithData(c, map[string]interface{}{
		"marked_count": count,
	}, "All notifications marked as read")
}

// ========================================================================
// DELETE NOTIFICATION
// ========================================================================

// DeleteNotification godoc
// @Summary Delete a notification
// @Description Delete a specific notification
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path int true "Notification ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/notifications/{id} [delete]
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "notification")
	if !ok {
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		RespondUnauthorized(c, "You must be logged in to delete notifications")
		return
	}

	err := h.notificationService.DeleteNotification(c.Request.Context(), id, userID)
	if err != nil {
		if err == repository.ErrNotificationNotFound || utils.ContainsAny(err.Error(), []string{"not found"}) {
			RespondNotFound(c, "Notification")
			return
		}
		RespondInternalError(c, "delete notification", err)
		return
	}

	RespondSuccessWithMessage(c, "Notification deleted successfully")
}

// ========================================================================
// CREATE NOTIFICATION (Admin only)
// ========================================================================

// CreateNotification godoc
// @Summary Create a notification
// @Description Create a new notification for a user (admin only)
// @Tags notifications
// @Accept json
// @Produce json
// @Param notification body services.CreateNotificationRequest true "Notification data"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 403 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/notifications [post]
func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	req, ok := BindJSON[services.CreateNotificationRequest](c)
	if !ok {
		return
	}

	// Note: In a real app, verify user is admin
	// userType, _ := middleware.GetUserType(c)
	// if userType != "admin" { ... }

	notification, err := h.notificationService.CreateNotification(c.Request.Context(), *req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == repository.ErrInvalidNotificationType || err == repository.ErrInvalidNotificationStatus {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, middleware.ErrorResponse{
			Error:   "Failed to create notification",
			Message: err.Error(),
		})
		return
	}

	RespondCreated(c, notification, "Notification created successfully")
}

// ========================================================================
// SEND BOOKING NOTIFICATION (Admin/System only)
// ========================================================================

// SendBookingNotification godoc
// @Summary Send booking notification
// @Description Send a notification related to a booking
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body services.SendBookingNotificationRequest true "Booking notification request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/notifications/booking [post]
func (h *NotificationHandler) SendBookingNotification(c *gin.Context) {
	req, ok := BindJSON[services.SendBookingNotificationRequest](c)
	if !ok {
		return
	}

	var err error
	ctx := c.Request.Context()

	switch req.NotificationType {
	case "booking_confirmation":
		err = h.notificationService.SendBookingConfirmation(ctx, req.BookingID)
	case "booking_reminder":
		err = h.notificationService.SendBookingReminder(ctx, req.BookingID)
	case "booking_cancelled":
		err = h.notificationService.SendBookingCancellation(ctx, req.BookingID, req.CustomMessage)
	case "review_request":
		err = h.notificationService.SendReviewRequest(ctx, req.BookingID)
	default:
		RespondBadRequest(c, "Invalid notification type", "Supported types: booking_confirmation, booking_reminder, booking_cancelled, review_request")
		return
	}

	if err != nil {
		if utils.ContainsAny(err.Error(), []string{"not found"}) {
			RespondNotFound(c, "Booking")
			return
		}
		RespondInternalError(c, "send booking notification", err)
		return
	}

	RespondSuccessWithMessage(c, "Notification sent successfully")
}

// ========================================================================
// WEBHOOK CALLBACK (for push notification services)
// ========================================================================

// DeliveryWebhook godoc
// @Summary Notification delivery webhook
// @Description Webhook endpoint for push notification delivery callbacks
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path int true "Notification ID"
// @Param status query string true "Delivery status (delivered/failed)"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/notifications/{id}/webhook [post]
func (h *NotificationHandler) DeliveryWebhook(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "notification")
	if !ok {
		return
	}

	status := c.Query("status")
	if status == "" {
		RespondBadRequest(c, "Missing status", "status query parameter is required")
		return
	}

	var err error
	ctx := c.Request.Context()

	switch status {
	case config.NotificationStatusDelivered:
		err = h.notificationService.MarkAsDelivered(ctx, id)
	case config.NotificationStatusFailed:
		errorMsg := c.Query("error")
		err = h.notificationService.MarkNotificationFailed(ctx, id, errorMsg)
	default:
		RespondBadRequest(c, "Invalid status", "status must be 'delivered' or 'failed'")
		return
	}

	if err != nil {
		if err == repository.ErrNotificationNotFound {
			RespondNotFound(c, "Notification")
			return
		}
		RespondInternalError(c, "process webhook", err)
		return
	}

	RespondSuccessWithMessage(c, "Webhook processed successfully")
}

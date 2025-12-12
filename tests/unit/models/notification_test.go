// tests/unit/models/notification_test.go
package models

import (
	"testing"
	"time"

	"barber-booking-system/internal/config"
	"barber-booking-system/internal/models"
)

// ========================================================================
// NOTIFICATION MODEL UNIT TESTS
// ========================================================================




func TestNotification_WithScheduling(t *testing.T) {
	scheduledFor := time.Now().Add(24 * time.Hour)
	expiresAt := time.Now().Add(48 * time.Hour)

	notification := &models.Notification{
		ID:           1,
		UserID:       100,
		Title:        "Reminder",
		Message:      "Don't forget your appointment",
		Type:         "booking_reminder",
		Status:       "pending",
		ScheduledFor: &scheduledFor,
		ExpiresAt:    &expiresAt,
	}

	if notification.ScheduledFor == nil {
		t.Fatal("Expected ScheduledFor to be set")
	}
	if notification.ExpiresAt == nil {
		t.Fatal("Expected ExpiresAt to be set")
	}
	if notification.ExpiresAt.Before(*notification.ScheduledFor) {
		t.Error("ExpiresAt should be after ScheduledFor")
	}
}

func TestNotification_DeliveryTracking(t *testing.T) {
	now := time.Now()
	sentAt := now.Add(-1 * time.Hour)
	deliveredAt := now.Add(-30 * time.Minute)
	readAt := now

	notification := &models.Notification{
		ID:          1,
		UserID:      100,
		Title:       "Read Notification",
		Message:     "This notification has been read",
		Type:        "system_alert",
		Status:      "read",
		SentAt:      &sentAt,
		DeliveredAt: &deliveredAt,
		ReadAt:      &readAt,
	}

	if notification.SentAt == nil {
		t.Fatal("Expected SentAt to be set")
	}
	if notification.DeliveredAt == nil {
		t.Fatal("Expected DeliveredAt to be set")
	}
	if notification.ReadAt == nil {
		t.Fatal("Expected ReadAt to be set")
	}

	// Verify chronological order
	if notification.DeliveredAt.Before(*notification.SentAt) {
		t.Error("DeliveredAt should be after SentAt")
	}
	if notification.ReadAt.Before(*notification.DeliveredAt) {
		t.Error("ReadAt should be after DeliveredAt")
	}
}





func TestNotification_PriorityLevels(t *testing.T) {
	priorities := []string{
		config.NotificationPriorityLow,
		config.NotificationPriorityNormal,
		config.NotificationPriorityHigh,
		config.NotificationPriorityUrgent,
	}

	for _, priority := range priorities {
		notification := &models.Notification{
			ID:       1,
			UserID:   100,
			Title:    "Priority Test",
			Message:  "Testing priority levels",
			Type:     "system_alert",
			Status:   "pending",
			Priority: priority,
		}

		if notification.Priority != priority {
			t.Errorf("Expected priority %s, got %s", priority, notification.Priority)
		}
	}
}

func TestNotification_StatusTransitions(t *testing.T) {
	statuses := []string{
		config.NotificationStatusPending,
		config.NotificationStatusSent,
		config.NotificationStatusDelivered,
		config.NotificationStatusRead,
		config.NotificationStatusFailed,
	}

	for _, status := range statuses {
		notification := &models.Notification{
			ID:      1,
			UserID:  100,
			Title:   "Status Test",
			Message: "Testing status values",
			Type:    "system_alert",
			Status:  status,
		}

		if notification.Status != status {
			t.Errorf("Expected status %s, got %s", status, notification.Status)
		}
	}
}

func TestNotification_Types(t *testing.T) {
	notificationTypes := []string{
		config.NotificationTypeBookingConfirmation,
		config.NotificationTypeBookingReminder,
		config.NotificationTypeBookingCancelled,
		config.NotificationTypeBookingRescheduled,
		config.NotificationTypeBookingCompleted,
		config.NotificationTypeReviewRequest,
		config.NotificationTypeReviewResponse,
		config.NotificationTypePaymentReceived,
		config.NotificationTypePaymentFailed,
		config.NotificationTypeAccountWelcome,
		config.NotificationTypeAccountVerification,
		config.NotificationTypePasswordReset,
		config.NotificationTypePromotion,
		config.NotificationTypeSystemAlert,
	}

	for _, notifType := range notificationTypes {
		notification := &models.Notification{
			ID:      1,
			UserID:  100,
			Title:   "Type Test",
			Message: "Testing notification types",
			Type:    notifType,
			Status:  "pending",
		}

		if notification.Type != notifType {
			t.Errorf("Expected type %s, got %s", notifType, notification.Type)
		}
	}
}

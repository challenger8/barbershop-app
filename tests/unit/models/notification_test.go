// tests/unit/models/notification_test.go
package models

import (
	"testing"
	"time"

	"barber-booking-system/internal/models"
)

// ========================================================================
// NOTIFICATION MODEL UNIT TESTS
// ========================================================================

func TestNotification_Creation(t *testing.T) {
	notification := &models.Notification{
		ID:       1,
		UserID:   100,
		Title:    "Test Notification",
		Message:  "This is a test message",
		Type:     "booking_confirmation",
		Status:   "pending",
		Priority: "normal",
	}

	if notification.UserID != 100 {
		t.Errorf("Expected UserID 100, got %d", notification.UserID)
	}
	if notification.Type != "booking_confirmation" {
		t.Errorf("Expected Type booking_confirmation, got %s", notification.Type)
	}
	if notification.Status != "pending" {
		t.Errorf("Expected Status pending, got %s", notification.Status)
	}
}

func TestNotification_WithRelatedEntity(t *testing.T) {
	entityType := "booking"
	entityID := 123

	notification := &models.Notification{
		ID:                1,
		UserID:            100,
		Title:             "Booking Confirmed",
		Message:           "Your booking has been confirmed",
		Type:              "booking_confirmation",
		Status:            "pending",
		RelatedEntityType: &entityType,
		RelatedEntityID:   &entityID,
	}

	if notification.RelatedEntityType == nil {
		t.Fatal("Expected RelatedEntityType to be set")
	}
	if *notification.RelatedEntityType != "booking" {
		t.Errorf("Expected RelatedEntityType booking, got %s", *notification.RelatedEntityType)
	}
	if *notification.RelatedEntityID != 123 {
		t.Errorf("Expected RelatedEntityID 123, got %d", *notification.RelatedEntityID)
	}
}

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

func TestNotification_Channels(t *testing.T) {
	notification := &models.Notification{
		ID:       1,
		UserID:   100,
		Title:    "Multi-channel Notification",
		Message:  "Sent via multiple channels",
		Type:     "booking_confirmation",
		Status:   "pending",
		Channels: models.StringArray{"app", "email", "sms"},
	}

	if len(notification.Channels) != 3 {
		t.Errorf("Expected 3 channels, got %d", len(notification.Channels))
	}

	expectedChannels := map[string]bool{"app": true, "email": true, "sms": true}
	for _, channel := range notification.Channels {
		if !expectedChannels[channel] {
			t.Errorf("Unexpected channel: %s", channel)
		}
	}
}

func TestNotification_DataField(t *testing.T) {
	notification := &models.Notification{
		ID:      1,
		UserID:  100,
		Title:   "Data Notification",
		Message: "Contains extra data",
		Type:    "booking_confirmation",
		Status:  "pending",
		Data: models.JSONMap{
			"booking_id":   123,
			"barber_name":  "John Doe",
			"service_name": "Haircut",
		},
	}

	if notification.Data == nil {
		t.Fatal("Expected Data to be set")
	}
	if notification.Data["booking_id"] != 123 {
		t.Errorf("Expected booking_id 123, got %v", notification.Data["booking_id"])
	}
}

func TestNotification_PriorityLevels(t *testing.T) {
	priorities := []string{"low", "normal", "high", "urgent"}

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
	statuses := []string{"pending", "sent", "delivered", "read", "failed"}

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
		"booking_confirmation",
		"booking_reminder",
		"booking_cancelled",
		"booking_rescheduled",
		"booking_completed",
		"review_request",
		"review_response",
		"payment_received",
		"payment_failed",
		"account_welcome",
		"account_verification",
		"password_reset",
		"promotion",
		"system_alert",
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

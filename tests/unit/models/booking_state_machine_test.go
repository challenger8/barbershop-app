// tests/unit/models/booking_state_machine_test.go
package models

import (
	"strings"
	"testing"

	"barber-booking-system/internal/config"
	"barber-booking-system/internal/models"
)

// ========================================================================
// STATE MACHINE UNIT TESTS
// ========================================================================

func TestNewBookingStateMachine(t *testing.T) {
	sm := models.NewBookingStateMachine()

	if sm == nil {
		t.Fatal("Expected state machine to be created")
	}
}

// ========================================================================
// VALID TRANSITIONS TESTS
// ========================================================================

func TestValidTransitions_PendingState(t *testing.T) {
	sm := models.NewBookingStateMachine()

	validTransitions := []string{
		config.BookingStatusConfirmed,
		config.BookingStatusCancelled,
		config.BookingStatusNoShow,
	}

	for _, toStatus := range validTransitions {
		err := sm.ValidateTransition(config.BookingStatusPending, toStatus)
		if err != nil {
			t.Errorf("Expected transition from pending to %s to be valid, got error: %v", toStatus, err)
		}
	}
}

func TestValidTransitions_ConfirmedState(t *testing.T) {
	sm := models.NewBookingStateMachine()

	validTransitions := []string{
		config.BookingStatusInProgress,
		config.BookingStatusCancelled,
		config.BookingStatusNoShow,
	}

	for _, toStatus := range validTransitions {
		err := sm.ValidateTransition(config.BookingStatusConfirmed, toStatus)
		if err != nil {
			t.Errorf("Expected transition from confirmed to %s to be valid, got error: %v", toStatus, err)
		}
	}
}

func TestValidTransitions_InProgressState(t *testing.T) {
	sm := models.NewBookingStateMachine()

	validTransitions := []string{
		config.BookingStatusCompleted,
		config.BookingStatusCancelled,
	}

	for _, toStatus := range validTransitions {
		err := sm.ValidateTransition(config.BookingStatusInProgress, toStatus)
		if err != nil {
			t.Errorf("Expected transition from in_progress to %s to be valid, got error: %v", toStatus, err)
		}
	}
}

// ========================================================================
// INVALID TRANSITIONS TESTS
// ========================================================================

func TestInvalidTransitions_FromCompleted(t *testing.T) {
	sm := models.NewBookingStateMachine()

	// Completed is a terminal state - no transitions allowed
	invalidTransitions := []string{
		config.BookingStatusPending,
		config.BookingStatusConfirmed,
		config.BookingStatusInProgress,
		config.BookingStatusCancelled,
		config.BookingStatusNoShow,
	}

	for _, toStatus := range invalidTransitions {
		err := sm.ValidateTransition(config.BookingStatusCompleted, toStatus)
		if err == nil {
			t.Errorf("Expected transition from completed to %s to be INVALID", toStatus)
		}
	}
}

func TestInvalidTransitions_FromCancelled(t *testing.T) {
	sm := models.NewBookingStateMachine()

	// Cancelled is a terminal state - no transitions allowed
	invalidTransitions := []string{
		config.BookingStatusPending,
		config.BookingStatusConfirmed,
		config.BookingStatusInProgress,
		config.BookingStatusCompleted,
		config.BookingStatusNoShow,
	}

	for _, toStatus := range invalidTransitions {
		err := sm.ValidateTransition(config.BookingStatusCancelled, toStatus)
		if err == nil {
			t.Errorf("Expected transition from cancelled to %s to be INVALID", toStatus)
		}
	}
}

func TestInvalidTransitions_FromNoShow(t *testing.T) {
	sm := models.NewBookingStateMachine()

	// No-show is a terminal state - no transitions allowed
	invalidTransitions := []string{
		config.BookingStatusPending,
		config.BookingStatusConfirmed,
		config.BookingStatusInProgress,
		config.BookingStatusCompleted,
		config.BookingStatusCancelled,
	}

	for _, toStatus := range invalidTransitions {
		err := sm.ValidateTransition(config.BookingStatusNoShow, toStatus)
		if err == nil {
			t.Errorf("Expected transition from no_show to %s to be INVALID", toStatus)
		}
	}
}

func TestInvalidTransitions_BackwardFlow(t *testing.T) {
	sm := models.NewBookingStateMachine()

	// Test backward transitions (should all be invalid)
	invalidBackwardTransitions := map[string]string{
		config.BookingStatusConfirmed:  config.BookingStatusPending,
		config.BookingStatusInProgress: config.BookingStatusPending,
		config.BookingStatusCompleted:  config.BookingStatusInProgress,
	}

	for fromStatus, toStatus := range invalidBackwardTransitions {
		err := sm.ValidateTransition(fromStatus, toStatus)
		if err == nil {
			t.Errorf("Expected backward transition from %s to %s to be INVALID", fromStatus, toStatus)
		}
	}
}

// ========================================================================
// ERROR MESSAGE TESTS
// ========================================================================

func TestValidateTransition_InvalidCurrentStatus(t *testing.T) {
	sm := models.NewBookingStateMachine()

	err := sm.ValidateTransition("invalid_status", config.BookingStatusConfirmed)
	if err == nil {
		t.Error("Expected error for invalid current status")
	}

	if !strings.Contains(err.Error(), "invalid current status") {
		t.Errorf("Expected error message about invalid current status, got: %v", err)
	}
}

func TestValidateTransition_InvalidTargetStatus(t *testing.T) {
	sm := models.NewBookingStateMachine()

	err := sm.ValidateTransition(config.BookingStatusPending, "invalid_target")
	if err == nil {
		t.Error("Expected error for invalid target status")
	}

	if !strings.Contains(err.Error(), "invalid target status") {
		t.Errorf("Expected error message about invalid target status, got: %v", err)
	}
}

func TestValidateTransition_ErrorContainsAllowedTransitions(t *testing.T) {
	sm := models.NewBookingStateMachine()

	// Try invalid transition from completed
	err := sm.ValidateTransition(config.BookingStatusCompleted, config.BookingStatusPending)
	if err == nil {
		t.Fatal("Expected error for invalid transition")
	}

	// Error message should contain current status, target status, and allowed transitions
	errorMsg := err.Error()
	if !strings.Contains(errorMsg, "completed") {
		t.Error("Error message should contain current status")
	}
	if !strings.Contains(errorMsg, "pending") {
		t.Error("Error message should contain target status")
	}
	if !strings.Contains(errorMsg, "Allowed transitions") {
		t.Error("Error message should mention allowed transitions")
	}
}

// ========================================================================
// TERMINAL STATE TESTS
// ========================================================================

func TestIsTerminalState(t *testing.T) {
	sm := models.NewBookingStateMachine()

	terminalStates := []string{
		config.BookingStatusCompleted,
		config.BookingStatusCancelled,
		config.BookingStatusNoShow,
	}

	for _, status := range terminalStates {
		if !sm.IsTerminalState(status) {
			t.Errorf("Expected %s to be a terminal state", status)
		}
	}

	nonTerminalStates := []string{
		config.BookingStatusPending,
		config.BookingStatusConfirmed,
		config.BookingStatusInProgress,
	}

	for _, status := range nonTerminalStates {
		if sm.IsTerminalState(status) {
			t.Errorf("Expected %s to NOT be a terminal state", status)
		}
	}
}

// ========================================================================
// GET ALLOWED TRANSITIONS TESTS
// ========================================================================

func TestGetAllowedTransitions_Pending(t *testing.T) {
	sm := models.NewBookingStateMachine()

	transitions, err := sm.GetAllowedTransitions(config.BookingStatusPending)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := 3 // confirmed, cancelled, no_show
	if len(transitions) != expected {
		t.Errorf("Expected %d allowed transitions, got %d", expected, len(transitions))
	}

	// Check all expected transitions are present
	expectedTransitions := map[string]bool{
		config.BookingStatusConfirmed: true,
		config.BookingStatusCancelled: true,
		config.BookingStatusNoShow:    true,
	}

	for _, trans := range transitions {
		if !expectedTransitions[trans] {
			t.Errorf("Unexpected transition: %s", trans)
		}
	}
}

func TestGetAllowedTransitions_TerminalState(t *testing.T) {
	sm := models.NewBookingStateMachine()

	transitions, err := sm.GetAllowedTransitions(config.BookingStatusCompleted)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(transitions) != 0 {
		t.Errorf("Terminal state should have no allowed transitions, got %d", len(transitions))
	}
}

func TestGetAllowedTransitions_InvalidStatus(t *testing.T) {
	sm := models.NewBookingStateMachine()

	_, err := sm.GetAllowedTransitions("invalid_status")
	if err == nil {
		t.Error("Expected error for invalid status")
	}
}

// ========================================================================
// BOOKING MODEL INTEGRATION TESTS
// ========================================================================

func TestBooking_CanTransitionTo(t *testing.T) {
	booking := &models.Booking{
		Status: config.BookingStatusPending,
	}

	// Valid transition
	if !booking.CanTransitionTo(config.BookingStatusConfirmed) {
		t.Error("Expected booking to allow transition to confirmed")
	}

	// Invalid transition
	if booking.CanTransitionTo(config.BookingStatusCompleted) {
		t.Error("Expected booking to NOT allow direct transition to completed")
	}
}

func TestBooking_GetAllowedStatusTransitions(t *testing.T) {
	booking := &models.Booking{
		Status: config.BookingStatusConfirmed,
	}

	transitions := booking.GetAllowedStatusTransitions()

	expected := 3 // in_progress, cancelled, no_show
	if len(transitions) != expected {
		t.Errorf("Expected %d transitions, got %d", expected, len(transitions))
	}
}

func TestBooking_IsInTerminalState(t *testing.T) {
	// Terminal state booking
	completedBooking := &models.Booking{
		Status: config.BookingStatusCompleted,
	}
	if !completedBooking.IsInTerminalState() {
		t.Error("Completed booking should be in terminal state")
	}

	// Non-terminal state booking
	pendingBooking := &models.Booking{
		Status: config.BookingStatusPending,
	}
	if pendingBooking.IsInTerminalState() {
		t.Error("Pending booking should NOT be in terminal state")
	}
}

func TestBooking_ValidateStatusTransition(t *testing.T) {
	booking := &models.Booking{
		Status: config.BookingStatusPending,
	}

	// Valid transition
	err := booking.ValidateStatusTransition(config.BookingStatusConfirmed)
	if err != nil {
		t.Errorf("Expected valid transition, got error: %v", err)
	}

	// Invalid transition
	err = booking.ValidateStatusTransition(config.BookingStatusCompleted)
	if err == nil {
		t.Error("Expected error for invalid transition")
	}
}

// ========================================================================
// COMPLETE FLOW TESTS
// ========================================================================

func TestCompleteBookingFlow_HappyPath(t *testing.T) {
	sm := models.NewBookingStateMachine()

	// Simulate complete happy path: pending → confirmed → in_progress → completed
	steps := []struct {
		from string
		to   string
	}{
		{config.BookingStatusPending, config.BookingStatusConfirmed},
		{config.BookingStatusConfirmed, config.BookingStatusInProgress},
		{config.BookingStatusInProgress, config.BookingStatusCompleted},
	}

	for _, step := range steps {
		err := sm.ValidateTransition(step.from, step.to)
		if err != nil {
			t.Errorf("Expected valid transition from %s to %s, got error: %v", step.from, step.to, err)
		}
	}
}

func TestCompleteBookingFlow_CancellationPath(t *testing.T) {
	sm := models.NewBookingStateMachine()

	// Test cancellation at various stages
	cancellationPoints := []string{
		config.BookingStatusPending,
		config.BookingStatusConfirmed,
		config.BookingStatusInProgress,
	}

	for _, fromStatus := range cancellationPoints {
		err := sm.ValidateTransition(fromStatus, config.BookingStatusCancelled)
		if err != nil {
			t.Errorf("Expected valid cancellation from %s, got error: %v", fromStatus, err)
		}
	}
}

func TestCompleteBookingFlow_NoShowPath(t *testing.T) {
	sm := models.NewBookingStateMachine()

	// Test no-show marking at valid stages
	noShowPoints := []string{
		config.BookingStatusPending,
		config.BookingStatusConfirmed,
	}

	for _, fromStatus := range noShowPoints {
		err := sm.ValidateTransition(fromStatus, config.BookingStatusNoShow)
		if err != nil {
			t.Errorf("Expected valid no-show marking from %s, got error: %v", fromStatus, err)
		}
	}

	// Cannot mark as no-show after service started
	err := sm.ValidateTransition(config.BookingStatusInProgress, config.BookingStatusNoShow)
	if err == nil {
		t.Error("Should not be able to mark as no-show after service started")
	}
}

// ========================================================================
// EDGE CASE TESTS
// ========================================================================

func TestCanTransition_ConvenienceMethod(t *testing.T) {
	sm := models.NewBookingStateMachine()

	// Valid transition
	if !sm.CanTransition(config.BookingStatusPending, config.BookingStatusConfirmed) {
		t.Error("CanTransition should return true for valid transition")
	}

	// Invalid transition
	if sm.CanTransition(config.BookingStatusCompleted, config.BookingStatusPending) {
		t.Error("CanTransition should return false for invalid transition")
	}
}

func TestStateMachine_SameStatusTransition(t *testing.T) {
	sm := models.NewBookingStateMachine()

	// Transitioning to the same status should be invalid
	err := sm.ValidateTransition(config.BookingStatusPending, config.BookingStatusPending)
	if err == nil {
		t.Error("Transitioning to the same status should be invalid")
	}
}

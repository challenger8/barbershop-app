// internal/models/booking_state_machine.go
package models

import (
	"fmt"

	"barber-booking-system/internal/config"
)

// ========================================================================
// BOOKING STATE MACHINE - Enforce Valid Status Transitions
// ========================================================================
//
// Purpose: Prevent invalid booking status transitions
// Benefits:
//   - Business rule enforcement at the model level
//   - Clear documentation of allowed transitions
//   - Prevents data corruption
//   - Easy to maintain and extend
//
// Valid State Flow:
//                    ┌─────────────┐
//                    │   PENDING   │
//                    └──────┬──────┘
//                           │
//              ┌────────────┼────────────┐
//              ↓            ↓            ↓
//         CONFIRMED    CANCELLED    NO_SHOW
//              │
//              ↓
//        IN_PROGRESS
//              │
//         ┌────┴────┐
//         ↓         ↓
//    COMPLETED  CANCELLED
//
// ========================================================================

// BookingState represents a booking status and its allowed transitions
type BookingState interface {
	// CanTransitionTo checks if transition to newStatus is valid
	CanTransitionTo(newStatus string) bool

	// GetAllowedTransitions returns all valid next states
	GetAllowedTransitions() []string

	// GetStatusName returns the current status name
	GetStatusName() string

	// IsTerminal returns true if this is a final state (no transitions allowed)
	IsTerminal() bool
}

// ========================================================================
// CONCRETE STATE IMPLEMENTATIONS
// ========================================================================

// PendingState represents a newly created booking
type PendingState struct{}

func (s *PendingState) CanTransitionTo(newStatus string) bool {
	allowedTransitions := map[string]bool{
		config.BookingStatusConfirmed:           true,
		config.BookingStatusCancelled:           true,
		config.BookingStatusCancelledByCustomer: true,
		config.BookingStatusCancelledByBarber:   true,
		config.BookingStatusNoShow:              true,
	}
	return allowedTransitions[newStatus]
}

func (s *PendingState) GetAllowedTransitions() []string {
	return []string{
		config.BookingStatusConfirmed,
		config.BookingStatusCancelled,
		config.BookingStatusNoShow,
	}
}

func (s *PendingState) GetStatusName() string {
	return config.BookingStatusPending
}

func (s *PendingState) IsTerminal() bool {
	return false
}

// ConfirmedState represents a confirmed booking
type ConfirmedState struct{}

func (s *ConfirmedState) CanTransitionTo(newStatus string) bool {
	allowedTransitions := map[string]bool{
		config.BookingStatusInProgress:          true,
		config.BookingStatusCancelled:           true,
		config.BookingStatusCancelledByCustomer: true,
		config.BookingStatusCancelledByBarber:   true,
		config.BookingStatusNoShow:              true,
	}
	return allowedTransitions[newStatus]
}

func (s *ConfirmedState) GetAllowedTransitions() []string {
	return []string{
		config.BookingStatusInProgress,
		config.BookingStatusCancelled,
		config.BookingStatusNoShow,
	}
}

func (s *ConfirmedState) GetStatusName() string {
	return config.BookingStatusConfirmed
}

func (s *ConfirmedState) IsTerminal() bool {
	return false
}

// InProgressState represents a booking in progress
type InProgressState struct{}

func (s *InProgressState) CanTransitionTo(newStatus string) bool {
	allowedTransitions := map[string]bool{
		config.BookingStatusCompleted:           true,
		config.BookingStatusCancelled:           true,
		config.BookingStatusCancelledByCustomer: true,
		config.BookingStatusCancelledByBarber:   true,
	}
	return allowedTransitions[newStatus]
}

func (s *InProgressState) GetAllowedTransitions() []string {
	return []string{
		config.BookingStatusCompleted,
		config.BookingStatusCancelled,
	}
}

func (s *InProgressState) GetStatusName() string {
	return config.BookingStatusInProgress
}

func (s *InProgressState) IsTerminal() bool {
	return false
}

// CompletedState represents a finished booking (TERMINAL STATE)
type CompletedState struct{}

func (s *CompletedState) CanTransitionTo(newStatus string) bool {
	// Completed is a terminal state - no transitions allowed
	return false
}

func (s *CompletedState) GetAllowedTransitions() []string {
	return []string{} // No transitions from completed
}

func (s *CompletedState) GetStatusName() string {
	return config.BookingStatusCompleted
}

func (s *CompletedState) IsTerminal() bool {
	return true
}

// CancelledState represents a cancelled booking (TERMINAL STATE)
type CancelledState struct{}

func (s *CancelledState) CanTransitionTo(newStatus string) bool {
	// Cancelled is a terminal state - no transitions allowed
	return false
}

func (s *CancelledState) GetAllowedTransitions() []string {
	return []string{} // No transitions from cancelled
}

func (s *CancelledState) GetStatusName() string {
	return config.BookingStatusCancelled
}

func (s *CancelledState) IsTerminal() bool {
	return true
}

// NoShowState represents a no-show booking (TERMINAL STATE)
type NoShowState struct{}

func (s *NoShowState) CanTransitionTo(newStatus string) bool {
	// No-show is a terminal state - no transitions allowed
	return false
}

func (s *NoShowState) GetAllowedTransitions() []string {
	return []string{} // No transitions from no-show
}

func (s *NoShowState) GetStatusName() string {
	return config.BookingStatusNoShow
}

func (s *NoShowState) IsTerminal() bool {
	return true
}

// ========================================================================
// STATE MACHINE
// ========================================================================

// BookingStateMachine manages booking status transitions
type BookingStateMachine struct {
	states map[string]BookingState
}

// NewBookingStateMachine creates a new state machine
func NewBookingStateMachine() *BookingStateMachine {
	return &BookingStateMachine{
		states: map[string]BookingState{
			config.BookingStatusPending:             &PendingState{},
			config.BookingStatusConfirmed:           &ConfirmedState{},
			config.BookingStatusInProgress:          &InProgressState{},
			config.BookingStatusCompleted:           &CompletedState{},
			config.BookingStatusCancelled:           &CancelledState{},
			config.BookingStatusCancelledByCustomer: &CancelledState{},
			config.BookingStatusCancelledByBarber:   &CancelledState{},
			config.BookingStatusNoShow:              &NoShowState{},
		},
	}
}

// ValidateTransition checks if a status transition is valid
func (sm *BookingStateMachine) ValidateTransition(fromStatus, toStatus string) error {
	// Get current state
	currentState, exists := sm.states[fromStatus]
	if !exists {
		return fmt.Errorf("invalid current status: %s", fromStatus)
	}

	// Get target state
	_, exists = sm.states[toStatus]
	if !exists {
		return fmt.Errorf("invalid target status: %s", toStatus)
	}

	// Check if transition is allowed
	if !currentState.CanTransitionTo(toStatus) {
		return fmt.Errorf(
			"invalid status transition: cannot change from '%s' to '%s'. Allowed transitions: %v",
			fromStatus,
			toStatus,
			currentState.GetAllowedTransitions(),
		)
	}

	return nil
}

// GetAllowedTransitions returns all valid transitions from a given status
func (sm *BookingStateMachine) GetAllowedTransitions(fromStatus string) ([]string, error) {
	state, exists := sm.states[fromStatus]
	if !exists {
		return nil, fmt.Errorf("invalid status: %s", fromStatus)
	}

	return state.GetAllowedTransitions(), nil
}

// IsTerminalState checks if a status is a terminal state
func (sm *BookingStateMachine) IsTerminalState(status string) bool {
	state, exists := sm.states[status]
	if !exists {
		return false
	}

	return state.IsTerminal()
}

// CanTransition is a convenience method to check if a transition is valid
func (sm *BookingStateMachine) CanTransition(fromStatus, toStatus string) bool {
	err := sm.ValidateTransition(fromStatus, toStatus)
	return err == nil
}

// ========================================================================
// BOOKING MODEL EXTENSIONS
// ========================================================================

// CanTransitionTo checks if this booking can transition to the new status
func (b *Booking) CanTransitionTo(newStatus string) bool {
	sm := NewBookingStateMachine()
	return sm.CanTransition(b.Status, newStatus)
}

// GetAllowedStatusTransitions returns all valid next states for this booking
func (b *Booking) GetAllowedStatusTransitions() []string {
	sm := NewBookingStateMachine()
	transitions, _ := sm.GetAllowedTransitions(b.Status)
	return transitions
}

// IsInTerminalState checks if the booking is in a terminal state
func (b *Booking) IsInTerminalState() bool {
	sm := NewBookingStateMachine()
	return sm.IsTerminalState(b.Status)
}

// ValidateStatusTransition validates if a status transition is allowed
func (b *Booking) ValidateStatusTransition(newStatus string) error {
	sm := NewBookingStateMachine()
	return sm.ValidateTransition(b.Status, newStatus)
}

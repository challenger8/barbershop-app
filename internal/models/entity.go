// internal/models/entity.go
package models

// ========================================================================
// ENTITY INTERFACE IMPLEMENTATIONS
// ========================================================================
// These methods enable the BaseRepository generic pattern.
// Each model implements TableName() and GetID().
// ========================================================================

// User entity methods
func (u User) TableName() string { return "users" }
func (u User) GetID() int        { return u.ID }

// Barber entity methods
func (b Barber) TableName() string { return "barbers" }
func (b Barber) GetID() int        { return b.ID }

// Booking entity methods
func (b Booking) TableName() string { return "bookings" }
func (b Booking) GetID() int        { return b.ID }

// Review entity methods
func (r Review) TableName() string { return "reviews" }
func (r Review) GetID() int        { return r.ID }

// Service entity methods
func (s Service) TableName() string { return "services" }
func (s Service) GetID() int        { return s.ID }

// ServiceCategory entity methods
func (sc ServiceCategory) TableName() string { return "service_categories" }
func (sc ServiceCategory) GetID() int        { return sc.ID }

// BarberService entity methods
func (bs BarberService) TableName() string { return "barber_services" }
func (bs BarberService) GetID() int        { return bs.ID }

// Notification entity methods
func (n Notification) TableName() string { return "notifications" }
func (n Notification) GetID() int        { return n.ID }

// TimeSlot entity methods
func (ts TimeSlot) TableName() string { return "time_slots" }
func (ts TimeSlot) GetID() int        { return ts.ID }

// BookingHistory entity methods
func (bh BookingHistory) TableName() string { return "booking_history" }
func (bh BookingHistory) GetID() int        { return bh.ID }
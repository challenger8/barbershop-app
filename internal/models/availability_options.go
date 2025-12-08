// internal/models/availability_options.go
package models

import "time"

// ========================================================================
// AVAILABILITY CHECK OPTIONS - Replace Long Parameter Lists
// ========================================================================
//
// Before: checkTimeSlotAvailability(ctx, barberID, startTime, endTime, excludeBookingID)
// After:  checkTimeSlotAvailability(ctx, barberID, opts)
//
// Benefits:
//   - Cleaner function signatures
//   - Optional parameters are truly optional
//   - Easy to extend without breaking changes
//   - Self-documenting code
// ========================================================================

// TimeSlotCheckOptions represents options for checking time slot availability
type TimeSlotCheckOptions struct {
	StartTime        time.Time
	EndTime          time.Time
	ExcludeBookingID int
	CheckBufferTime  bool
	BufferMinutes    int
}

// TimeSlotCheckOption is a function that modifies TimeSlotCheckOptions
type TimeSlotCheckOption func(*TimeSlotCheckOptions)

// NewTimeSlotCheckOptions creates options for time slot checking
func NewTimeSlotCheckOptions(startTime, endTime time.Time, opts ...TimeSlotCheckOption) *TimeSlotCheckOptions {
	options := &TimeSlotCheckOptions{
		StartTime:        startTime,
		EndTime:          endTime,
		ExcludeBookingID: 0,
		CheckBufferTime:  false,
		BufferMinutes:    0,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// Option functions for TimeSlotCheckOptions

// WithExcludeBooking excludes a specific booking from conflict check (for rescheduling)
func WithExcludeBooking(bookingID int) TimeSlotCheckOption {
	return func(o *TimeSlotCheckOptions) {
		o.ExcludeBookingID = bookingID
	}
}

// WithBufferTime adds buffer time before/after the booking
func WithBufferTime(minutes int) TimeSlotCheckOption {
	return func(o *TimeSlotCheckOptions) {
		o.CheckBufferTime = true
		o.BufferMinutes = minutes
	}
}

// GetEffectiveStartTime returns start time minus buffer
func (o *TimeSlotCheckOptions) GetEffectiveStartTime() time.Time {
	if o.CheckBufferTime && o.BufferMinutes > 0 {
		return o.StartTime.Add(-time.Duration(o.BufferMinutes) * time.Minute)
	}
	return o.StartTime
}

// GetEffectiveEndTime returns end time plus buffer
func (o *TimeSlotCheckOptions) GetEffectiveEndTime() time.Time {
	if o.CheckBufferTime && o.BufferMinutes > 0 {
		return o.EndTime.Add(time.Duration(o.BufferMinutes) * time.Minute)
	}
	return o.EndTime
}

// ========================================================================
// STATS OPTIONS - Enhanced from booking_options.go
// ========================================================================

// StatsQueryOptions represents enhanced options for statistics queries
type StatsQueryOptions struct {
	FromDate       time.Time
	ToDate         time.Time
	GroupBy        string // "day", "week", "month"
	IncludeTrends  bool
	IncludeRevenue bool
	IncludeRatings bool
	CompareWith    *time.Time // Compare with previous period
	Filters        map[string]interface{}
}

// StatsQueryOption is a function that modifies StatsQueryOptions
type StatsQueryOption func(*StatsQueryOptions)

// NewStatsQueryOptions creates enhanced stats options
func NewStatsQueryOptions(from, to time.Time, opts ...StatsQueryOption) *StatsQueryOptions {
	options := &StatsQueryOptions{
		FromDate:       from,
		ToDate:         to,
		GroupBy:        "day",
		IncludeTrends:  false,
		IncludeRevenue: true,
		IncludeRatings: false,
		Filters:        make(map[string]interface{}),
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// Option functions for StatsQueryOptions

// WithGrouping sets the time grouping for stats
func WithGrouping(groupBy string) StatsQueryOption {
	return func(o *StatsQueryOptions) {
		o.GroupBy = groupBy
	}
}

// WithTrendsAnalysis includes trend analysis in results
func WithTrendsAnalysis() StatsQueryOption {
	return func(o *StatsQueryOptions) {
		o.IncludeTrends = true
	}
}

// WithRevenueBreakdown includes detailed revenue breakdown
func WithRevenueBreakdown() StatsQueryOption {
	return func(o *StatsQueryOptions) {
		o.IncludeRevenue = true
	}
}

// WithRatingsAnalysis includes rating and review metrics
func WithRatingsAnalysis() StatsQueryOption {
	return func(o *StatsQueryOptions) {
		o.IncludeRatings = true
	}
}

// WithComparison compares with a previous period
func WithComparison(compareDate time.Time) StatsQueryOption {
	return func(o *StatsQueryOptions) {
		o.CompareWith = &compareDate
	}
}

// WithFilter adds a custom filter
func WithFilter(key string, value interface{}) StatsQueryOption {
	return func(o *StatsQueryOptions) {
		o.Filters[key] = value
	}
}

// ========================================================================
// USAGE EXAMPLES
// ========================================================================
//
// Example 1: Simple time slot check
//   opts := NewTimeSlotCheckOptions(startTime, endTime)
//   err := service.checkTimeSlotAvailability(ctx, barberID, opts)
//
// Example 2: Time slot check with exclusion (for rescheduling)
//   opts := NewTimeSlotCheckOptions(startTime, endTime,
//       WithExcludeBooking(existingBookingID))
//   err := service.checkTimeSlotAvailability(ctx, barberID, opts)
//
// Example 3: Time slot check with buffer time
//   opts := NewTimeSlotCheckOptions(startTime, endTime,
//       WithBufferTime(15)) // 15 minute buffer
//   err := service.checkTimeSlotAvailability(ctx, barberID, opts)
//
// Example 4: Enhanced stats query
//   opts := NewStatsQueryOptions(from, to,
//       WithGrouping("week"),
//       WithTrendsAnalysis(),
//       WithRatingsAnalysis(),
//       WithFilter("service_type", "haircut"))
//   stats := service.GetBarberStatsEnhanced(ctx, barberID, opts)
// ========================================================================

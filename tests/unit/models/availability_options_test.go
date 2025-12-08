// tests/unit/models/availability_options_test.go
package models

import (
	"testing"
	"time"

	"barber-booking-system/internal/models"
)

// ========================================================================
// TIME SLOT CHECK OPTIONS TESTS
// ========================================================================

func TestNewTimeSlotCheckOptions_Defaults(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)

	opts := models.NewTimeSlotCheckOptions(startTime, endTime)

	if !opts.StartTime.Equal(startTime) {
		t.Error("StartTime not set correctly")
	}

	if !opts.EndTime.Equal(endTime) {
		t.Error("EndTime not set correctly")
	}

	if opts.ExcludeBookingID != 0 {
		t.Errorf("Expected ExcludeBookingID 0, got %d", opts.ExcludeBookingID)
	}

	if opts.CheckBufferTime {
		t.Error("Expected CheckBufferTime to be false by default")
	}

	if opts.BufferMinutes != 0 {
		t.Errorf("Expected BufferMinutes 0, got %d", opts.BufferMinutes)
	}
}

func TestTimeSlotCheckOptions_WithExcludeBooking(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)
	bookingID := 123

	opts := models.NewTimeSlotCheckOptions(startTime, endTime,
		models.WithExcludeBooking(bookingID))

	if opts.ExcludeBookingID != bookingID {
		t.Errorf("Expected ExcludeBookingID %d, got %d", bookingID, opts.ExcludeBookingID)
	}
}

func TestTimeSlotCheckOptions_WithBufferTime(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)
	bufferMinutes := 15

	opts := models.NewTimeSlotCheckOptions(startTime, endTime,
		models.WithBufferTime(bufferMinutes))

	if !opts.CheckBufferTime {
		t.Error("Expected CheckBufferTime to be true")
	}

	if opts.BufferMinutes != bufferMinutes {
		t.Errorf("Expected BufferMinutes %d, got %d", bufferMinutes, opts.BufferMinutes)
	}
}

func TestTimeSlotCheckOptions_MultipleOptions(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)
	bookingID := 123
	bufferMinutes := 15

	opts := models.NewTimeSlotCheckOptions(startTime, endTime,
		models.WithExcludeBooking(bookingID),
		models.WithBufferTime(bufferMinutes))

	if opts.ExcludeBookingID != bookingID {
		t.Error("ExcludeBookingID not set correctly")
	}

	if !opts.CheckBufferTime {
		t.Error("CheckBufferTime not set correctly")
	}

	if opts.BufferMinutes != bufferMinutes {
		t.Error("BufferMinutes not set correctly")
	}
}

// ========================================================================
// EFFECTIVE TIME TESTS
// ========================================================================

func TestTimeSlotCheckOptions_GetEffectiveStartTime_NoBuffer(t *testing.T) {
	startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	endTime := startTime.Add(1 * time.Hour)

	opts := models.NewTimeSlotCheckOptions(startTime, endTime)

	effectiveStart := opts.GetEffectiveStartTime()

	if !effectiveStart.Equal(startTime) {
		t.Error("Effective start time should equal start time when no buffer")
	}
}

func TestTimeSlotCheckOptions_GetEffectiveStartTime_WithBuffer(t *testing.T) {
	startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	endTime := startTime.Add(1 * time.Hour)
	bufferMinutes := 15

	opts := models.NewTimeSlotCheckOptions(startTime, endTime,
		models.WithBufferTime(bufferMinutes))

	effectiveStart := opts.GetEffectiveStartTime()

	// Should be 15 minutes BEFORE start time
	expected := startTime.Add(-15 * time.Minute)

	if !effectiveStart.Equal(expected) {
		t.Errorf("Expected effective start %v, got %v", expected, effectiveStart)
	}
}

func TestTimeSlotCheckOptions_GetEffectiveEndTime_NoBuffer(t *testing.T) {
	startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	endTime := startTime.Add(1 * time.Hour)

	opts := models.NewTimeSlotCheckOptions(startTime, endTime)

	effectiveEnd := opts.GetEffectiveEndTime()

	if !effectiveEnd.Equal(endTime) {
		t.Error("Effective end time should equal end time when no buffer")
	}
}

func TestTimeSlotCheckOptions_GetEffectiveEndTime_WithBuffer(t *testing.T) {
	startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	endTime := startTime.Add(1 * time.Hour)
	bufferMinutes := 15

	opts := models.NewTimeSlotCheckOptions(startTime, endTime,
		models.WithBufferTime(bufferMinutes))

	effectiveEnd := opts.GetEffectiveEndTime()

	// Should be 15 minutes AFTER end time
	expected := endTime.Add(15 * time.Minute)

	if !effectiveEnd.Equal(expected) {
		t.Errorf("Expected effective end %v, got %v", expected, effectiveEnd)
	}
}

func TestTimeSlotCheckOptions_EffectiveTimes_CompleteScenario(t *testing.T) {
	// Booking: 10:00 AM - 11:00 AM
	startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)
	bufferMinutes := 15

	opts := models.NewTimeSlotCheckOptions(startTime, endTime,
		models.WithBufferTime(bufferMinutes))

	effectiveStart := opts.GetEffectiveStartTime()
	effectiveEnd := opts.GetEffectiveEndTime()

	// With 15-min buffer: 9:45 AM - 11:15 AM
	expectedStart := time.Date(2024, 1, 15, 9, 45, 0, 0, time.UTC)
	expectedEnd := time.Date(2024, 1, 15, 11, 15, 0, 0, time.UTC)

	if !effectiveStart.Equal(expectedStart) {
		t.Errorf("Expected effective start %v, got %v", expectedStart, effectiveStart)
	}

	if !effectiveEnd.Equal(expectedEnd) {
		t.Errorf("Expected effective end %v, got %v", expectedEnd, effectiveEnd)
	}
}

// ========================================================================
// STATS QUERY OPTIONS TESTS
// ========================================================================

func TestNewStatsQueryOptions_Defaults(t *testing.T) {
	from := time.Now().AddDate(0, -1, 0)
	to := time.Now()

	opts := models.NewStatsQueryOptions(from, to)

	if !opts.FromDate.Equal(from) {
		t.Error("FromDate not set correctly")
	}

	if !opts.ToDate.Equal(to) {
		t.Error("ToDate not set correctly")
	}

	if opts.GroupBy != "day" {
		t.Errorf("Expected GroupBy 'day', got %s", opts.GroupBy)
	}

	if opts.IncludeTrends {
		t.Error("Expected IncludeTrends to be false by default")
	}

	if !opts.IncludeRevenue {
		t.Error("Expected IncludeRevenue to be true by default")
	}

	if opts.IncludeRatings {
		t.Error("Expected IncludeRatings to be false by default")
	}

	if opts.CompareWith != nil {
		t.Error("Expected CompareWith to be nil by default")
	}

	if opts.Filters == nil {
		t.Error("Expected Filters map to be initialized")
	}
}

func TestStatsQueryOptions_WithGrouping(t *testing.T) {
	from := time.Now().AddDate(0, -1, 0)
	to := time.Now()

	opts := models.NewStatsQueryOptions(from, to,
		models.WithGrouping("week"))

	if opts.GroupBy != "week" {
		t.Errorf("Expected GroupBy 'week', got %s", opts.GroupBy)
	}
}

func TestStatsQueryOptions_WithTrendsAnalysis(t *testing.T) {
	from := time.Now().AddDate(0, -1, 0)
	to := time.Now()

	opts := models.NewStatsQueryOptions(from, to,
		models.WithTrendsAnalysis())

	if !opts.IncludeTrends {
		t.Error("Expected IncludeTrends to be true")
	}
}

func TestStatsQueryOptions_WithRevenueBreakdown(t *testing.T) {
	from := time.Now().AddDate(0, -1, 0)
	to := time.Now()

	opts := models.NewStatsQueryOptions(from, to,
		models.WithRevenueBreakdown())

	if !opts.IncludeRevenue {
		t.Error("Expected IncludeRevenue to be true")
	}
}

func TestStatsQueryOptions_WithRatingsAnalysis(t *testing.T) {
	from := time.Now().AddDate(0, -1, 0)
	to := time.Now()

	opts := models.NewStatsQueryOptions(from, to,
		models.WithRatingsAnalysis())

	if !opts.IncludeRatings {
		t.Error("Expected IncludeRatings to be true")
	}
}

func TestStatsQueryOptions_WithComparison(t *testing.T) {
	from := time.Now().AddDate(0, -1, 0)
	to := time.Now()
	compareDate := time.Now().AddDate(0, -2, 0)

	opts := models.NewStatsQueryOptions(from, to,
		models.WithComparison(compareDate))

	if opts.CompareWith == nil {
		t.Fatal("Expected CompareWith to be set")
	}

	if !opts.CompareWith.Equal(compareDate) {
		t.Error("CompareWith not set correctly")
	}
}

func TestStatsQueryOptions_WithFilter(t *testing.T) {
	from := time.Now().AddDate(0, -1, 0)
	to := time.Now()

	opts := models.NewStatsQueryOptions(from, to,
		models.WithFilter("service_type", "haircut"),
		models.WithFilter("status", "completed"))

	serviceType, exists := opts.Filters["service_type"]
	if !exists {
		t.Error("Expected service_type filter to exist")
	}

	if serviceType != "haircut" {
		t.Errorf("Expected service_type 'haircut', got %v", serviceType)
	}

	status, exists := opts.Filters["status"]
	if !exists {
		t.Error("Expected status filter to exist")
	}

	if status != "completed" {
		t.Errorf("Expected status 'completed', got %v", status)
	}
}

func TestStatsQueryOptions_MultipleOptions(t *testing.T) {
	from := time.Now().AddDate(0, -1, 0)
	to := time.Now()
	compareDate := time.Now().AddDate(0, -2, 0)

	opts := models.NewStatsQueryOptions(from, to,
		models.WithGrouping("month"),
		models.WithTrendsAnalysis(),
		models.WithRatingsAnalysis(),
		models.WithComparison(compareDate),
		models.WithFilter("service_type", "haircut"))

	if opts.GroupBy != "month" {
		t.Error("GroupBy not set correctly")
	}

	if !opts.IncludeTrends {
		t.Error("IncludeTrends not set correctly")
	}

	if !opts.IncludeRatings {
		t.Error("IncludeRatings not set correctly")
	}

	if opts.CompareWith == nil || !opts.CompareWith.Equal(compareDate) {
		t.Error("CompareWith not set correctly")
	}

	if opts.Filters["service_type"] != "haircut" {
		t.Error("Filter not set correctly")
	}
}

// ========================================================================
// INTEGRATION TESTS
// ========================================================================

func TestTimeSlotCheckOptions_RealWorldScenario_Rescheduling(t *testing.T) {
	// Scenario: Rescheduling booking #456 from 2PM to 3PM
	existingBookingID := 456
	newStartTime := time.Date(2024, 1, 15, 15, 0, 0, 0, time.UTC)
	newEndTime := time.Date(2024, 1, 15, 16, 0, 0, 0, time.UTC)

	opts := models.NewTimeSlotCheckOptions(newStartTime, newEndTime,
		models.WithExcludeBooking(existingBookingID),
		models.WithBufferTime(15))

	// Verify exclusion
	if opts.ExcludeBookingID != existingBookingID {
		t.Error("Booking exclusion not working for reschedule scenario")
	}

	// Verify buffer is applied
	effectiveStart := opts.GetEffectiveStartTime()
	expectedStart := newStartTime.Add(-15 * time.Minute) // 2:45 PM

	if !effectiveStart.Equal(expectedStart) {
		t.Error("Buffer time not applied correctly for reschedule")
	}
}

func TestStatsQueryOptions_RealWorldScenario_MonthlyReport(t *testing.T) {
	// Scenario: Monthly performance report with trends and comparisons
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
	compareFrom := time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC)

	opts := models.NewStatsQueryOptions(from, to,
		models.WithGrouping("week"),
		models.WithTrendsAnalysis(),
		models.WithRatingsAnalysis(),
		models.WithRevenueBreakdown(),
		models.WithComparison(compareFrom))

	// Verify all options
	if opts.GroupBy != "week" {
		t.Error("Weekly grouping not set for monthly report")
	}

	if !opts.IncludeTrends || !opts.IncludeRatings || !opts.IncludeRevenue {
		t.Error("Not all analysis options enabled for monthly report")
	}

	if opts.CompareWith == nil {
		t.Error("Comparison not set for monthly report")
	}
}

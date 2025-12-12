// tests/integration/booking_test_helpers.go
package integration

import (
	"time"
)

// =============================================================================
// BOOKING TEST FIXTURES
// =============================================================================

// getTestBookingRequest returns a valid booking request for testing
func getTestBookingRequest() map[string]interface{} {
	return map[string]interface{}{
		"barber_id":        1,
		"service_id":       1,
		"start_time":       time.Now().Add(2 * time.Hour).Format(time.RFC3339),
		"duration_minutes": 45,
		"customer_name":    "Test Customer",
		"customer_email":   "customer@test.com",
		"customer_phone":   "+1234567890",
		"service_price":    50.00,
		"notes":            "Test booking",
	}
}

// getTestBookingRequestWithCustomer returns a booking request with customer ID
func getTestBookingRequestWithCustomer(customerID int) map[string]interface{} {
	req := getTestBookingRequest()
	req["customer_id"] = customerID
	return req
}

// getTestBookingRequestForBarber returns a booking request for a specific barber
func getTestBookingRequestForBarber(barberID int, serviceID int) map[string]interface{} {
	req := getTestBookingRequest()
	req["barber_id"] = barberID
	req["service_id"] = serviceID
	return req
}

// getTestRescheduleRequest returns a valid reschedule request
func getTestRescheduleRequest() map[string]interface{} {
	return map[string]interface{}{
		"new_start_time":   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		"duration_minutes": 45,
		"reason":           "Customer requested different time",
	}
}

// getTestCancelRequest returns a cancel request with reason
func getTestCancelRequest(isByCustomer bool) map[string]interface{} {
	return map[string]interface{}{
		"reason":         "Changed my plans",
		"is_by_customer": isByCustomer,
	}
}

// getTestStatusUpdateRequest returns a status update request
func getTestStatusUpdateRequest(status string) map[string]interface{} {
	return map[string]interface{}{
		"status": status,
	}
}

// getTestBookingRequestInPast returns a booking request with past start time (for testing validation)
func getTestBookingRequestInPast() map[string]interface{} {
	req := getTestBookingRequest()
	req["start_time"] = time.Now().Add(-1 * time.Hour).Format(time.RFC3339) // Past time
	return req
}


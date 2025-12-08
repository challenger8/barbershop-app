// tests/unit/models/pricing_test.go
package models

import (
	"testing"

	"barber-booking-system/internal/models"
)

// ========================================================================
// PRICING BREAKDOWN TESTS
// ========================================================================

func TestCalculatePricing_Basic(t *testing.T) {
	pricing := models.CalculatePricing(100.0, 10.0, 0.08)

	if pricing.ServicePrice != 100.0 {
		t.Errorf("Expected ServicePrice 100.0, got %f", pricing.ServicePrice)
	}

	if pricing.DiscountAmount != 10.0 {
		t.Errorf("Expected DiscountAmount 10.0, got %f", pricing.DiscountAmount)
	}

	// SubTotal = 100 - 10 = 90
	if pricing.SubTotal != 90.0 {
		t.Errorf("Expected SubTotal 90.0, got %f", pricing.SubTotal)
	}

	// TaxAmount = 90 * 0.08 = 7.2
	expectedTax := 7.2
	if pricing.TaxAmount != expectedTax {
		t.Errorf("Expected TaxAmount %f, got %f", expectedTax, pricing.TaxAmount)
	}

	// TotalPrice = 90 + 7.2 = 97.2
	expectedTotal := 97.2
	if pricing.TotalPrice != expectedTotal {
		t.Errorf("Expected TotalPrice %f, got %f", expectedTotal, pricing.TotalPrice)
	}

	if pricing.Currency != "USD" {
		t.Errorf("Expected Currency USD, got %s", pricing.Currency)
	}
}

func TestCalculatePricing_NoDiscount(t *testing.T) {
	pricing := models.CalculatePricing(100.0, 0.0, 0.08)

	// SubTotal = 100 - 0 = 100
	if pricing.SubTotal != 100.0 {
		t.Errorf("Expected SubTotal 100.0, got %f", pricing.SubTotal)
	}

	// TaxAmount = 100 * 0.08 = 8.0
	if pricing.TaxAmount != 8.0 {
		t.Errorf("Expected TaxAmount 8.0, got %f", pricing.TaxAmount)
	}

	// TotalPrice = 100 + 8 = 108
	if pricing.TotalPrice != 108.0 {
		t.Errorf("Expected TotalPrice 108.0, got %f", pricing.TotalPrice)
	}
}

func TestCalculatePricing_FullDiscount(t *testing.T) {
	pricing := models.CalculatePricing(100.0, 100.0, 0.08)

	// SubTotal = 100 - 100 = 0
	if pricing.SubTotal != 0.0 {
		t.Errorf("Expected SubTotal 0.0, got %f", pricing.SubTotal)
	}

	// TaxAmount = 0 * 0.08 = 0
	if pricing.TaxAmount != 0.0 {
		t.Errorf("Expected TaxAmount 0.0, got %f", pricing.TaxAmount)
	}

	// TotalPrice = 0 + 0 = 0
	if pricing.TotalPrice != 0.0 {
		t.Errorf("Expected TotalPrice 0.0, got %f", pricing.TotalPrice)
	}
}

func TestNewPricingBreakdown_CustomCurrency(t *testing.T) {
	pricing := models.NewPricingBreakdown(100.0, 10.0, 0.20, "EUR")

	if pricing.Currency != "EUR" {
		t.Errorf("Expected Currency EUR, got %s", pricing.Currency)
	}

	if pricing.TaxRate != 0.20 {
		t.Errorf("Expected TaxRate 0.20, got %f", pricing.TaxRate)
	}

	// TaxAmount = 90 * 0.20 = 18.0
	if pricing.TaxAmount != 18.0 {
		t.Errorf("Expected TaxAmount 18.0, got %f", pricing.TaxAmount)
	}

	// TotalPrice = 90 + 18 = 108
	if pricing.TotalPrice != 108.0 {
		t.Errorf("Expected TotalPrice 108.0, got %f", pricing.TotalPrice)
	}
}

func TestPricingBreakdown_AddTip(t *testing.T) {
	pricing := models.CalculatePricing(100.0, 10.0, 0.08)

	// TotalPrice before tip = 97.2
	// Add $15 tip
	finalTotal := pricing.AddTip(15.0)

	// Expected: 97.2 + 15 = 112.2
	expected := 112.2
	if finalTotal != expected {
		t.Errorf("Expected final total %f, got %f", expected, finalTotal)
	}
}

func TestPricingBreakdown_GetSavings(t *testing.T) {
	pricing := models.CalculatePricing(100.0, 25.0, 0.08)

	savings := pricing.GetSavings()
	if savings != 25.0 {
		t.Errorf("Expected savings 25.0, got %f", savings)
	}
}

func TestPricingBreakdown_GetEffectivePrice(t *testing.T) {
	pricing := models.CalculatePricing(100.0, 25.0, 0.08)

	effectivePrice := pricing.GetEffectivePrice()
	// Effective price = ServicePrice - Discount = 100 - 25 = 75
	if effectivePrice != 75.0 {
		t.Errorf("Expected effective price 75.0, got %f", effectivePrice)
	}
}

// ========================================================================
// VALIDATION TESTS
// ========================================================================

func TestPricingBreakdown_Validate_Valid(t *testing.T) {
	pricing := models.CalculatePricing(100.0, 10.0, 0.08)

	err := pricing.Validate()
	if err != nil {
		t.Errorf("Expected valid pricing, got error: %v", err)
	}
}

func TestPricingBreakdown_Validate_NegativeServicePrice(t *testing.T) {
	pricing := &models.PricingBreakdown{
		ServicePrice: -100.0,
		TaxRate:      0.08,
	}

	err := pricing.Validate()
	if err == nil {
		t.Error("Expected error for negative service price")
	}
}

func TestPricingBreakdown_Validate_NegativeDiscount(t *testing.T) {
	pricing := &models.PricingBreakdown{
		ServicePrice:   100.0,
		DiscountAmount: -10.0,
		TaxRate:        0.08,
	}

	err := pricing.Validate()
	if err == nil {
		t.Error("Expected error for negative discount")
	}
}

func TestPricingBreakdown_Validate_DiscountExceedsPrice(t *testing.T) {
	pricing := &models.PricingBreakdown{
		ServicePrice:   100.0,
		DiscountAmount: 150.0,
		TaxRate:        0.08,
	}

	err := pricing.Validate()
	if err == nil {
		t.Error("Expected error for discount exceeding price")
	}
}

func TestPricingBreakdown_Validate_InvalidTaxRate(t *testing.T) {
	// Test negative tax rate
	pricing1 := &models.PricingBreakdown{
		ServicePrice: 100.0,
		TaxRate:      -0.08,
	}
	err := pricing1.Validate()
	if err == nil {
		t.Error("Expected error for negative tax rate")
	}

	// Test tax rate > 1
	pricing2 := &models.PricingBreakdown{
		ServicePrice: 100.0,
		TaxRate:      1.5,
	}
	err = pricing2.Validate()
	if err == nil {
		t.Error("Expected error for tax rate > 1")
	}
}

// ========================================================================
// EDGE CASE TESTS
// ========================================================================

func TestCalculatePricing_ZeroValues(t *testing.T) {
	pricing := models.CalculatePricing(0.0, 0.0, 0.0)

	if pricing.TotalPrice != 0.0 {
		t.Errorf("Expected TotalPrice 0.0, got %f", pricing.TotalPrice)
	}
}

func TestCalculatePricing_LargeValues(t *testing.T) {
	pricing := models.CalculatePricing(10000.0, 1000.0, 0.25)

	// SubTotal = 10000 - 1000 = 9000
	if pricing.SubTotal != 9000.0 {
		t.Errorf("Expected SubTotal 9000.0, got %f", pricing.SubTotal)
	}

	// TaxAmount = 9000 * 0.25 = 2250
	if pricing.TaxAmount != 2250.0 {
		t.Errorf("Expected TaxAmount 2250.0, got %f", pricing.TaxAmount)
	}

	// TotalPrice = 9000 + 2250 = 11250
	if pricing.TotalPrice != 11250.0 {
		t.Errorf("Expected TotalPrice 11250.0, got %f", pricing.TotalPrice)
	}
}

func TestCalculatePricing_SmallValues(t *testing.T) {
	pricing := models.CalculatePricing(5.0, 0.5, 0.08)

	// SubTotal = 5 - 0.5 = 4.5
	if pricing.SubTotal != 4.5 {
		t.Errorf("Expected SubTotal 4.5, got %f", pricing.SubTotal)
	}

	// TaxAmount = 4.5 * 0.08 = 0.36
	expectedTax := 0.36
	if pricing.TaxAmount != expectedTax {
		t.Errorf("Expected TaxAmount %f, got %f", expectedTax, pricing.TaxAmount)
	}
}

// ========================================================================
// COMPARISON WITH OLD IMPLEMENTATION
// ========================================================================

func TestPricingBreakdown_MatchesOldImplementation(t *testing.T) {
	// Old implementation returned: (servicePrice, discountAmount, taxAmount, totalPrice)
	servicePrice := 100.0
	discountAmount := 10.0
	taxRate := 0.08

	// Old calculation
	taxableAmount := servicePrice - discountAmount
	oldTaxAmount := taxableAmount * taxRate
	oldTotalPrice := taxableAmount + oldTaxAmount

	// New implementation
	pricing := models.CalculatePricing(servicePrice, discountAmount, taxRate)

	// Compare results
	if pricing.ServicePrice != servicePrice {
		t.Error("ServicePrice doesn't match old implementation")
	}

	if pricing.DiscountAmount != discountAmount {
		t.Error("DiscountAmount doesn't match old implementation")
	}

	if pricing.TaxAmount != oldTaxAmount {
		t.Errorf("TaxAmount doesn't match old implementation. Expected %f, got %f",
			oldTaxAmount, pricing.TaxAmount)
	}

	if pricing.TotalPrice != oldTotalPrice {
		t.Errorf("TotalPrice doesn't match old implementation. Expected %f, got %f",
			oldTotalPrice, pricing.TotalPrice)
	}
}

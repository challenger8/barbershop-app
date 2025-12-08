// internal/models/pricing.go
package models

import "fmt"

// ========================================================================
// PRICING MODELS - Replace Multiple Return Values
// ========================================================================
//
// Before: func calculateTotalPrice(...) (float64, float64, float64, float64)
// After:  func calculateTotalPrice(...) *PricingBreakdown
//
// Benefits:
//   - Self-documenting (field names vs. positional returns)
//   - Easy to extend (add new fields without breaking signatures)
//   - Type-safe (can't mix up return values)
//   - Better for JSON serialization
// ========================================================================

// PricingBreakdown represents a complete pricing calculation
type PricingBreakdown struct {
	ServicePrice   float64 `json:"service_price"`
	DiscountAmount float64 `json:"discount_amount"`
	TaxAmount      float64 `json:"tax_amount"`
	TaxRate        float64 `json:"tax_rate"`
	SubTotal       float64 `json:"sub_total"`      // ServicePrice - DiscountAmount
	TotalPrice     float64 `json:"total_price"`    // SubTotal + TaxAmount
	Currency       string  `json:"currency"`
}

// NewPricingBreakdown creates a new pricing breakdown with calculations
func NewPricingBreakdown(servicePrice, discountAmount float64, taxRate float64, currency string) *PricingBreakdown {
	subTotal := servicePrice - discountAmount
	taxAmount := subTotal * taxRate
	totalPrice := subTotal + taxAmount

	return &PricingBreakdown{
		ServicePrice:   servicePrice,
		DiscountAmount: discountAmount,
		TaxAmount:      taxAmount,
		TaxRate:        taxRate,
		SubTotal:       subTotal,
		TotalPrice:     totalPrice,
		Currency:       currency,
	}
}

// CalculatePricing is a convenience function for calculating pricing
func CalculatePricing(servicePrice, discountAmount, taxRate float64) *PricingBreakdown {
	return NewPricingBreakdown(servicePrice, discountAmount, taxRate, "USD")
}

// ========================================================================
// HELPER METHODS
// ========================================================================

// AddTip adds a tip amount to the total price
func (p *PricingBreakdown) AddTip(tipAmount float64) float64 {
	return p.TotalPrice + tipAmount
}

// GetSavings returns the total savings (discount + any promotions)
func (p *PricingBreakdown) GetSavings() float64 {
	return p.DiscountAmount
}

// GetEffectivePrice returns the price after all discounts
func (p *PricingBreakdown) GetEffectivePrice() float64 {
	return p.SubTotal
}

// ========================================================================
// VALIDATION
// ========================================================================

// Validate ensures the pricing breakdown is valid
func (p *PricingBreakdown) Validate() error {
	if p.ServicePrice < 0 {
		return fmt.Errorf("service price cannot be negative")
	}
	if p.DiscountAmount < 0 {
		return fmt.Errorf("discount amount cannot be negative")
	}
	if p.DiscountAmount > p.ServicePrice {
		return fmt.Errorf("discount cannot exceed service price")
	}
	if p.TaxRate < 0 || p.TaxRate > 1 {
		return fmt.Errorf("tax rate must be between 0 and 1")
	}
	return nil
}

// ========================================================================
// USAGE EXAMPLES
// ========================================================================
//
// Example 1: Simple calculation
//   pricing := models.CalculatePricing(100.0, 10.0, 0.08)
//   fmt.Printf("Total: $%.2f\n", pricing.TotalPrice) // Total: $97.20
//
// Example 2: With custom currency
//   pricing := models.NewPricingBreakdown(100.0, 10.0, 0.08, "EUR")
//
// Example 3: Add tip
//   finalTotal := pricing.AddTip(15.0) // $97.20 + $15.00 = $112.20
//
// Example 4: Get savings
//   saved := pricing.GetSavings() // $10.00
// ========================================================================
// tests/unit/validation/validator_test.go
package validation

import (
	"strings"
	"testing"

	"barber-booking-system/internal/validation"
)

// ========================================================================
// VALIDATION TESTS
// ========================================================================

func TestInitialize(t *testing.T) {
	validation.Initialize()
	validator := validation.GetValidator()

	if validator == nil {
		t.Fatal("Validator should not be nil after initialization")
	}
}

// ========================================================================
// EMAIL VALIDATION TESTS
// ========================================================================

func TestValidateEmail_Valid(t *testing.T) {
	validEmails := []string{
		"test@example.com",
		"user.name@example.co.uk",
		"user+tag@example.com",
		"user123@test-domain.com",
	}

	for _, email := range validEmails {
		err := validation.ValidateEmail(email)
		if err != nil {
			t.Errorf("Expected email %s to be valid, got error: %v", email, err)
		}
	}
}

func TestValidateEmail_Invalid(t *testing.T) {
	invalidEmails := []string{
		"",
		"not-an-email",
		"@example.com",
		"user@",
		"user @example.com",
		"user@.com",
	}

	for _, email := range invalidEmails {
		err := validation.ValidateEmail(email)
		if err == nil {
			t.Errorf("Expected email %s to be invalid", email)
		}
	}
}

// ========================================================================
// PHONE VALIDATION TESTS
// ========================================================================

func TestValidatePhone_Valid(t *testing.T) {
	validPhones := []string{
		"+14155552671",  // US number
		"+442071838750", // UK number
		"+81312345678",  // Japan number
		"+33123456789",  // France number
		"+61212345678",  // Australia number
	}

	for _, phone := range validPhones {
		err := validation.ValidatePhone(phone)
		if err != nil {
			t.Errorf("Expected phone %s to be valid, got error: %v", phone, err)
		}
	}
}

func TestValidatePhone_Invalid(t *testing.T) {
	invalidPhones := []string{
		"",               // Empty
		"123",            // Too short
		"+1234",          // Too short
		"abc123",         // Contains letters
		"(555) 123-4567", // Not E.164 format
		"555-1234",       // No country code
		"+012345678",     // Starts with 0
		"14155552671",    // Missing +
	}

	for _, phone := range invalidPhones {
		err := validation.ValidatePhone(phone)
		if err == nil {
			t.Errorf("Expected phone %s to be invalid", phone)
		}
	}
}

// ========================================================================
// STRUCT VALIDATION TESTS
// ========================================================================

func TestValidateStruct_AllValid(t *testing.T) {
	type TestStruct struct {
		Name  string `validate:"required,min=2,max=50"`
		Email string `validate:"required,email"`
		Age   int    `validate:"required,gte=0,lte=150"`
	}

	s := &TestStruct{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	err := validation.ValidateStruct(s)
	if err != nil {
		t.Errorf("Expected valid struct, got error: %v", err)
	}
}

func TestValidateStruct_RequiredFieldMissing(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"required"`
	}

	s := &TestStruct{
		Name: "",
	}

	err := validation.ValidateStruct(s)
	if err == nil {
		t.Error("Expected error for missing required field")
	}

	if !strings.Contains(err.Error(), "required") {
		t.Errorf("Expected error message to contain 'required', got: %v", err)
	}
}

func TestValidateStruct_MinLength(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"required,min=5"`
	}

	s := &TestStruct{
		Name: "abc", // Only 3 characters
	}

	err := validation.ValidateStruct(s)
	if err == nil {
		t.Error("Expected error for min length violation")
	}

	if !strings.Contains(err.Error(), "at least") {
		t.Errorf("Expected error message about min length, got: %v", err)
	}
}

func TestValidateStruct_MaxLength(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"required,max=10"`
	}

	s := &TestStruct{
		Name: "This is a very long name", // More than 10 characters
	}

	err := validation.ValidateStruct(s)
	if err == nil {
		t.Error("Expected error for max length violation")
	}

	if !strings.Contains(err.Error(), "at most") {
		t.Errorf("Expected error message about max length, got: %v", err)
	}
}

func TestValidateStruct_EmailFormat(t *testing.T) {
	type TestStruct struct {
		Email string `validate:"required,email"`
	}

	s := &TestStruct{
		Email: "not-an-email",
	}

	err := validation.ValidateStruct(s)
	if err == nil {
		t.Error("Expected error for invalid email format")
	}

	if !strings.Contains(err.Error(), "email") {
		t.Errorf("Expected error message about email format, got: %v", err)
	}
}

func TestValidateStruct_NumericRange(t *testing.T) {
	type TestStruct struct {
		Age int `validate:"required,gte=0,lte=150"`
	}

	// Test below minimum
	s1 := &TestStruct{Age: -1}
	err := validation.ValidateStruct(s1)
	if err == nil {
		t.Error("Expected error for value below minimum")
	}

	// Test above maximum
	s2 := &TestStruct{Age: 200}
	err = validation.ValidateStruct(s2)
	if err == nil {
		t.Error("Expected error for value above maximum")
	}

	// Test valid value
	s3 := &TestStruct{Age: 30}
	err = validation.ValidateStruct(s3)
	if err != nil {
		t.Errorf("Expected valid value, got error: %v", err)
	}
}

func TestValidateStruct_OneOf(t *testing.T) {
	type TestStruct struct {
		Status string `validate:"required,oneof=active inactive pending"`
	}

	// Valid value
	s1 := &TestStruct{Status: "active"}
	err := validation.ValidateStruct(s1)
	if err != nil {
		t.Errorf("Expected valid oneof value, got error: %v", err)
	}

	// Invalid value
	s2 := &TestStruct{Status: "invalid"}
	err = validation.ValidateStruct(s2)
	if err == nil {
		t.Error("Expected error for invalid oneof value")
	}

	if !strings.Contains(err.Error(), "one of") {
		t.Errorf("Expected error message about oneof, got: %v", err)
	}
}

func TestValidateStruct_NestedStruct(t *testing.T) {
	type Address struct {
		Street string `validate:"required,min=5"`
		City   string `validate:"required,min=2"`
	}

	type Person struct {
		Name    string  `validate:"required,min=2"`
		Address Address `validate:"required"`
	}

	// Valid nested struct
	p1 := &Person{
		Name: "John",
		Address: Address{
			Street: "123 Main St",
			City:   "New York",
		},
	}
	err := validation.ValidateStruct(p1)
	if err != nil {
		t.Errorf("Expected valid nested struct, got error: %v", err)
	}

	// Invalid nested struct (city too short)
	p2 := &Person{
		Name: "John",
		Address: Address{
			Street: "123 Main St",
			City:   "N", // Too short
		},
	}
	err = validation.ValidateStruct(p2)
	if err == nil {
		t.Error("Expected error for invalid nested struct")
	}
}

// ========================================================================
// CUSTOM VALIDATOR TESTS
// ========================================================================

func TestCustomValidator_BookingStatus(t *testing.T) {
	type TestStruct struct {
		Status string `validate:"required,booking_status"`
	}

	validStatuses := []string{
		"pending",
		"confirmed",
		"in_progress",
		"completed",
		"cancelled",
		"no_show",
	}

	for _, status := range validStatuses {
		s := &TestStruct{Status: status}
		err := validation.ValidateStruct(s)
		if err != nil {
			t.Errorf("Expected status %s to be valid, got error: %v", status, err)
		}
	}

	// Invalid status
	s := &TestStruct{Status: "invalid_status"}
	err := validation.ValidateStruct(s)
	if err == nil {
		t.Error("Expected error for invalid booking status")
	}
}

func TestCustomValidator_BarberStatus(t *testing.T) {
	type TestStruct struct {
		Status string `validate:"required,barber_status"`
	}

	validStatuses := []string{
		"pending",
		"active",
		"inactive",
		"suspended",
		"rejected",
	}

	for _, status := range validStatuses {
		s := &TestStruct{Status: status}
		err := validation.ValidateStruct(s)
		if err != nil {
			t.Errorf("Expected status %s to be valid, got error: %v", status, err)
		}
	}

	// Invalid status
	s := &TestStruct{Status: "invalid_status"}
	err := validation.ValidateStruct(s)
	if err == nil {
		t.Error("Expected error for invalid barber status")
	}
}

// ========================================================================
// ERROR MESSAGE FORMATTING TESTS
// ========================================================================

func TestErrorMessageFormatting(t *testing.T) {
	type TestStruct struct {
		Name  string `validate:"required"`
		Email string `validate:"required,email"`
		Age   int    `validate:"required,gte=18"`
	}

	// Multiple errors
	s := &TestStruct{
		Name:  "",
		Email: "invalid",
		Age:   10,
	}

	err := validation.ValidateStruct(s)
	if err == nil {
		t.Fatal("Expected validation errors")
	}

	errorMsg := err.Error()

	// Check that error message contains multiple field errors
	if !strings.Contains(errorMsg, "name") {
		t.Error("Error message should mention 'name' field")
	}
	if !strings.Contains(errorMsg, "email") {
		t.Error("Error message should mention 'email' field")
	}
	if !strings.Contains(errorMsg, "age") {
		t.Error("Error message should mention 'age' field")
	}
}

func TestToSnakeCase(t *testing.T) {
	// This tests the internal snake_case conversion
	type TestStruct struct {
		CustomerName string `validate:"required"`
	}

	s := &TestStruct{CustomerName: ""}
	err := validation.ValidateStruct(s)

	if err == nil {
		t.Fatal("Expected validation error")
	}

	// Should convert CustomerName to customer_name in error message
	if !strings.Contains(err.Error(), "customer_name") {
		t.Errorf("Expected snake_case field name in error, got: %v", err)
	}
}

// ========================================================================
// OPTIONAL FIELD TESTS
// ========================================================================

func TestValidateStruct_OptionalFields(t *testing.T) {
	type TestStruct struct {
		RequiredField string  `validate:"required"`
		OptionalField *string `validate:"omitempty,min=5"`
	}

	// Optional field not provided - should pass
	s1 := &TestStruct{
		RequiredField: "test",
		OptionalField: nil,
	}
	err := validation.ValidateStruct(s1)
	if err != nil {
		t.Errorf("Expected valid struct with nil optional field, got error: %v", err)
	}

	// Optional field provided but too short - should fail
	shortValue := "abc"
	s2 := &TestStruct{
		RequiredField: "test",
		OptionalField: &shortValue,
	}
	err = validation.ValidateStruct(s2)
	if err == nil {
		t.Error("Expected error for optional field with invalid value")
	}

	// Optional field provided and valid - should pass
	validValue := "valid value"
	s3 := &TestStruct{
		RequiredField: "test",
		OptionalField: &validValue,
	}
	err = validation.ValidateStruct(s3)
	if err != nil {
		t.Errorf("Expected valid struct with valid optional field, got error: %v", err)
	}
}

// ========================================================================
// ARRAY VALIDATION TESTS
// ========================================================================

func TestValidateStruct_ArrayElements(t *testing.T) {
	type TestStruct struct {
		Tags []string `validate:"required,dive,min=2,max=20"`
	}

	// Valid array
	s1 := &TestStruct{
		Tags: []string{"tag1", "tag2", "tag3"},
	}
	err := validation.ValidateStruct(s1)
	if err != nil {
		t.Errorf("Expected valid array, got error: %v", err)
	}

	// Array with element too short
	s2 := &TestStruct{
		Tags: []string{"tag1", "a", "tag3"},
	}
	err = validation.ValidateStruct(s2)
	if err == nil {
		t.Error("Expected error for array element too short")
	}

	// Array with element too long
	longTag := strings.Repeat("a", 25)
	s3 := &TestStruct{
		Tags: []string{"tag1", longTag, "tag3"},
	}
	err = validation.ValidateStruct(s3)
	if err == nil {
		t.Error("Expected error for array element too long")
	}
}

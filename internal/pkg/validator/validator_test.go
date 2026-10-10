package validator_test

import (
	"testing"

	"github.com/eveeze/warung-backend/internal/pkg/validator"
)

func TestValidator(t *testing.T) {
	v := validator.New()
	if v.HasErrors() {
		t.Fatalf("expected new validator to have no errors")
	}

	// Required
	if !v.Required("name", "John", "name is required") {
		t.Errorf("expected true")
	}
	if v.Required("empty_name", "  ", "empty_name is required") {
		t.Errorf("expected false")
	}

	// RequiredInt & RequiredInt64
	if !v.RequiredInt("age", 25, "age is required") {
		t.Errorf("expected true")
	}
	if v.RequiredInt("zero_age", 0, "zero_age is required") {
		t.Errorf("expected false")
	}
	if !v.RequiredInt64("price", 1000, "price is required") {
		t.Errorf("expected true")
	}
	if v.RequiredInt64("zero_price", 0, "zero_price is required") {
		t.Errorf("expected false")
	}

	// MinLength & MaxLength
	if !v.MinLength("pass", "secret", 6, "min 6") {
		t.Errorf("expected true")
	}
	if v.MinLength("pass_short", "abc", 6, "min 6") {
		t.Errorf("expected false")
	}
	if !v.MaxLength("short", "abc", 5, "max 5") {
		t.Errorf("expected true")
	}
	if v.MaxLength("long", "abcdef", 5, "max 5") {
		t.Errorf("expected false")
	}

	// Min, Max, MinInt64
	if !v.Min("qty", 10, 5, "min 5") {
		t.Errorf("expected true")
	}
	if v.Min("qty_small", 3, 5, "min 5") {
		t.Errorf("expected false")
	}
	if !v.Max("qty", 5, 10, "max 10") {
		t.Errorf("expected true")
	}
	if v.Max("qty_large", 15, 10, "max 10") {
		t.Errorf("expected false")
	}
	if !v.MinInt64("amt", 100, 50, "min 50") {
		t.Errorf("expected true")
	}
	if v.MinInt64("amt_small", 30, 50, "min 50") {
		t.Errorf("expected false")
	}

	// Positive & NonNegative
	if !v.Positive("pos", 10, "must be positive") {
		t.Errorf("expected true")
	}
	if v.Positive("non_pos", 0, "must be positive") {
		t.Errorf("expected false")
	}
	if !v.NonNegative("zero", 0, "must be non negative") {
		t.Errorf("expected true")
	}
	if v.NonNegative("neg", -1, "must be non negative") {
		t.Errorf("expected false")
	}

	// Email
	if !v.Email("email", "test@example.com", "invalid email") {
		t.Errorf("expected true")
	}
	if v.Email("bad_email", "invalid-email", "invalid email") {
		t.Errorf("expected false")
	}

	// Phone
	if !v.Phone("empty_phone", "", "phone optional") {
		t.Errorf("expected true")
	}
	if !v.Phone("phone", "081234567890", "invalid phone") {
		t.Errorf("expected true")
	}
	if !v.Phone("phone_intl", "+628123456789", "invalid phone") {
		t.Errorf("expected true")
	}
	if v.Phone("bad_phone", "12345", "invalid phone") {
		t.Errorf("expected false")
	}

	// InSlice
	allowed := []string{"cash", "qris", "kasbon"}
	if !v.InSlice("method", "cash", allowed, "invalid method") {
		t.Errorf("expected true")
	}
	if v.InSlice("method_bad", "bitcoin", allowed, "invalid method") {
		t.Errorf("expected false")
	}

	// UUID
	validUUID := "c56a4180-65aa-42ec-a945-5fd21dec0538"
	if !v.UUID("uuid", validUUID, "invalid uuid") {
		t.Errorf("expected true")
	}
	if v.UUID("bad_uuid", "not-a-uuid", "invalid uuid") {
		t.Errorf("expected false")
	}

	// Barcode
	if !v.Barcode("empty_barcode", "", "optional") {
		t.Errorf("expected true")
	}
	if !v.Barcode("barcode", "8991234567890", "invalid barcode") {
		t.Errorf("expected true")
	}
	if v.Barcode("bad_barcode", "!!", "invalid barcode") {
		t.Errorf("expected false")
	}

	// Custom
	if !v.Custom("custom_ok", true, "custom fail") {
		t.Errorf("expected true")
	}
	if v.Custom("custom_fail", false, "custom fail") {
		t.Errorf("expected false")
	}

	// ValidationErrors methods
	errs := v.Errors()
	if !errs.HasErrors() {
		t.Errorf("expected errors")
	}
	errStr := errs.Error()
	if errStr == "" {
		t.Errorf("expected non-empty error string")
	}
}

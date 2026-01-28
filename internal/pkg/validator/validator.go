package validator

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// ValidationErrors holds validation error messages
type ValidationErrors map[string]string

// Error implements the error interface
func (v ValidationErrors) Error() string {
	var sb strings.Builder
	for field, msg := range v {
		sb.WriteString(field)
		sb.WriteString(": ")
		sb.WriteString(msg)
		sb.WriteString("; ")
	}
	return strings.TrimSuffix(sb.String(), "; ")
}

// HasErrors returns true if there are validation errors
func (v ValidationErrors) HasErrors() bool {
	return len(v) > 0
}

// Add adds an error message for a field
func (v ValidationErrors) Add(field, message string) {
	v[field] = message
}

// Validator provides validation methods
type Validator struct {
	errors ValidationErrors
}

// New creates a new Validator
func New() *Validator {
	return &Validator{
		errors: make(ValidationErrors),
	}
}

// Errors returns the validation errors
func (v *Validator) Errors() ValidationErrors {
	return v.errors
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return v.errors.HasErrors()
}

// Required checks if a string is not empty
func (v *Validator) Required(field, value, message string) bool {
	if strings.TrimSpace(value) == "" {
		v.errors.Add(field, message)
		return false
	}
	return true
}

// RequiredInt checks if an integer is not zero
func (v *Validator) RequiredInt(field string, value int, message string) bool {
	if value == 0 {
		v.errors.Add(field, message)
		return false
	}
	return true
}

// RequiredInt64 checks if an int64 is not zero
func (v *Validator) RequiredInt64(field string, value int64, message string) bool {
	if value == 0 {
		v.errors.Add(field, message)
		return false
	}
	return true
}

// MinLength checks if a string has minimum length
func (v *Validator) MinLength(field, value string, min int, message string) bool {
	if utf8.RuneCountInString(value) < min {
		v.errors.Add(field, message)
		return false
	}
	return true
}

// MaxLength checks if a string doesn't exceed maximum length
func (v *Validator) MaxLength(field, value string, max int, message string) bool {
	if utf8.RuneCountInString(value) > max {
		v.errors.Add(field, message)
		return false
	}
	return true
}

// Min checks if an integer is at least the minimum value
func (v *Validator) Min(field string, value, min int, message string) bool {
	if value < min {
		v.errors.Add(field, message)
		return false
	}
	return true
}

// Max checks if an integer doesn't exceed the maximum value
func (v *Validator) Max(field string, value, max int, message string) bool {
	if value > max {
		v.errors.Add(field, message)
		return false
	}
	return true
}

// MinInt64 checks if an int64 is at least the minimum value
func (v *Validator) MinInt64(field string, value, min int64, message string) bool {
	if value < min {
		v.errors.Add(field, message)
		return false
	}
	return true
}

// Positive checks if a value is positive (greater than 0)
func (v *Validator) Positive(field string, value int64, message string) bool {
	if value <= 0 {
		v.errors.Add(field, message)
		return false
	}
	return true
}

// NonNegative checks if a value is non-negative (>= 0)
func (v *Validator) NonNegative(field string, value int64, message string) bool {
	if value < 0 {
		v.errors.Add(field, message)
		return false
	}
	return true
}

// Email validates an email address format
func (v *Validator) Email(field, value, message string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(value) {
		v.errors.Add(field, message)
		return false
	}
	return true
}

// Phone validates a phone number format (Indonesian format)
func (v *Validator) Phone(field, value, message string) bool {
	if value == "" {
		return true // Phone is optional, use Required for mandatory check
	}
	// Indonesian phone: 08xxx, +628xxx, 628xxx
	phoneRegex := regexp.MustCompile(`^(\+62|62|0)[0-9]{9,13}$`)
	cleaned := strings.ReplaceAll(value, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	if !phoneRegex.MatchString(cleaned) {
		v.errors.Add(field, message)
		return false
	}
	return true
}

// InSlice checks if a value is in the allowed list
func (v *Validator) InSlice(field, value string, allowed []string, message string) bool {
	for _, a := range allowed {
		if value == a {
			return true
		}
	}
	v.errors.Add(field, message)
	return false
}

// UUID validates a UUID v4 format
func (v *Validator) UUID(field, value, message string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
	if !uuidRegex.MatchString(value) {
		v.errors.Add(field, message)
		return false
	}
	return true
}

// Barcode validates barcode format (EAN-13, UPC-A, etc.)
func (v *Validator) Barcode(field, value, message string) bool {
	if value == "" {
		return true // Barcode is optional
	}
	// Allow alphanumeric barcodes, 4-50 characters
	barcodeRegex := regexp.MustCompile(`^[A-Za-z0-9]{4,50}$`)
	if !barcodeRegex.MatchString(value) {
		v.errors.Add(field, message)
		return false
	}
	return true
}

// Custom adds a custom validation with condition
func (v *Validator) Custom(field string, condition bool, message string) bool {
	if !condition {
		v.errors.Add(field, message)
		return false
	}
	return true
}

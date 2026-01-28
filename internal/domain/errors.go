package domain

import "errors"

// Common errors
var (
	// ErrNotFound is returned when a resource is not found
	ErrNotFound = errors.New("resource not found")
	
	// ErrAlreadyExists is returned when a resource already exists
	ErrAlreadyExists = errors.New("resource already exists")
	
	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")
	
	// ErrInsufficientStock is returned when there's not enough stock
	ErrInsufficientStock = errors.New("insufficient stock")
	
	// ErrCreditLimitExceeded is returned when credit limit is exceeded
	ErrCreditLimitExceeded = errors.New("credit limit exceeded")
	
	// ErrTransactionCancelled is returned when transaction is already cancelled
	ErrTransactionCancelled = errors.New("transaction is already cancelled")
	
	// ErrTransactionCompleted is returned when trying to modify completed transaction
	ErrTransactionCompleted = errors.New("cannot modify completed transaction")
	
	// ErrInvalidPaymentAmount is returned when payment amount is invalid
	ErrInvalidPaymentAmount = errors.New("invalid payment amount")
	
	// ErrEmptyCart is returned when cart is empty
	ErrEmptyCart = errors.New("cart is empty")
	
	// ErrProductInactive is returned when product is inactive
	ErrProductInactive = errors.New("product is inactive")
	
	// ErrCustomerInactive is returned when customer is inactive
	ErrCustomerInactive = errors.New("customer is inactive")
)

package domain

import (
	"time"

	"github.com/google/uuid"
)

// KasbonType represents the type of kasbon record
type KasbonType string

const (
	KasbonTypeDebt    KasbonType = "debt"    // hutang baru
	KasbonTypePayment KasbonType = "payment" // pembayaran hutang
)

// KasbonRecord represents a debt or payment record
type KasbonRecord struct {
	ID            uuid.UUID  `json:"id"`
	CustomerID    uuid.UUID  `json:"customer_id"`
	TransactionID *uuid.UUID `json:"transaction_id,omitempty"`
	Type          KasbonType `json:"type"`
	Amount        int64      `json:"amount"`        // jumlah hutang atau pembayaran
	BalanceBefore int64      `json:"balance_before"`
	BalanceAfter  int64      `json:"balance_after"`
	Notes         *string    `json:"notes,omitempty"`
	CreatedBy     *string    `json:"created_by,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`

	// Relations
	Customer    *Customer    `json:"customer,omitempty"`
	Transaction *Transaction `json:"transaction,omitempty"`
}

// KasbonPaymentInput is the input for recording a kasbon payment
type KasbonPaymentInput struct {
	CustomerID uuid.UUID `json:"customer_id"`
	Amount     int64     `json:"amount"`
	Notes      *string   `json:"notes,omitempty"`
	CreatedBy  *string   `json:"created_by,omitempty"`
}

// KasbonFilter is the filter for listing kasbon records
type KasbonFilter struct {
	CustomerID *uuid.UUID  `json:"customer_id,omitempty"`
	Type       *KasbonType `json:"type,omitempty"`
	DateFrom   *time.Time  `json:"date_from,omitempty"`
	DateTo     *time.Time  `json:"date_to,omitempty"`
	Page       int         `json:"page,omitempty"`
	PerPage    int         `json:"per_page,omitempty"`
}

// KasbonSummary is a summary of kasbon for a customer
type KasbonSummary struct {
	CustomerID       uuid.UUID `json:"customer_id"`
	CustomerName     string    `json:"customer_name"`
	TotalDebt        int64     `json:"total_debt"`
	TotalPayment     int64     `json:"total_payment"`
	CurrentBalance   int64     `json:"current_balance"`
	CreditLimit      int64     `json:"credit_limit"`
	RemainingCredit  int64     `json:"remaining_credit"`
	LastTransactionAt *time.Time `json:"last_transaction_at,omitempty"`
}

// KasbonReport is a report of all outstanding kasbon
type KasbonReport struct {
	TotalOutstanding   int64           `json:"total_outstanding"`
	TotalCustomers     int             `json:"total_customers"`
	CustomersWithDebt  int             `json:"customers_with_debt"`
	Summaries          []KasbonSummary `json:"summaries"`
}

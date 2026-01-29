package domain

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending       PaymentStatus = "pending"
	PaymentStatusSettlement    PaymentStatus = "settlement"
	PaymentStatusCapture       PaymentStatus = "capture"
	PaymentStatusDeny          PaymentStatus = "deny"
	PaymentStatusCancel        PaymentStatus = "cancel"
	PaymentStatusExpire        PaymentStatus = "expire"
	PaymentStatusFailure       PaymentStatus = "failure"
	PaymentStatusRefund        PaymentStatus = "refund"
	PaymentStatusPartialRefund PaymentStatus = "partial_refund"
)

// IsSuccess returns true if the payment is considered successful
func (s PaymentStatus) IsSuccess() bool {
	return s == PaymentStatusSettlement || s == PaymentStatusCapture
}

// IsFinal returns true if the payment status is final (no more changes expected)
func (s PaymentStatus) IsFinal() bool {
	switch s {
	case PaymentStatusSettlement, PaymentStatusCapture, PaymentStatusDeny,
		PaymentStatusCancel, PaymentStatusExpire, PaymentStatusFailure,
		PaymentStatusRefund, PaymentStatusPartialRefund:
		return true
	}
	return false
}

// PaymentRecord represents a payment gateway record
type PaymentRecord struct {
	ID            uuid.UUID       `json:"id"`
	TransactionID uuid.UUID       `json:"transaction_id"`
	OrderID       string          `json:"order_id"`
	SnapToken     *string         `json:"snap_token,omitempty"`
	RedirectURL   *string         `json:"redirect_url,omitempty"`
	PaymentType   *string         `json:"payment_type,omitempty"`
	GrossAmount   int64           `json:"gross_amount"`
	Currency      string          `json:"currency"`
	Status        PaymentStatus   `json:"status"`
	FraudStatus   *string         `json:"fraud_status,omitempty"`
	MidtransResp  json.RawMessage `json:"midtrans_response,omitempty"`
	PaidAt        *time.Time      `json:"paid_at,omitempty"`
	ExpiredAt     *time.Time      `json:"expired_at,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`

	// Relations
	Transaction *Transaction `json:"transaction,omitempty"`
}

// PaymentRepository defines the interface for payment data access
type PaymentRepository interface {
	Create(ctx context.Context, tx *sql.Tx, record *PaymentRecord) error
	GetByID(ctx context.Context, id uuid.UUID) (*PaymentRecord, error)
	GetByOrderID(ctx context.Context, orderID string) (*PaymentRecord, error)
	GetByTransactionID(ctx context.Context, transactionID uuid.UUID) (*PaymentRecord, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status PaymentStatus, response json.RawMessage) error
	UpdateSnapToken(ctx context.Context, id uuid.UUID, snapToken, redirectURL string) error
}

// SnapTokenRequest represents the request to generate a Snap token
type SnapTokenRequest struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	GrossAmount   int64     `json:"gross_amount"`
	CustomerName  *string   `json:"customer_name,omitempty"`
	CustomerEmail *string   `json:"customer_email,omitempty"`
	CustomerPhone *string   `json:"customer_phone,omitempty"`
	ItemDetails   []SnapItemDetail `json:"item_details,omitempty"`
}

// SnapItemDetail represents an item in the Snap request
type SnapItemDetail struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Price    int64  `json:"price"`
	Quantity int    `json:"quantity"`
}

// SnapTokenResponse represents the response from Snap token generation
type SnapTokenResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
	OrderID     string `json:"order_id"`
}

// MidtransNotification represents the webhook payload from Midtrans
type MidtransNotification struct {
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	TransactionID     string `json:"transaction_id"`
	StatusMessage     string `json:"status_message"`
	StatusCode        string `json:"status_code"`
	SignatureKey      string `json:"signature_key"`
	SettlementTime    string `json:"settlement_time,omitempty"`
	PaymentType       string `json:"payment_type"`
	OrderID           string `json:"order_id"`
	MerchantID        string `json:"merchant_id"`
	GrossAmount       string `json:"gross_amount"`
	FraudStatus       string `json:"fraud_status,omitempty"`
	Currency          string `json:"currency"`
}

// ManualVerifyRequest represents a request to manually verify a payment
type ManualVerifyRequest struct {
	PaymentID uuid.UUID `json:"payment_id"`
	Notes     *string   `json:"notes,omitempty"`
}

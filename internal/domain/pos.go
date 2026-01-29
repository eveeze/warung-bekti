package domain

import (
	"time"

	"github.com/google/uuid"
)

// -- Held Carts --

type HeldCartStatus string

const (
	HeldCartStatusHeld      HeldCartStatus = "held"
	HeldCartStatusResumed   HeldCartStatus = "resumed"
	HeldCartStatusDiscarded HeldCartStatus = "discarded"
)

type HeldCart struct {
	ID           uuid.UUID      `json:"id"`
	HoldCode     string         `json:"hold_code"`
	CustomerID   *uuid.UUID     `json:"customer_id,omitempty"`
	CustomerName *string        `json:"customer_name,omitempty"`
	Status       HeldCartStatus `json:"status"`
	Subtotal     int64          `json:"subtotal"`
	Notes        *string        `json:"notes,omitempty"`
	HeldBy       *string        `json:"held_by,omitempty"`
	ResumedBy    *string        `json:"resumed_by,omitempty"`
	HeldAt       time.Time      `json:"held_at"`
	ResumedAt    *time.Time     `json:"resumed_at,omitempty"`
	ExpiresAt    *time.Time     `json:"expires_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`

	Items []HeldCartItem `json:"items,omitempty"`
}

type HeldCartItem struct {
	ID             uuid.UUID `json:"id"`
	CartID         uuid.UUID `json:"cart_id"`
	ProductID      uuid.UUID `json:"product_id"`
	ProductName    string    `json:"product_name"`
	ProductBarcode *string   `json:"product_barcode,omitempty"`
	Quantity       int       `json:"quantity"`
	Unit           string    `json:"unit"`
	UnitPrice      int64     `json:"unit_price"`
	Subtotal       int64     `json:"subtotal"`
	Notes          *string   `json:"notes,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type HoldCartInput struct {
	CustomerID *uuid.UUID       `json:"customer_id,omitempty"`
	Items      []HoldCartItemInput `json:"items"`
	Notes      *string          `json:"notes,omitempty"`
	HeldBy     string           `json:"held_by"`
}

type HoldCartItemInput struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Notes     *string   `json:"notes,omitempty"`
}

// -- Refunds --

type RefundStatus string

const (
	RefundStatusPending   RefundStatus = "pending"
	RefundStatusApproved  RefundStatus = "approved"
	RefundStatusRejected  RefundStatus = "rejected"
	RefundStatusCompleted RefundStatus = "completed"
)

type RefundRecord struct {
	ID                uuid.UUID    `json:"id"`
	RefundNumber      string       `json:"refund_number"`
	TransactionID     uuid.UUID    `json:"transaction_id"`
	CustomerID        *uuid.UUID   `json:"customer_id,omitempty"`
	TotalRefundAmount int64        `json:"total_refund_amount"`
	RefundMethod      string       `json:"refund_method"`
	Status            RefundStatus `json:"status"`
	Reason            string       `json:"reason"`
	Notes             *string      `json:"notes,omitempty"`
	RequestedBy       *string      `json:"requested_by,omitempty"`
	ApprovedBy        *string      `json:"approved_by,omitempty"`
	CompletedAt       *time.Time   `json:"completed_at,omitempty"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`

	Items []RefundItem `json:"items,omitempty"`
	Transaction *Transaction `json:"transaction,omitempty"`
}

type RefundItem struct {
	ID                uuid.UUID `json:"id"`
	RefundID          uuid.UUID `json:"refund_id"`
	TransactionItemID uuid.UUID `json:"transaction_item_id"`
	ProductID         uuid.UUID `json:"product_id"`
	ProductName       string    `json:"product_name"`
	Quantity          int       `json:"quantity"`
	UnitPrice         int64     `json:"unit_price"`
	RefundAmount      int64     `json:"refund_amount"`
	Reason            *string   `json:"reason,omitempty"`
	Restock           bool      `json:"restock"`
	CreatedAt         time.Time `json:"created_at"`
}

type CreateRefundInput struct {
	TransactionID uuid.UUID         `json:"transaction_id"`
	RefundMethod  string            `json:"refund_method"`
	Reason        string            `json:"reason"`
	Notes         *string           `json:"notes,omitempty"`
	RequestedBy   string            `json:"requested_by"`
	Items         []RefundItemInput `json:"items"`
}

type RefundItemInput struct {
	TransactionItemID uuid.UUID `json:"transaction_item_id"`
	Quantity          int       `json:"quantity"`
	Reason            *string   `json:"reason,omitempty"`
	Restock           bool      `json:"restock"`
}

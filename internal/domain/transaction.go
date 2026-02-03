package domain

import (
	"time"

	"github.com/google/uuid"
)

// PaymentMethod represents the payment method for a transaction
type PaymentMethod string

const (
	PaymentMethodCash     PaymentMethod = "cash"
	PaymentMethodKasbon   PaymentMethod = "kasbon"
	PaymentMethodTransfer PaymentMethod = "transfer"
	PaymentMethodQRIS     PaymentMethod = "qris"
	PaymentMethodMixed    PaymentMethod = "mixed"
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusCancelled TransactionStatus = "cancelled"
	TransactionStatusRefunded  TransactionStatus = "refunded"
)

// Transaction represents a sales transaction
type Transaction struct {
	ID             uuid.UUID          `json:"id"`
	InvoiceNumber  string             `json:"invoice_number"`
	CustomerID     *uuid.UUID         `json:"customer_id,omitempty"`
	Subtotal       int64              `json:"subtotal"`        // total sebelum diskon & pajak
	DiscountAmount int64              `json:"discount_amount"` // total diskon
	TaxAmount      int64              `json:"tax_amount"`      // total pajak
	TotalAmount    int64              `json:"total_amount"`    // total akhir
	PaymentMethod  PaymentMethod      `json:"payment_method"`
	AmountPaid     int64              `json:"amount_paid"`
	ChangeAmount   int64              `json:"change_amount"`
	Status         TransactionStatus  `json:"status"`
	Notes          *string            `json:"notes,omitempty"`
	CashierName    *string            `json:"cashier_name,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`

	// Relations (populated when needed)
	Customer *Customer          `json:"customer,omitempty"`
	Items    []TransactionItem  `json:"items,omitempty"`
}

// TransactionItem represents an item in a transaction
type TransactionItem struct {
	ID              uuid.UUID  `json:"id"`
	TransactionID   uuid.UUID  `json:"transaction_id"`
	ProductID       uuid.UUID  `json:"product_id"`
	ProductName     string     `json:"product_name"`
	ProductBarcode  *string    `json:"product_barcode,omitempty"`
	Quantity        int        `json:"quantity"`
	Unit            string     `json:"unit"`
	UnitPrice       int64      `json:"unit_price"`
	CostPrice       int64      `json:"cost_price"`
	Subtotal        int64      `json:"subtotal"`        // quantity * unit_price
	DiscountAmount  int64      `json:"discount_amount"`
	TotalAmount     int64      `json:"total_amount"`    // subtotal - discount
	PricingTierID   *uuid.UUID `json:"pricing_tier_id,omitempty"`
	PricingTierName *string    `json:"pricing_tier_name,omitempty"`
	Notes           *string    `json:"notes,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`

	// Relations (populated when needed)
	Product *Product `json:"product,omitempty"`
}

// Profit calculates the profit for this item
func (ti *TransactionItem) Profit() int64 {
	return ti.TotalAmount - (ti.CostPrice * int64(ti.Quantity))
}

// TransactionCreateInput is the input for creating a transaction
type TransactionCreateInput struct {
	CustomerID     *uuid.UUID              `json:"customer_id,omitempty"`
	Items          []TransactionItemInput  `json:"items"`
	DiscountAmount *int64                  `json:"discount_amount,omitempty"`
	TaxAmount      *int64                  `json:"tax_amount,omitempty"`
	PaymentMethod  PaymentMethod           `json:"payment_method"`
	AmountPaid     int64                   `json:"amount_paid"`
	Notes          *string                 `json:"notes,omitempty"`
	CashierName    *string                 `json:"cashier_name,omitempty"`
}

// TransactionItemInput is the input for a transaction item
type TransactionItemInput struct {
	ProductID      uuid.UUID `json:"product_id"`
	Quantity       int       `json:"quantity"`
	DiscountAmount *int64    `json:"discount_amount,omitempty"`
	Notes          *string   `json:"notes,omitempty"`
}

// CartCalculateInput is the input for calculating cart totals (preview)
type CartCalculateInput struct {
	Items []CartItem `json:"items"`
}

// CartItem represents an item in the cart for calculation
type CartItem struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

// CartCalculateResult is the result of cart calculation
type CartCalculateResult struct {
	Items    []CartItemResult `json:"items"`
	Subtotal int64            `json:"subtotal"`
}

// CartItemResult is the result for each item in cart calculation
type CartItemResult struct {
	ProductID     uuid.UUID `json:"product_id"`
	ProductName   string    `json:"product_name"`
	Quantity      int       `json:"quantity"`
	Unit          string    `json:"unit"`
	UnitPrice     int64     `json:"unit_price"`
	TierName      string    `json:"tier_name"`
	Subtotal      int64     `json:"subtotal"`
	IsAvailable   bool      `json:"is_available"`
	AvailableQty  *int      `json:"available_qty,omitempty"`
}

// TransactionFilter is the filter options for listing transactions
type TransactionFilter struct {
	Search        *string            `json:"search,omitempty"`
	CustomerID    *uuid.UUID         `json:"customer_id,omitempty"`
	Status        *TransactionStatus `json:"status,omitempty"`
	PaymentMethod *PaymentMethod     `json:"payment_method,omitempty"`
	DateFrom      *time.Time         `json:"date_from,omitempty"`
	DateTo        *time.Time         `json:"date_to,omitempty"`
	Page          int                `json:"page,omitempty"`
	PerPage       int                `json:"per_page,omitempty"`
	SortBy        string             `json:"sort_by,omitempty"`
	SortOrder     string             `json:"sort_order,omitempty"`
}

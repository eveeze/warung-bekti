package domain

import (
	"time"

	"github.com/google/uuid"
)

// OpnameStatus represents the status of a stock opname session
type OpnameStatus string

const (
	OpnameStatusDraft      OpnameStatus = "draft"
	OpnameStatusInProgress OpnameStatus = "in_progress"
	OpnameStatusCompleted  OpnameStatus = "completed"
	OpnameStatusCancelled  OpnameStatus = "cancelled"
)

// StockOpnameSession represents a stock taking session
type StockOpnameSession struct {
	ID             uuid.UUID    `json:"id"`
	SessionCode    string       `json:"session_code"`
	Status         OpnameStatus `json:"status"`
	Notes          *string      `json:"notes,omitempty"`
	TotalProducts  int          `json:"total_products"`
	TotalVariance  int          `json:"total_variance"`
	TotalLossValue int64        `json:"total_loss_value"`
	TotalGainValue int64        `json:"total_gain_value"`
	StartedAt      *time.Time   `json:"started_at,omitempty"`
	CompletedAt    *time.Time   `json:"completed_at,omitempty"`
	CreatedBy      *string      `json:"created_by,omitempty"`
	CompletedBy    *string      `json:"completed_by,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`

	// Relations
	Items []StockOpnameItem `json:"items,omitempty"`
}

// StockOpnameItem represents a single product count in an opname session
type StockOpnameItem struct {
	ID            uuid.UUID  `json:"id"`
	SessionID     uuid.UUID  `json:"session_id"`
	ProductID     uuid.UUID  `json:"product_id"`
	SystemStock   int        `json:"system_stock"`
	PhysicalStock int        `json:"physical_stock"`
	Variance      int        `json:"variance"`
	CostPerUnit   int64      `json:"cost_per_unit"`
	VarianceValue int64      `json:"variance_value"`
	Notes         *string    `json:"notes,omitempty"`
	CountedBy     *string    `json:"counted_by,omitempty"`
	CountedAt     time.Time  `json:"counted_at"`

	// Relations
	Product *Product `json:"product,omitempty"`
}

// VarianceReport represents a summary of stock variances
type VarianceReport struct {
	SessionID      uuid.UUID           `json:"session_id"`
	SessionCode    string              `json:"session_code"`
	TotalProducts  int                 `json:"total_products"`
	TotalVariance  int                 `json:"total_variance"`
	TotalLossValue int64               `json:"total_loss_value"`
	TotalGainValue int64               `json:"total_gain_value"`
	NetValue       int64               `json:"net_value"` // gain - loss
	Items          []VarianceReportItem `json:"items"`
}

// VarianceReportItem represents a single item in the variance report
type VarianceReportItem struct {
	ProductID     uuid.UUID `json:"product_id"`
	ProductName   string    `json:"product_name"`
	Barcode       *string   `json:"barcode,omitempty"`
	SystemStock   int       `json:"system_stock"`
	PhysicalStock int       `json:"physical_stock"`
	Variance      int       `json:"variance"`
	CostPerUnit   int64     `json:"cost_per_unit"`
	VarianceValue int64     `json:"variance_value"`
	Notes         *string   `json:"notes,omitempty"`
}

// StartOpnameInput is the input for starting a new opname session
type StartOpnameInput struct {
	Notes     *string `json:"notes,omitempty"`
	CreatedBy string  `json:"created_by"`
}

// RecordCountInput is the input for recording a physical count
type RecordCountInput struct {
	SessionID     uuid.UUID `json:"session_id"`
	ProductID     uuid.UUID `json:"product_id"`
	PhysicalStock int       `json:"physical_stock"`
	Notes         *string   `json:"notes,omitempty"`
	CountedBy     string    `json:"counted_by"`
}

// FinalizeOpnameInput is the input for finalizing an opname session
type FinalizeOpnameInput struct {
	SessionID   uuid.UUID `json:"session_id"`
	CompletedBy string    `json:"completed_by"`
	ApplyAdjustments bool `json:"apply_adjustments"` // Whether to auto-adjust stock
}

// ShoppingListItem represents an item in the auto-generated shopping list
type ShoppingListItem struct {
	ID            uuid.UUID `json:"id"`
	ProductID     uuid.UUID `json:"product_id"`
	CurrentStock  int       `json:"current_stock"`
	MinStock      int       `json:"min_stock"`
	SuggestedQty  int       `json:"suggested_qty"`
	EstimatedCost *int64    `json:"estimated_cost,omitempty"`
	IsPurchased   bool      `json:"is_purchased"`
	Notes         *string   `json:"notes,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relations
	Product *Product `json:"product,omitempty"`
}

// ShoppingList represents the auto-generated restock plan
type ShoppingList struct {
	GeneratedAt   time.Time          `json:"generated_at"`
	TotalItems    int                `json:"total_items"`
	TotalCost     int64              `json:"total_cost"`
	Items         []ShoppingListItem `json:"items"`
}

// NearExpiryItem represents an item nearing expiry
type NearExpiryItem struct {
	ProductID    uuid.UUID  `json:"product_id"`
	ProductName  string     `json:"product_name"`
	Barcode      *string    `json:"barcode,omitempty"`
	BatchNumber  *string    `json:"batch_number,omitempty"`
	ExpiryDate   time.Time  `json:"expiry_date"`
	DaysUntilExpiry int     `json:"days_until_expiry"`
	Quantity     int        `json:"quantity"`
	CostPrice    int64      `json:"cost_price"`
}

// NearExpiryReport represents items nearing expiry date
type NearExpiryReport struct {
	GeneratedAt time.Time        `json:"generated_at"`
	DaysAhead   int              `json:"days_ahead"` // How many days ahead we're looking
	TotalItems  int              `json:"total_items"`
	TotalValue  int64            `json:"total_value"`
	Items       []NearExpiryItem `json:"items"`
}

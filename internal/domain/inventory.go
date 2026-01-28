package domain

import (
	"time"

	"github.com/google/uuid"
)

// StockMovementType represents the type of stock movement
type StockMovementType string

const (
	StockMovementTypeInitial     StockMovementType = "initial"
	StockMovementTypePurchase    StockMovementType = "purchase"
	StockMovementTypeSale        StockMovementType = "sale"
	StockMovementTypeAdjustment  StockMovementType = "adjustment"
	StockMovementTypeReturn      StockMovementType = "return"
	StockMovementTypeDamage      StockMovementType = "damage"
	StockMovementTypeTransferIn  StockMovementType = "transfer_in"
	StockMovementTypeTransferOut StockMovementType = "transfer_out"
)

// StockMovement represents a stock movement record
type StockMovement struct {
	ID            uuid.UUID          `json:"id"`
	ProductID     uuid.UUID          `json:"product_id"`
	Type          StockMovementType  `json:"type"`
	Quantity      int                `json:"quantity"` // positive for in, negative for out
	StockBefore   int                `json:"stock_before"`
	StockAfter    int                `json:"stock_after"`
	ReferenceType *string            `json:"reference_type,omitempty"` // 'transaction', 'purchase', 'adjustment'
	ReferenceID   *uuid.UUID         `json:"reference_id,omitempty"`
	CostPerUnit   *int64             `json:"cost_per_unit,omitempty"`
	Notes         *string            `json:"notes,omitempty"`
	CreatedBy     *string            `json:"created_by,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`

	// Relations
	Product *Product `json:"product,omitempty"`
}

// StockAdjustmentInput is the input for manual stock adjustment
type StockAdjustmentInput struct {
	ProductID   uuid.UUID `json:"product_id"`
	Quantity    int       `json:"quantity"` // can be positive or negative
	Reason      string    `json:"reason"`   // will be used as notes
	CreatedBy   *string   `json:"created_by,omitempty"`
}

// RestockInput is the input for restocking products
type RestockInput struct {
	ProductID   uuid.UUID `json:"product_id"`
	Quantity    int       `json:"quantity"`
	CostPerUnit int64     `json:"cost_per_unit"`
	Notes       *string   `json:"notes,omitempty"`
	CreatedBy   *string   `json:"created_by,omitempty"`
}

// BulkRestockInput is the input for restocking multiple products
type BulkRestockInput struct {
	SupplierID *uuid.UUID    `json:"supplier_id,omitempty"`
	Items      []RestockItem `json:"items"`
	Notes      *string       `json:"notes,omitempty"`
	CreatedBy  *string       `json:"created_by,omitempty"`
}

// RestockItem is a single item in a bulk restock
type RestockItem struct {
	ProductID   uuid.UUID `json:"product_id"`
	Quantity    int       `json:"quantity"`
	CostPerUnit int64     `json:"cost_per_unit"`
}

// StockMovementFilter is the filter for listing stock movements
type StockMovementFilter struct {
	ProductID     *uuid.UUID         `json:"product_id,omitempty"`
	Type          *StockMovementType `json:"type,omitempty"`
	ReferenceType *string            `json:"reference_type,omitempty"`
	DateFrom      *time.Time         `json:"date_from,omitempty"`
	DateTo        *time.Time         `json:"date_to,omitempty"`
	Page          int                `json:"page,omitempty"`
	PerPage       int                `json:"per_page,omitempty"`
}

// LowStockProduct represents a product with low stock
type LowStockProduct struct {
	Product       Product `json:"product"`
	DeficitAmount int     `json:"deficit_amount"` // how much below min_stock
}

// StockReport is a stock inventory report
type StockReport struct {
	TotalProducts     int               `json:"total_products"`
	TotalStockValue   int64             `json:"total_stock_value"` // current_stock * cost_price
	LowStockCount     int               `json:"low_stock_count"`
	OutOfStockCount   int               `json:"out_of_stock_count"`
	LowStockProducts  []LowStockProduct `json:"low_stock_products,omitempty"`
}

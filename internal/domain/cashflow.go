package domain

import (
	"time"

	"github.com/google/uuid"
)

// CashFlowType represents the type of cash flow
type CashFlowType string

const (
	CashFlowTypeIncome  CashFlowType = "income"
	CashFlowTypeExpense CashFlowType = "expense"
)

// DrawerSessionStatus represents the status of a drawer session
type DrawerSessionStatus string

const (
	DrawerSessionStatusOpen   DrawerSessionStatus = "open"
	DrawerSessionStatusClosed DrawerSessionStatus = "closed"
)

// CashFlowCategory represents a category for cash flow records
type CashFlowCategory struct {
	ID          uuid.UUID    `json:"id"`
	Name        string       `json:"name"`
	Type        CashFlowType `json:"type"`
	Description *string      `json:"description,omitempty"`
	IsActive    bool         `json:"is_active"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// CashDrawerSession represents a cash drawer session (shift)
type CashDrawerSession struct {
	ID              uuid.UUID           `json:"id"`
	SessionDate     time.Time           `json:"session_date"`
	OpeningBalance  int64               `json:"opening_balance"`
	ClosingBalance  *int64              `json:"closing_balance,omitempty"`
	ExpectedClosing *int64              `json:"expected_closing,omitempty"`
	Difference      *int64              `json:"difference,omitempty"`
	Status          DrawerSessionStatus `json:"status"`
	OpenedBy        *string             `json:"opened_by,omitempty"`
	ClosedBy        *string             `json:"closed_by,omitempty"`
	Notes           *string             `json:"notes,omitempty"`
	OpenedAt        time.Time           `json:"opened_at"`
	ClosedAt        *time.Time          `json:"closed_at,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`

	// Computed fields
	TotalIncome  int64 `json:"total_income,omitempty"`
	TotalExpense int64 `json:"total_expense,omitempty"`
}

// CashFlowRecord represents a single cash flow entry
type CashFlowRecord struct {
	ID              uuid.UUID    `json:"id"`
	DrawerSessionID *uuid.UUID   `json:"drawer_session_id,omitempty"`
	CategoryID      *uuid.UUID   `json:"category_id,omitempty"`
	Type            CashFlowType `json:"type"`
	Amount          int64        `json:"amount"`
	Description     *string      `json:"description,omitempty"`
	ReferenceType   *string      `json:"reference_type,omitempty"`
	ReferenceID     *uuid.UUID   `json:"reference_id,omitempty"`
	CreatedBy       *string      `json:"created_by,omitempty"`
	CreatedAt       time.Time    `json:"created_at"`

	// Relations
	Category *CashFlowCategory `json:"category,omitempty"`
}

// OpenDrawerInput is the input for opening a drawer session
type OpenDrawerInput struct {
	OpeningBalance int64   `json:"opening_balance"`
	OpenedBy       string  `json:"opened_by"`
	Notes          *string `json:"notes,omitempty"`
}

// CloseDrawerInput is the input for closing a drawer session
type CloseDrawerInput struct {
	SessionID      uuid.UUID `json:"session_id"`
	ClosingBalance int64     `json:"closing_balance"`
	ClosedBy       string    `json:"closed_by"`
	Notes          *string   `json:"notes,omitempty"`
}

// CashFlowInput is the input for recording a cash flow
type CashFlowInput struct {
	CategoryID  *uuid.UUID   `json:"category_id,omitempty"`
	Type        CashFlowType `json:"type"`
	Amount      int64        `json:"amount"`
	Description *string      `json:"description,omitempty"`
	CreatedBy   string       `json:"created_by"`
}

// CashFlowFilter is the filter for listing cash flows
type CashFlowFilter struct {
	SessionID  *uuid.UUID    `json:"session_id,omitempty"`
	CategoryID *uuid.UUID    `json:"category_id,omitempty"`
	Type       *CashFlowType `json:"type,omitempty"`
	DateFrom   *time.Time    `json:"date_from,omitempty"`
	DateTo     *time.Time    `json:"date_to,omitempty"`
	Page       int           `json:"page,omitempty"`
	PerPage    int           `json:"per_page,omitempty"`
}

// ProfitLossReport represents the profit/loss summary
type ProfitLossReport struct {
	DateFrom       time.Time           `json:"date_from"`
	DateTo         time.Time           `json:"date_to"`
	TotalRevenue   int64               `json:"total_revenue"`      // Total sales
	TotalCOGS      int64               `json:"total_cogs"`         // HPP (Harga Pokok Penjualan)
	GrossProfit    int64               `json:"gross_profit"`       // Revenue - COGS
	TotalExpenses  int64               `json:"total_expenses"`     // Operational expenses
	NetProfit      int64               `json:"net_profit"`         // Gross - Expenses
	ProfitMargin   float64             `json:"profit_margin"`      // Net Profit / Revenue * 100
	ByPaymentMethod map[string]int64   `json:"by_payment_method"`  // Sales breakdown
	TopExpenses    []ExpenseBreakdown  `json:"top_expenses,omitempty"`
}

// ExpenseBreakdown represents expense by category
type ExpenseBreakdown struct {
	CategoryName string `json:"category_name"`
	Amount       int64  `json:"amount"`
	Percentage   float64 `json:"percentage"`
}

// SalesByMethodReport represents sales breakdown by payment method
type SalesByMethodReport struct {
	DateFrom   time.Time        `json:"date_from"`
	DateTo     time.Time        `json:"date_to"`
	TotalSales int64            `json:"total_sales"`
	Methods    []MethodBreakdown `json:"methods"`
}

// MethodBreakdown represents sales for a single payment method
type MethodBreakdown struct {
	Method     PaymentMethod `json:"method"`
	Amount     int64         `json:"amount"`
	Count      int           `json:"count"`
	Percentage float64       `json:"percentage"`
}

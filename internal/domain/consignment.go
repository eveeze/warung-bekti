package domain

import (
	"time"

	"github.com/google/uuid"
)

type Consignor struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Phone       *string   `json:"phone,omitempty"`
	Address     *string   `json:"address,omitempty"`
	BankAccount *string   `json:"bank_account,omitempty"`
	BankName    *string   `json:"bank_name,omitempty"`
	Notes       *string   `json:"notes,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Products []Product `json:"products,omitempty"`
}

type SettlementStatus string

const (
	SettlementStatusDraft     SettlementStatus = "draft"
	SettlementStatusConfirmed SettlementStatus = "confirmed"
	SettlementStatusPaid      SettlementStatus = "paid"
)

type ConsignmentSettlement struct {
	ID               uuid.UUID        `json:"id"`
	SettlementNumber string           `json:"settlement_number"`
	ConsignorID      uuid.UUID        `json:"consignor_id"`
	PeriodStart      time.Time        `json:"period_start"`
	PeriodEnd        time.Time        `json:"period_end"`
	TotalSales       int64            `json:"total_sales"`
	CommissionAmount int64            `json:"commission_amount"`
	ConsignorAmount  int64            `json:"consignor_amount"`
	Status           SettlementStatus `json:"status"`
	Notes            *string          `json:"notes,omitempty"`
	CreatedBy        *string          `json:"created_by,omitempty"`
	PaidBy           *string          `json:"paid_by,omitempty"`
	PaidAt           *time.Time       `json:"paid_at,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`

	Consignor *Consignor                 `json:"consignor,omitempty"`
	Items     []ConsignmentSettlementItem `json:"items,omitempty"`
}

type ConsignmentSettlementItem struct {
	ID              uuid.UUID `json:"id"`
	SettlementID    uuid.UUID `json:"settlement_id"`
	ProductID       uuid.UUID `json:"product_id"`
	ProductName     string    `json:"product_name"`
	QuantitySold    int       `json:"quantity_sold"`
	UnitPrice       int64     `json:"unit_price"`
	TotalSales      int64     `json:"total_sales"`
	CommissionRate  float64   `json:"commission_rate"`
	CommissionAmount int64    `json:"commission_amount"`
	ConsignorAmount int64     `json:"consignor_amount"`
	CreatedAt       time.Time `json:"created_at"`
}

type CreateConsignorInput struct {
	Name        string  `json:"name"`
	Phone       *string `json:"phone,omitempty"`
	Address     *string `json:"address,omitempty"`
	BankAccount *string `json:"bank_account,omitempty"`
	BankName    *string `json:"bank_name,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

type UpdateConsignorInput struct {
	Name        *string `json:"name,omitempty"`
	Phone       *string `json:"phone,omitempty"`
	Address     *string `json:"address,omitempty"`
	BankAccount *string `json:"bank_account,omitempty"`
	BankName    *string `json:"bank_name,omitempty"`
	Notes       *string `json:"notes,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

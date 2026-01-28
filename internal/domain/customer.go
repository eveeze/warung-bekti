package domain

import (
	"time"

	"github.com/google/uuid"
)

// Customer represents a customer for kasbon tracking
type Customer struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Phone       *string    `json:"phone,omitempty"`
	Address     *string    `json:"address,omitempty"`
	Notes       *string    `json:"notes,omitempty"`
	CreditLimit int64      `json:"credit_limit"` // batas maksimal kasbon
	CurrentDebt int64      `json:"current_debt"` // total hutang saat ini
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relations (populated when needed)
	KasbonRecords []KasbonRecord `json:"kasbon_records,omitempty"`
}

// CanAddDebt checks if customer can add more debt
func (c *Customer) CanAddDebt(amount int64) bool {
	if c.CreditLimit == 0 {
		return true // unlimited credit
	}
	return (c.CurrentDebt + amount) <= c.CreditLimit
}

// RemainingCredit returns how much credit is still available
func (c *Customer) RemainingCredit() int64 {
	if c.CreditLimit == 0 {
		return -1 // unlimited
	}
	remaining := c.CreditLimit - c.CurrentDebt
	if remaining < 0 {
		return 0
	}
	return remaining
}

// HasDebt checks if customer has any debt
func (c *Customer) HasDebt() bool {
	return c.CurrentDebt > 0
}

// CustomerCreateInput is the input for creating a customer
type CustomerCreateInput struct {
	Name        string  `json:"name"`
	Phone       *string `json:"phone,omitempty"`
	Address     *string `json:"address,omitempty"`
	Notes       *string `json:"notes,omitempty"`
	CreditLimit *int64  `json:"credit_limit,omitempty"`
}

// CustomerUpdateInput is the input for updating a customer
type CustomerUpdateInput struct {
	Name        *string `json:"name,omitempty"`
	Phone       *string `json:"phone,omitempty"`
	Address     *string `json:"address,omitempty"`
	Notes       *string `json:"notes,omitempty"`
	CreditLimit *int64  `json:"credit_limit,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// CustomerFilter is the filter options for listing customers
type CustomerFilter struct {
	Search       *string `json:"search,omitempty"`
	HasDebt      *bool   `json:"has_debt,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
	Page         int     `json:"page,omitempty"`
	PerPage      int     `json:"per_page,omitempty"`
	SortBy       string  `json:"sort_by,omitempty"`
	SortOrder    string  `json:"sort_order,omitempty"`
}

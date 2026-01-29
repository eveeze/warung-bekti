package domain

import (
	"sort"
	"time"

	"github.com/google/uuid"
)

// Product represents a product in the inventory
type Product struct {
	ID            uuid.UUID  `json:"id"`
	Barcode       *string    `json:"barcode,omitempty"`
	SKU           *string    `json:"sku,omitempty"`
	Name          string     `json:"name"`
	Description   *string    `json:"description,omitempty"`
	CategoryID    *uuid.UUID `json:"category_id,omitempty"`
	Unit          string     `json:"unit"`
	BasePrice     int64      `json:"base_price"`      // harga jual dasar (rupiah)
	CostPrice     int64      `json:"cost_price"`      // harga beli/HPP
	IsStockActive bool       `json:"is_stock_active"` // hybrid stock toggle
	CurrentStock  int        `json:"current_stock"`
	MinStockAlert int        `json:"min_stock_alert"`
	MaxStock      *int       `json:"max_stock,omitempty"`
	ImageURL       *string    `json:"image_url,omitempty"`
	IsRefillable   bool       `json:"is_refillable"`
	EmptyProductID *uuid.UUID `json:"empty_product_id,omitempty"`
	FullProductID  *uuid.UUID `json:"full_product_id,omitempty"`
	IsActive       bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	// Relations (populated when needed)
	Category     *Category      `json:"category,omitempty"`
	PricingTiers []PricingTier  `json:"pricing_tiers,omitempty"`
}

// PricingTier represents a pricing tier for variable pricing
type PricingTier struct {
	ID          uuid.UUID  `json:"id"`
	ProductID   uuid.UUID  `json:"product_id"`
	Name        *string    `json:"name,omitempty"`       // "Eceran", "Grosir", "Promo 3+"
	MinQuantity int        `json:"min_quantity"`
	MaxQuantity *int       `json:"max_quantity,omitempty"` // NULL = unlimited
	Price       int64      `json:"price"`                  // harga per unit di tier ini
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CalculatePrice calculates the best price for a given quantity
// It returns the price per unit and the tier name that was applied
func (p *Product) CalculatePrice(quantity int) (pricePerUnit int64, tierName string, tierID *uuid.UUID) {
	if len(p.PricingTiers) == 0 {
		return p.BasePrice, "Harga Dasar", nil
	}

	// Sort tiers by min_quantity descending to find the best matching tier
	sortedTiers := make([]PricingTier, len(p.PricingTiers))
	copy(sortedTiers, p.PricingTiers)
	sort.Slice(sortedTiers, func(i, j int) bool {
		return sortedTiers[i].MinQuantity > sortedTiers[j].MinQuantity
	})

	for _, tier := range sortedTiers {
		if !tier.IsActive {
			continue
		}
		
		// Check if quantity falls within this tier's range
		if quantity >= tier.MinQuantity {
			// If max_quantity is set, check upper bound
			if tier.MaxQuantity != nil && quantity > *tier.MaxQuantity {
				continue
			}
			
			name := "Tier"
			if tier.Name != nil {
				name = *tier.Name
			}
			tierIDCopy := tier.ID
			return tier.Price, name, &tierIDCopy
		}
	}

	// If no tier matches, use base price
	return p.BasePrice, "Harga Dasar", nil
}

// CalculateTotal calculates the total price for a given quantity
func (p *Product) CalculateTotal(quantity int) (total int64, pricePerUnit int64, tierName string, tierID *uuid.UUID) {
	pricePerUnit, tierName, tierID = p.CalculatePrice(quantity)
	total = pricePerUnit * int64(quantity)
	return
}

// IsLowStock checks if the product is below minimum stock level
func (p *Product) IsLowStock() bool {
	return p.IsStockActive && p.CurrentStock <= p.MinStockAlert
}

// IsOutOfStock checks if the product is out of stock
func (p *Product) IsOutOfStock() bool {
	return p.IsStockActive && p.CurrentStock <= 0
}

// CanSell checks if the product can be sold with the given quantity
func (p *Product) CanSell(quantity int) bool {
	if !p.IsStockActive {
		return true // tidak pakai stock tracking
	}
	return p.CurrentStock >= quantity
}

// EstimatedProfit calculates the estimated profit per unit
func (p *Product) EstimatedProfit() int64 {
	return p.BasePrice - p.CostPrice
}

// ProductCreateInput is the input for creating a product
type ProductCreateInput struct {
	Barcode       *string            `json:"barcode,omitempty"`
	SKU           *string            `json:"sku,omitempty"`
	Name          string             `json:"name"`
	Description   *string            `json:"description,omitempty"`
	CategoryID    *uuid.UUID         `json:"category_id,omitempty"`
	Unit          string             `json:"unit"`
	BasePrice     int64              `json:"base_price"`
	CostPrice     int64              `json:"cost_price"`
	IsStockActive *bool              `json:"is_stock_active,omitempty"`
	CurrentStock  *int               `json:"current_stock,omitempty"`
	MinStockAlert *int               `json:"min_stock_alert,omitempty"`
	MaxStock      *int               `json:"max_stock,omitempty"`
	IsRefillable  *bool              `json:"is_refillable,omitempty"`
	EmptyProductID *uuid.UUID        `json:"empty_product_id,omitempty"`
	FullProductID  *uuid.UUID        `json:"full_product_id,omitempty"`
	PricingTiers   []PricingTierInput `json:"pricing_tiers,omitempty"`
	ImageURL       *string            `json:"image_url,omitempty"`
}

// ProductUpdateInput is the input for updating a product
type ProductUpdateInput struct {
	Barcode       *string    `json:"barcode,omitempty"`
	SKU           *string    `json:"sku,omitempty"`
	Name          *string    `json:"name,omitempty"`
	Description   *string    `json:"description,omitempty"`
	CategoryID    *uuid.UUID `json:"category_id,omitempty"`
	Unit          *string    `json:"unit,omitempty"`
	BasePrice     *int64     `json:"base_price,omitempty"`
	CostPrice     *int64     `json:"cost_price,omitempty"`
	IsStockActive *bool      `json:"is_stock_active,omitempty"`
	MinStockAlert *int       `json:"min_stock_alert,omitempty"`
	MaxStock      *int       `json:"max_stock,omitempty"`
	IsRefillable  *bool      `json:"is_refillable,omitempty"`
	EmptyProductID *uuid.UUID `json:"empty_product_id,omitempty"`
	FullProductID  *uuid.UUID `json:"full_product_id,omitempty"`
	IsActive      *bool      `json:"is_active,omitempty"`
	ImageURL      *string    `json:"image_url,omitempty"`
}

// PricingTierInput is the input for creating/updating a pricing tier
type PricingTierInput struct {
	Name        *string `json:"name,omitempty"`
	MinQuantity int     `json:"min_quantity"`
	MaxQuantity *int    `json:"max_quantity,omitempty"`
	Price       int64   `json:"price"`
}

// ProductFilter is the filter options for listing products
type ProductFilter struct {
	Search        *string    `json:"search,omitempty"`
	CategoryID    *uuid.UUID `json:"category_id,omitempty"`
	IsActive      *bool      `json:"is_active,omitempty"`
	IsStockActive *bool      `json:"is_stock_active,omitempty"`
	LowStockOnly  bool       `json:"low_stock_only,omitempty"`
	Page          int        `json:"page,omitempty"`
	PerPage       int        `json:"per_page,omitempty"`
	SortBy        string     `json:"sort_by,omitempty"`
	SortOrder     string     `json:"sort_order,omitempty"` // "asc" or "desc"
}

// DefaultFilter returns a filter with default values
func DefaultFilter() ProductFilter {
	isActive := true
	return ProductFilter{
		IsActive:  &isActive,
		Page:      1,
		PerPage:   20,
		SortBy:    "name",
		SortOrder: "asc",
	}
}

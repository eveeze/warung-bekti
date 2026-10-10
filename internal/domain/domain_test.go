package domain_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/domain"
)

func TestCustomerDomain(t *testing.T) {
	c := &domain.Customer{
		CreditLimit: 100000,
		CurrentDebt: 40000,
	}

	if !c.HasDebt() {
		t.Errorf("expected customer to have debt")
	}
	if rem := c.RemainingCredit(); rem != 60000 {
		t.Errorf("expected remaining credit 60000, got %d", rem)
	}
	if !c.CanAddDebt(50000) {
		t.Errorf("expected to be able to add 50000 debt")
	}
	if c.CanAddDebt(70000) {
		t.Errorf("expected to NOT be able to add 70000 debt")
	}

	noDebt := &domain.Customer{CurrentDebt: 0}
	if noDebt.HasDebt() {
		t.Errorf("expected customer to not have debt")
	}

	// Unlimited credit limit (CreditLimit = 0)
	unlimited := &domain.Customer{CreditLimit: 0, CurrentDebt: 500000}
	if !unlimited.CanAddDebt(1000000) {
		t.Errorf("expected unlimited customer to add debt")
	}
	if rem := unlimited.RemainingCredit(); rem != -1 {
		t.Errorf("expected -1 remaining credit for unlimited, got %d", rem)
	}

	// Over limit credit (CurrentDebt > CreditLimit)
	overLimit := &domain.Customer{CreditLimit: 10000, CurrentDebt: 15000}
	if rem := overLimit.RemainingCredit(); rem != 0 {
		t.Errorf("expected 0 remaining credit for overLimit, got %d", rem)
	}
}

func TestProductDomain(t *testing.T) {
	tier1ID := uuid.New()
	maxTier := 20
	tName := "Grosir 1"
	p := &domain.Product{
		BasePrice:     10000,
		CostPrice:     7000,
		CurrentStock:  15,
		MinStockAlert: 5,
		IsStockActive: true,
		PricingTiers: []domain.PricingTier{
			{
				ID:          tier1ID,
				MinQuantity: 10,
				MaxQuantity: &maxTier,
				Price:       9000,
				Name:        &tName,
				IsActive:    true,
			},
		},
	}

	// Product with no tiers
	pNoTiers := &domain.Product{BasePrice: 5000}
	pPrice, pName, pTierID := pNoTiers.CalculatePrice(10)
	if pPrice != 5000 || pName != "Harga Dasar" || pTierID != nil {
		t.Errorf("expected base price for product without tiers")
	}

	// Product with tier having nil name and quantity exceeding max tier
	pNilTier := &domain.Product{
		BasePrice: 10000,
		PricingTiers: []domain.PricingTier{
			{
				MinQuantity: 5,
				MaxQuantity: &maxTier, // 20
				Price:       8500,
				Name:        nil,
				IsActive:    true,
			},
		},
	}
	pPriceNil, pNameNil, _ := pNilTier.CalculatePrice(10)
	if pPriceNil != 8500 || pNameNil != "Tier" {
		t.Errorf("expected default tier name 'Tier', got %s", pNameNil)
	}

	// Quantity exceeds max quantity (25 > 20) -> should fallback to base price
	pPriceExceed, _, _ := pNilTier.CalculatePrice(25)
	if pPriceExceed != 10000 {
		t.Errorf("expected base price when quantity exceeds max tier, got %d", pPriceExceed)
	}

	priceTier, nameTier, idTier := p.CalculatePrice(12)
	if priceTier != 9000 || nameTier != "Grosir 1" || idTier == nil || *idTier != tier1ID {
		t.Errorf("expected tier price 9000, got %d, %s", priceTier, nameTier)
	}

	// CalculateTotal
	total, unitPrice, _, _ := p.CalculateTotal(12)
	if total != 108000 || unitPrice != 9000 {
		t.Errorf("expected total 108000, got %d", total)
	}

	// Stock checks
	if p.IsLowStock() {
		t.Errorf("expected stock 15 not low for alert 5")
	}
	if p.IsOutOfStock() {
		t.Errorf("expected not out of stock")
	}
	if !p.CanSell(10) {
		t.Errorf("expected can sell 10")
	}
	if p.CanSell(20) {
		t.Errorf("expected cannot sell 20")
	}

	p.CurrentStock = 3
	if !p.IsLowStock() {
		t.Errorf("expected low stock for 3 <= 5")
	}

	p.CurrentStock = 0
	if !p.IsOutOfStock() {
		t.Errorf("expected out of stock for 0")
	}

	p.IsStockActive = false
	if !p.CanSell(999) {
		t.Errorf("expected unmonitored stock to always allow selling")
	}

	if profit := p.EstimatedProfit(); profit != 3000 {
		t.Errorf("expected estimated profit 3000, got %d", profit)
	}

	filter := domain.DefaultFilter()
	if filter.Page != 1 || filter.PerPage != 20 || filter.SortBy != "name" || filter.SortOrder != "asc" {
		t.Errorf("unexpected default filter: %+v", filter)
	}
}

func TestTransactionDomain(t *testing.T) {
	item := &domain.TransactionItem{
		Quantity:    3,
		UnitPrice:   10000,
		CostPrice:   7000,
		TotalAmount: 30000,
		Subtotal:    30000,
	}
	if prof := item.Profit(); prof != 9000 {
		t.Errorf("expected profit 9000, got %d", prof)
	}
}

func TestPaymentDomain(t *testing.T) {
	success := domain.PaymentStatusSettlement
	if !success.IsSuccess() || !success.IsFinal() {
		t.Errorf("expected settlement to be success and final")
	}

	capture := domain.PaymentStatusCapture
	if !capture.IsSuccess() || !capture.IsFinal() {
		t.Errorf("expected capture to be success and final")
	}

	pending := domain.PaymentStatusPending
	if pending.IsSuccess() || pending.IsFinal() {
		t.Errorf("expected pending not success and not final")
	}

	cancel := domain.PaymentStatusCancel
	if cancel.IsSuccess() || !cancel.IsFinal() {
		t.Errorf("expected cancel not success but final")
	}
}

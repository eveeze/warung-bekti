package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/domain"
)

// Mock repositories for unit testing
// These allow testing service logic without database

// MockProductRepository implements product repository for testing
type MockProductRepository struct {
	products map[uuid.UUID]*domain.Product
}

func NewMockProductRepository() *MockProductRepository {
	return &MockProductRepository{
		products: make(map[uuid.UUID]*domain.Product),
	}
}

func (m *MockProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	product, ok := m.products[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return product, nil
}

func (m *MockProductRepository) AddProduct(product *domain.Product) {
	m.products[product.ID] = product
}

// MockCustomerRepository implements customer repository for testing
type MockCustomerRepository struct {
	customers map[uuid.UUID]*domain.Customer
}

func NewMockCustomerRepository() *MockCustomerRepository {
	return &MockCustomerRepository{
		customers: make(map[uuid.UUID]*domain.Customer),
	}
}

func (m *MockCustomerRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	customer, ok := m.customers[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return customer, nil
}

func (m *MockCustomerRepository) AddCustomer(customer *domain.Customer) {
	m.customers[customer.ID] = customer
}

// Test Cases

// TestCartCalculation tests cart price calculation logic
func TestCartCalculation(t *testing.T) {
	// Setup mock products
	productRepo := NewMockProductRepository()
	
	// Product with tiered pricing
	product1 := &domain.Product{
		ID:            uuid.New(),
		Name:          "Test Product 1",
		BasePrice:     10000,
		CostPrice:     7500,
		IsActive:      true,
		IsStockActive: true,
		CurrentStock:  100,
		PricingTiers: []domain.PricingTier{
			{MinQuantity: 10, MaxQuantity: intPtr(49), Price: 9000},
			{MinQuantity: 50, MaxQuantity: nil, Price: 8000},
		},
	}
	productRepo.AddProduct(product1)
	
	// Product without tiers
	product2 := &domain.Product{
		ID:            uuid.New(),
		Name:          "Test Product 2",
		BasePrice:     5000,
		CostPrice:     3500,
		IsActive:      true,
		IsStockActive: true,
		CurrentStock:  50,
	}
	productRepo.AddProduct(product2)
	
	tests := []struct {
		name           string
		items          []domain.CartItem
		expectedTotal  int64
		expectedMargin int64
	}{
		{
			name: "Single item - base price",
			items: []domain.CartItem{
				{ProductID: product1.ID, Quantity: 5},
			},
			expectedTotal: 50000, // 5 * 10000
		},
		{
			name: "Single item - tier 1 price",
			items: []domain.CartItem{
				{ProductID: product1.ID, Quantity: 15},
			},
			expectedTotal: 135000, // 15 * 9000
		},
		{
			name: "Single item - tier 2 price",
			items: []domain.CartItem{
				{ProductID: product1.ID, Quantity: 100},
			},
			expectedTotal: 800000, // 100 * 8000
		},
		{
			name: "Multiple items",
			items: []domain.CartItem{
				{ProductID: product1.ID, Quantity: 10}, // 90000 (tier 1)
				{ProductID: product2.ID, Quantity: 3},  // 15000
			},
			expectedTotal: 105000,
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			total := calculateCartTotal(productRepo, tc.items)
			
			if total != tc.expectedTotal {
				t.Errorf("Expected total %d, got %d", tc.expectedTotal, total)
			}
		})
	}
}

// TestStockValidation tests stock availability logic
func TestStockValidation(t *testing.T) {
	productRepo := NewMockProductRepository()
	
	product := &domain.Product{
		ID:            uuid.New(),
		Name:          "Limited Stock Product",
		BasePrice:     10000,
		IsActive:      true,
		IsStockActive: true,
		CurrentStock:  10,
	}
	productRepo.AddProduct(product)
	
	tests := []struct {
		name        string
		quantity    int
		expectError bool
	}{
		{"Within stock", 5, false},
		{"Exact stock", 10, false},
		{"Over stock", 15, true},
		{"Zero quantity", 0, true},
		{"Negative quantity", -1, true},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateStock(product, tc.quantity)
			
			hasError := err != nil
			if hasError != tc.expectError {
				t.Errorf("Expected error=%v, got error=%v (%v)", tc.expectError, hasError, err)
			}
		})
	}
}

// TestKasbonCreditLimit tests credit limit validation
func TestKasbonCreditLimit(t *testing.T) {
	customerRepo := NewMockCustomerRepository()
	
	customer := &domain.Customer{
		ID:          uuid.New(),
		Name:        "Test Customer",
		CreditLimit: 500000,
		CurrentDebt: 300000,
		IsActive:    true,
	}
	customerRepo.AddCustomer(customer)
	
	tests := []struct {
		name        string
		amount      int64
		expectError bool
	}{
		{"Within limit", 100000, false},
		{"Exact remaining", 200000, false},
		{"Over limit", 300000, true},
		{"Way over limit", 1000000, true},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCreditLimit(customer, tc.amount)
			
			hasError := err != nil
			if hasError != tc.expectError {
				t.Errorf("Expected error=%v, got error=%v", tc.expectError, hasError)
			}
		})
	}
}

// TestPaymentValidation tests payment amount validation
func TestPaymentValidation(t *testing.T) {
	tests := []struct {
		name          string
		totalAmount   int64
		amountPaid    int64
		paymentMethod domain.PaymentMethod
		expectError   bool
	}{
		{"Cash exact", 50000, 50000, domain.PaymentMethodCash, false},
		{"Cash overpay", 50000, 100000, domain.PaymentMethodCash, false},
		{"Cash underpay", 50000, 30000, domain.PaymentMethodCash, true},
		{"Kasbon no payment", 50000, 0, domain.PaymentMethodKasbon, false},
		{"Transfer exact", 50000, 50000, domain.PaymentMethodTransfer, false},
		{"QRIS exact", 50000, 50000, domain.PaymentMethodQRIS, false},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePayment(tc.totalAmount, tc.amountPaid, tc.paymentMethod)
			
			hasError := err != nil
			if hasError != tc.expectError {
				t.Errorf("Expected error=%v, got error=%v", tc.expectError, hasError)
			}
		})
	}
}

// TestInvoiceNumberGeneration tests invoice number format
func TestInvoiceNumberGeneration(t *testing.T) {
	// Invoice should be INV-YYYYMMDD-XXXX format
	invoices := []string{
		"INV-20260130-0001",
		"INV-20260130-0002",
		"INV-20260130-9999",
	}
	
	for _, inv := range invoices {
		if !isValidInvoiceNumber(inv) {
			t.Errorf("Invalid invoice format: %s", inv)
		}
	}
}

// Helper functions (simulating service logic)

func calculateCartTotal(repo *MockProductRepository, items []domain.CartItem) int64 {
	var total int64
	ctx := context.Background()
	
	for _, item := range items {
		product, err := repo.GetByID(ctx, item.ProductID)
		if err != nil {
			continue
		}
		
		// Get price based on quantity and pricing tiers
		price := product.BasePrice
		for _, tier := range product.PricingTiers {
			if item.Quantity >= tier.MinQuantity {
				if tier.MaxQuantity == nil || item.Quantity <= *tier.MaxQuantity {
					price = tier.Price
				}
			}
		}
		
		total += price * int64(item.Quantity)
	}
	
	return total
}

func validateStock(product *domain.Product, quantity int) error {
	if quantity <= 0 {
		return domain.ErrEmptyCart
	}
	if product.IsStockActive && product.CurrentStock < quantity {
		return domain.ErrInsufficientStock
	}
	return nil
}

func validateCreditLimit(customer *domain.Customer, amount int64) error {
	if customer.CreditLimit == 0 {
		return nil // No limit
	}
	if customer.CurrentDebt+amount > customer.CreditLimit {
		return domain.ErrCreditLimitExceeded
	}
	return nil
}

func validatePayment(total, paid int64, method domain.PaymentMethod) error {
	if method == domain.PaymentMethodKasbon {
		return nil // Kasbon doesn't require payment
	}
	if paid < total {
		return domain.ErrInvalidPaymentAmount
	}
	return nil
}

func isValidInvoiceNumber(inv string) bool {
	if len(inv) != 18 { // INV-YYYYMMDD-XXXX
		return false
	}
	if inv[:4] != "INV-" {
		return false
	}
	return true
}

func intPtr(i int) *int {
	return &i
}

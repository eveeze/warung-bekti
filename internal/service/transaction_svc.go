package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/repository"
)

// TransactionService handles transaction business logic
type TransactionService struct {
	db              *database.PostgresDB
	transactionRepo *repository.TransactionRepository
	productRepo     *repository.ProductRepository
	customerRepo    *repository.CustomerRepository
	kasbonRepo      *repository.KasbonRepository
	inventoryRepo   *repository.InventoryRepository
	refillableRepo  *repository.RefillableRepository
}

// NewTransactionService creates a new TransactionService
func NewTransactionService(
	db *database.PostgresDB,
	transactionRepo *repository.TransactionRepository,
	productRepo *repository.ProductRepository,
	customerRepo *repository.CustomerRepository,
	kasbonRepo *repository.KasbonRepository,
	inventoryRepo *repository.InventoryRepository,
	refillableRepo *repository.RefillableRepository,
) *TransactionService {
	return &TransactionService{
		db:              db,
		transactionRepo: transactionRepo,
		productRepo:     productRepo,
		customerRepo:    customerRepo,
		kasbonRepo:      kasbonRepo,
		inventoryRepo:   inventoryRepo,
		refillableRepo:  refillableRepo,
	}
}

// CalculateCart calculates cart totals for preview
func (s *TransactionService) CalculateCart(ctx context.Context, input domain.CartCalculateInput) (*domain.CartCalculateResult, error) {
	if len(input.Items) == 0 {
		return nil, domain.ErrEmptyCart
	}

	result := &domain.CartCalculateResult{
		Items:    make([]domain.CartItemResult, 0, len(input.Items)),
		Subtotal: 0,
	}

	for _, item := range input.Items {
		product, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			if err == domain.ErrNotFound {
				continue
			}
			return nil, err
		}

		if !product.IsActive {
			continue
		}

		unitPrice, tierName, _ := product.CalculatePrice(item.Quantity)
		subtotal := unitPrice * int64(item.Quantity)

		itemResult := domain.CartItemResult{
			ProductID:   product.ID,
			ProductName: product.Name,
			Quantity:    item.Quantity,
			Unit:        product.Unit,
			UnitPrice:   unitPrice,
			TierName:    tierName,
			Subtotal:    subtotal,
			IsAvailable: product.CanSell(item.Quantity),
		}

		if product.IsStockActive {
			stock := product.CurrentStock
			itemResult.AvailableQty = &stock
		}

		result.Items = append(result.Items, itemResult)
		result.Subtotal += subtotal
	}

	return result, nil
}

// CreateTransaction processes a checkout
func (s *TransactionService) CreateTransaction(ctx context.Context, input domain.TransactionCreateInput) (*domain.Transaction, error) {
	if len(input.Items) == 0 {
		return nil, domain.ErrEmptyCart
	}

	// Validate customer for kasbon
	if input.PaymentMethod == domain.PaymentMethodKasbon && input.CustomerID == nil {
		return nil, fmt.Errorf("customer is required for kasbon payment")
	}

	var customer *domain.Customer
	if input.CustomerID != nil {
		var err error
		customer, err = s.customerRepo.GetByID(ctx, *input.CustomerID)
		if err != nil {
			return nil, err
		}
		if !customer.IsActive {
			return nil, domain.ErrCustomerInactive
		}
	}

	// Build transaction within a database transaction
	var transaction *domain.Transaction

	err := s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		transaction = &domain.Transaction{
			CustomerID:     input.CustomerID,
			PaymentMethod:  input.PaymentMethod,
			AmountPaid:     input.AmountPaid,
			Status:         domain.TransactionStatusCompleted,
			Notes:          input.Notes,
			CashierName:    input.CashierName,
			Items:          make([]domain.TransactionItem, 0, len(input.Items)),
		}

		var subtotal int64
		var refillableMovements []*domain.ContainerMovement

		// Process each item
		for _, itemInput := range input.Items {
			product, err := s.productRepo.GetByID(ctx, itemInput.ProductID)
			if err != nil {
				return fmt.Errorf("product %s not found", itemInput.ProductID)
			}

			if !product.IsActive {
				return fmt.Errorf("product %s is inactive", product.Name)
			}

			if !product.CanSell(itemInput.Quantity) {
				return fmt.Errorf("insufficient stock for %s (available: %d, requested: %d)",
					product.Name, product.CurrentStock, itemInput.Quantity)
			}

			unitPrice, tierName, tierID := product.CalculatePrice(itemInput.Quantity)
			itemSubtotal := unitPrice * int64(itemInput.Quantity)
			discountAmount := int64(0)
			if itemInput.DiscountAmount != nil {
				discountAmount = *itemInput.DiscountAmount
			}
			totalAmount := itemSubtotal - discountAmount

			item := domain.TransactionItem{
				ProductID:       product.ID,
				ProductName:     product.Name,
				ProductBarcode:  product.Barcode,
				Quantity:        itemInput.Quantity,
				Unit:            product.Unit,
				UnitPrice:       unitPrice,
				CostPrice:       product.CostPrice,
				Subtotal:        itemSubtotal,
				DiscountAmount:  discountAmount,
				TotalAmount:     totalAmount,
				PricingTierID:   tierID,
				Notes:           itemInput.Notes,
			}
			if tierID != nil {
				item.PricingTierName = &tierName
			}

			transaction.Items = append(transaction.Items, item)
			subtotal += totalAmount

			// Deduct stock
			if err := s.inventoryRepo.DeductStock(ctx, tx, product.ID, uuid.Nil, itemInput.Quantity, input.CashierName); err != nil {
				return err
			}

			// Check if Refillable
			container, err := s.refillableRepo.GetByProductID(ctx, product.ID)
			if err != nil {
				// Log error but don't fail transaction? Or fail? 
				// Better to fail if consistency is key.
				return fmt.Errorf("failed to check refillable: %w", err)
			}
			if container != nil {
				// Swap logic: Full -Qty, Empty +Qty
				emptyChange := itemInput.Quantity
				fullChange := -itemInput.Quantity
				
				if err := s.refillableRepo.UpdateContainerStock(ctx, tx, container.ID, emptyChange, fullChange); err != nil {
					return fmt.Errorf("failed to update container stock: %w", err)
				}
				
				refType := "transaction"
				notes := "Sold via POS"

				movement := &domain.ContainerMovement{
					ContainerID:   container.ID,
					Type:          domain.ContainerMovementSaleExchange, // Fix: Use correct Enum
					EmptyChange:   emptyChange,
					FullChange:    fullChange,
					ReferenceType: &refType,
					CreatedBy:     input.CashierName,
					Notes:         &notes,
				}
				
				refillableMovements = append(refillableMovements, movement)
			}
		}

		transaction.Subtotal = subtotal
		transaction.DiscountAmount = 0
		if input.DiscountAmount != nil {
			transaction.DiscountAmount = *input.DiscountAmount
		}
		transaction.TaxAmount = 0
		if input.TaxAmount != nil {
			transaction.TaxAmount = *input.TaxAmount
		}
		transaction.TotalAmount = subtotal - transaction.DiscountAmount + transaction.TaxAmount

		// Calculate change
		if input.PaymentMethod == domain.PaymentMethodCash {
			if input.AmountPaid < transaction.TotalAmount {
				return domain.ErrInvalidPaymentAmount
			}
			transaction.ChangeAmount = input.AmountPaid - transaction.TotalAmount
		}

		// Create transaction record
		if err := s.transactionRepo.Create(ctx, tx, transaction); err != nil {
			return err
		}

		// Handle kasbon
		if input.PaymentMethod == domain.PaymentMethodKasbon {
			// Check credit limit
			if !customer.CanAddDebt(transaction.TotalAmount) {
				return domain.ErrCreditLimitExceeded
			}

			// Create kasbon record
			_, err := s.kasbonRepo.CreateDebt(ctx, tx, *input.CustomerID, &transaction.ID, transaction.TotalAmount, input.Notes, input.CashierName)
			if err != nil {
				return err
			}
		}

		// Record refillable movements now that we have Transaction ID
		for _, m := range refillableMovements {
			m.ReferenceID = &transaction.ID
			if err := s.refillableRepo.RecordMovement(ctx, tx, m); err != nil {
				return fmt.Errorf("failed to record container movement: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload transaction with all relations
	return s.transactionRepo.GetByID(ctx, transaction.ID)
}

// CancelTransaction cancels a transaction
func (s *TransactionService) CancelTransaction(ctx context.Context, id uuid.UUID) error {
	transaction, err := s.transactionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if transaction.Status == domain.TransactionStatusCancelled {
		return domain.ErrTransactionCancelled
	}

	return s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// Restore stock for each item
		for _, item := range transaction.Items {
			product, err := s.productRepo.GetByID(ctx, item.ProductID)
			if err != nil {
				continue
			}

			if product.IsStockActive {
				newStock := product.CurrentStock + item.Quantity
				if err := s.productRepo.UpdateStock(ctx, product.ID, newStock); err != nil {
					return err
				}
			}
		}

		// Update transaction status
		if err := s.transactionRepo.UpdateStatus(ctx, id, domain.TransactionStatusCancelled); err != nil {
			return err
		}

		// Handle kasbon reversal if applicable
		if transaction.PaymentMethod == domain.PaymentMethodKasbon && transaction.CustomerID != nil {
			if err := s.customerRepo.SubtractDebt(ctx, *transaction.CustomerID, transaction.TotalAmount); err != nil {
				return err
			}
		}

		return nil
	})
}

package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/repository"
)

// StockOpnameService handles stock opname business logic
type StockOpnameService struct {
	db           *database.PostgresDB
	opnameRepo   *repository.StockOpnameRepository
	productRepo  *repository.ProductRepository
	inventoryRepo *repository.InventoryRepository
}

// NewStockOpnameService creates a new StockOpnameService
func NewStockOpnameService(
	db *database.PostgresDB,
	opnameRepo *repository.StockOpnameRepository,
	productRepo *repository.ProductRepository,
	inventoryRepo *repository.InventoryRepository,
) *StockOpnameService {
	return &StockOpnameService{
		db:           db,
		opnameRepo:   opnameRepo,
		productRepo:  productRepo,
		inventoryRepo: inventoryRepo,
	}
}

// StartSession starts a new stock opname session
func (s *StockOpnameService) StartSession(ctx context.Context, input domain.StartOpnameInput) (*domain.StockOpnameSession, error) {
	session := &domain.StockOpnameSession{
		Status:    domain.OpnameStatusInProgress,
		Notes:     input.Notes,
		CreatedBy: &input.CreatedBy,
	}

	if err := s.opnameRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// GetSession retrieves a session by ID with items
func (s *StockOpnameService) GetSession(ctx context.Context, id uuid.UUID) (*domain.StockOpnameSession, error) {
	session, err := s.opnameRepo.GetSessionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	items, err := s.opnameRepo.GetSessionItems(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get session items: %w", err)
	}

	session.Items = items
	return session, nil
}

// ListSessions lists all sessions
func (s *StockOpnameService) ListSessions(ctx context.Context, status *domain.OpnameStatus, page, perPage int) ([]domain.StockOpnameSession, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	return s.opnameRepo.ListSessions(ctx, status, perPage, offset)
}

// RecordCount records a physical count for a product
func (s *StockOpnameService) RecordCount(ctx context.Context, input domain.RecordCountInput) (*domain.StockOpnameItem, error) {
	// Verify session exists and is in progress
	session, err := s.opnameRepo.GetSessionByID(ctx, input.SessionID)
	if err != nil {
		return nil, err
	}

	if session.Status != domain.OpnameStatusInProgress {
		return nil, fmt.Errorf("session is not in progress, current status: %s", session.Status)
	}

	// Get product to capture system stock and cost
	product, err := s.productRepo.GetByID(ctx, input.ProductID)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	item := &domain.StockOpnameItem{
		SessionID:     input.SessionID,
		ProductID:     input.ProductID,
		SystemStock:   product.CurrentStock,
		PhysicalStock: input.PhysicalStock,
		CostPerUnit:   product.CostPrice,
		Notes:         input.Notes,
		CountedBy:     &input.CountedBy,
	}

	if err := s.opnameRepo.RecordCount(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to record count: %w", err)
	}

	item.Product = product
	return item, nil
}

// FinalizeSession finalizes a session and optionally applies adjustments
func (s *StockOpnameService) FinalizeSession(ctx context.Context, input domain.FinalizeOpnameInput) (*domain.VarianceReport, error) {
	session, err := s.opnameRepo.GetSessionByID(ctx, input.SessionID)
	if err != nil {
		return nil, err
	}

	if session.Status != domain.OpnameStatusInProgress {
		return nil, fmt.Errorf("session is not in progress")
	}

	// Get all items with variance
	items, err := s.opnameRepo.GetSessionItems(ctx, input.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}

	// Calculate summary
	var totalVariance int
	var totalLossValue, totalGainValue int64

	varianceItems := make([]domain.VarianceReportItem, 0)

	for _, item := range items {
		if item.Variance != 0 {
			totalVariance += abs(item.Variance)
			
			if item.VarianceValue < 0 {
				totalLossValue += abs64(item.VarianceValue)
			} else {
				totalGainValue += item.VarianceValue
			}

			var barcode *string
			var productName string
			if item.Product != nil {
				barcode = item.Product.Barcode
				productName = item.Product.Name
			}

			varianceItems = append(varianceItems, domain.VarianceReportItem{
				ProductID:     item.ProductID,
				ProductName:   productName,
				Barcode:       barcode,
				SystemStock:   item.SystemStock,
				PhysicalStock: item.PhysicalStock,
				Variance:      item.Variance,
				CostPerUnit:   item.CostPerUnit,
				VarianceValue: item.VarianceValue,
				Notes:         item.Notes,
			})
		}
	}

	// Apply adjustments if requested
	if input.ApplyAdjustments {
		err = s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
			for _, item := range items {
				if item.Variance != 0 {
					// Update product stock
					if err := s.productRepo.UpdateStock(ctx, item.ProductID, item.PhysicalStock); err != nil {
						return fmt.Errorf("failed to update stock for %s: %w", item.ProductID, err)
					}

					// Record adjustment movement
					movementType := domain.StockMovementTypeAdjustment
					if err := s.inventoryRepo.RecordMovement(ctx, tx, item.ProductID, movementType, 
						item.Variance, item.SystemStock, item.PhysicalStock, 
						"stock_opname", &input.SessionID, nil, 
						fmt.Sprintf("Stock Opname Adjustment - Session: %s", session.SessionCode),
						&input.CompletedBy); err != nil {
						return fmt.Errorf("failed to record movement: %w", err)
					}
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	// Update session summary
	if err := s.opnameRepo.UpdateSessionSummary(ctx, input.SessionID, len(items), totalVariance, totalLossValue, totalGainValue); err != nil {
		return nil, fmt.Errorf("failed to update summary: %w", err)
	}

	// Update session status
	if err := s.opnameRepo.UpdateSessionStatus(ctx, input.SessionID, domain.OpnameStatusCompleted, &input.CompletedBy); err != nil {
		return nil, fmt.Errorf("failed to update status: %w", err)
	}

	return &domain.VarianceReport{
		SessionID:      input.SessionID,
		SessionCode:    session.SessionCode,
		TotalProducts:  len(items),
		TotalVariance:  totalVariance,
		TotalLossValue: totalLossValue,
		TotalGainValue: totalGainValue,
		NetValue:       totalGainValue - totalLossValue,
		Items:          varianceItems,
	}, nil
}

// GetVarianceReport generates a variance report for a session
func (s *StockOpnameService) GetVarianceReport(ctx context.Context, sessionID uuid.UUID) (*domain.VarianceReport, error) {
	session, err := s.opnameRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	items, err := s.opnameRepo.GetItemsWithVariance(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	varianceItems := make([]domain.VarianceReportItem, len(items))
	for i, item := range items {
		var barcode *string
		var productName string
		if item.Product != nil {
			barcode = item.Product.Barcode
			productName = item.Product.Name
		}

		varianceItems[i] = domain.VarianceReportItem{
			ProductID:     item.ProductID,
			ProductName:   productName,
			Barcode:       barcode,
			SystemStock:   item.SystemStock,
			PhysicalStock: item.PhysicalStock,
			Variance:      item.Variance,
			CostPerUnit:   item.CostPerUnit,
			VarianceValue: item.VarianceValue,
			Notes:         item.Notes,
		}
	}

	return &domain.VarianceReport{
		SessionID:      sessionID,
		SessionCode:    session.SessionCode,
		TotalProducts:  session.TotalProducts,
		TotalVariance:  session.TotalVariance,
		TotalLossValue: session.TotalLossValue,
		TotalGainValue: session.TotalGainValue,
		NetValue:       session.TotalGainValue - session.TotalLossValue,
		Items:          varianceItems,
	}, nil
}

// GetShoppingList generates a shopping list from low stock products
func (s *StockOpnameService) GetShoppingList(ctx context.Context) (*domain.ShoppingList, error) {
	items, err := s.opnameRepo.GetShoppingList(ctx)
	if err != nil {
		return nil, err
	}

	var totalCost int64
	for _, item := range items {
		if item.EstimatedCost != nil {
			totalCost += *item.EstimatedCost
		}
	}

	return &domain.ShoppingList{
		GeneratedAt: time.Now(),
		TotalItems:  len(items),
		TotalCost:   totalCost,
		Items:       items,
	}, nil
}

// GetNearExpiryReport generates a report of items nearing expiry
func (s *StockOpnameService) GetNearExpiryReport(ctx context.Context, daysAhead int) (*domain.NearExpiryReport, error) {
	if daysAhead <= 0 {
		daysAhead = 30 // Default to 30 days
	}

	items, err := s.opnameRepo.GetNearExpiryItems(ctx, daysAhead)
	if err != nil {
		return nil, err
	}

	var totalValue int64
	for _, item := range items {
		totalValue += item.CostPrice * int64(item.Quantity)
	}

	return &domain.NearExpiryReport{
		GeneratedAt: time.Now(),
		DaysAhead:   daysAhead,
		TotalItems:  len(items),
		TotalValue:  totalValue,
		Items:       items,
	}, nil
}

// CancelSession cancels an in-progress session
func (s *StockOpnameService) CancelSession(ctx context.Context, sessionID uuid.UUID) error {
	session, err := s.opnameRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return err
	}

	if session.Status == domain.OpnameStatusCompleted {
		return fmt.Errorf("cannot cancel a completed session")
	}

	return s.opnameRepo.UpdateSessionStatus(ctx, sessionID, domain.OpnameStatusCancelled, nil)
}

// Helper functions
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

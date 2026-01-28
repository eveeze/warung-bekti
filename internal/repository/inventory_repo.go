package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
)

// InventoryRepository handles stock movement database operations
type InventoryRepository struct {
	db *database.PostgresDB
}

// NewInventoryRepository creates a new InventoryRepository
func NewInventoryRepository(db *database.PostgresDB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

// CreateMovement creates a stock movement record
func (r *InventoryRepository) CreateMovement(ctx context.Context, tx *sql.Tx, movement *domain.StockMovement) error {
	query := `
		INSERT INTO stock_movements (product_id, type, quantity, stock_before, stock_after, reference_type, reference_id, cost_per_unit, notes, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`

	var err error
	if tx != nil {
		err = tx.QueryRowContext(ctx, query,
			movement.ProductID, movement.Type, movement.Quantity, movement.StockBefore, movement.StockAfter,
			movement.ReferenceType, movement.ReferenceID, movement.CostPerUnit, movement.Notes, movement.CreatedBy,
		).Scan(&movement.ID, &movement.CreatedAt)
	} else {
		err = r.db.QueryRowContext(ctx, query,
			movement.ProductID, movement.Type, movement.Quantity, movement.StockBefore, movement.StockAfter,
			movement.ReferenceType, movement.ReferenceID, movement.CostPerUnit, movement.Notes, movement.CreatedBy,
		).Scan(&movement.ID, &movement.CreatedAt)
	}
	return err
}

// GetByProduct retrieves stock movements for a product
func (r *InventoryRepository) GetByProduct(ctx context.Context, productID uuid.UUID, filter domain.StockMovementFilter) ([]domain.StockMovement, int64, error) {
	args := []interface{}{productID}
	whereClause := "WHERE product_id = $1"
	argIndex := 2

	if filter.Type != nil {
		whereClause += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, *filter.Type)
		argIndex++
	}
	if filter.DateFrom != nil {
		whereClause += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *filter.DateFrom)
		argIndex++
	}
	if filter.DateTo != nil {
		whereClause += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *filter.DateTo)
		argIndex++
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM stock_movements %s", whereClause), args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	page, perPage := filter.Page, filter.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	query := fmt.Sprintf(`
		SELECT id, product_id, type, quantity, stock_before, stock_after, reference_type, reference_id, cost_per_unit, notes, created_by, created_at
		FROM stock_movements %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, perPage, (page-1)*perPage)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var movements []domain.StockMovement
	for rows.Next() {
		var m domain.StockMovement
		if err := rows.Scan(
			&m.ID, &m.ProductID, &m.Type, &m.Quantity, &m.StockBefore, &m.StockAfter,
			&m.ReferenceType, &m.ReferenceID, &m.CostPerUnit, &m.Notes, &m.CreatedBy, &m.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		movements = append(movements, m)
	}
	return movements, total, rows.Err()
}

// Restock adds stock to a product
func (r *InventoryRepository) Restock(ctx context.Context, input domain.RestockInput) (*domain.StockMovement, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var currentStock int
	err = tx.QueryRowContext(ctx, "SELECT current_stock FROM products WHERE id = $1", input.ProductID).Scan(&currentStock)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	newStock := currentStock + input.Quantity

	movement := &domain.StockMovement{
		ProductID:   input.ProductID,
		Type:        domain.StockMovementTypePurchase,
		Quantity:    input.Quantity,
		StockBefore: currentStock,
		StockAfter:  newStock,
		CostPerUnit: &input.CostPerUnit,
		Notes:       input.Notes,
		CreatedBy:   input.CreatedBy,
	}

	if err := r.CreateMovement(ctx, tx, movement); err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, "UPDATE products SET current_stock = $1, updated_at = NOW() WHERE id = $2", newStock, input.ProductID); err != nil {
		return nil, err
	}

	return movement, tx.Commit()
}

// Adjust adjusts stock manually
func (r *InventoryRepository) Adjust(ctx context.Context, input domain.StockAdjustmentInput) (*domain.StockMovement, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var currentStock int
	err = tx.QueryRowContext(ctx, "SELECT current_stock FROM products WHERE id = $1", input.ProductID).Scan(&currentStock)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	newStock := currentStock + input.Quantity
	if newStock < 0 {
		newStock = 0
	}

	movement := &domain.StockMovement{
		ProductID:   input.ProductID,
		Type:        domain.StockMovementTypeAdjustment,
		Quantity:    input.Quantity,
		StockBefore: currentStock,
		StockAfter:  newStock,
		Notes:       &input.Reason,
		CreatedBy:   input.CreatedBy,
	}

	if err := r.CreateMovement(ctx, tx, movement); err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, "UPDATE products SET current_stock = $1, updated_at = NOW() WHERE id = $2", newStock, input.ProductID); err != nil {
		return nil, err
	}

	return movement, tx.Commit()
}

// DeductStock deducts stock for a sale (used within transaction)
func (r *InventoryRepository) DeductStock(ctx context.Context, tx *sql.Tx, productID, transactionID uuid.UUID, quantity int, createdBy *string) error {
	var currentStock int
	var isStockActive bool
	err := tx.QueryRowContext(ctx, "SELECT current_stock, is_stock_active FROM products WHERE id = $1", productID).Scan(&currentStock, &isStockActive)
	if err != nil {
		return err
	}

	if !isStockActive {
		return nil // Skip for non-tracked products
	}

	if currentStock < quantity {
		return domain.ErrInsufficientStock
	}

	newStock := currentStock - quantity
	refType := "transaction"

	movement := &domain.StockMovement{
		ProductID:     productID,
		Type:          domain.StockMovementTypeSale,
		Quantity:      -quantity,
		StockBefore:   currentStock,
		StockAfter:    newStock,
		ReferenceType: &refType,
		ReferenceID:   &transactionID,
		CreatedBy:     createdBy,
	}

	if err := r.CreateMovement(ctx, tx, movement); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE products SET current_stock = $1, updated_at = NOW() WHERE id = $2", newStock, productID)
	return err
}

// GetStockReport returns stock inventory report
func (r *InventoryRepository) GetStockReport(ctx context.Context) (*domain.StockReport, error) {
	query := `
		SELECT COUNT(*), 
			COALESCE(SUM(current_stock * cost_price), 0),
			COUNT(*) FILTER (WHERE is_stock_active AND current_stock <= min_stock_alert),
			COUNT(*) FILTER (WHERE is_stock_active AND current_stock = 0)
		FROM products WHERE is_active = true
	`

	var report domain.StockReport
	err := r.db.QueryRowContext(ctx, query).Scan(
		&report.TotalProducts, &report.TotalStockValue, &report.LowStockCount, &report.OutOfStockCount,
	)
	return &report, err
}

package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
)

// StockOpnameRepository handles stock opname data access
type StockOpnameRepository struct {
	db *database.PostgresDB
}

// NewStockOpnameRepository creates a new StockOpnameRepository
func NewStockOpnameRepository(db *database.PostgresDB) *StockOpnameRepository {
	return &StockOpnameRepository{db: db}
}

// CreateSession creates a new opname session
func (r *StockOpnameRepository) CreateSession(ctx context.Context, session *domain.StockOpnameSession) error {
	query := `
		INSERT INTO stock_opname_sessions (session_code, status, notes, created_by, started_at)
		VALUES (generate_opname_session_code(), $1, $2, $3, $4)
		RETURNING id, session_code, created_at, updated_at
	`

	now := time.Now()
	return r.db.QueryRowContext(ctx, query,
		session.Status,
		session.Notes,
		session.CreatedBy,
		now,
	).Scan(&session.ID, &session.SessionCode, &session.CreatedAt, &session.UpdatedAt)
}

// GetSessionByID retrieves a session by ID
func (r *StockOpnameRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*domain.StockOpnameSession, error) {
	query := `
		SELECT id, session_code, status, notes, total_products, total_variance,
			total_loss_value, total_gain_value, started_at, completed_at,
			created_by, completed_by, created_at, updated_at
		FROM stock_opname_sessions
		WHERE id = $1
	`

	session := &domain.StockOpnameSession{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID,
		&session.SessionCode,
		&session.Status,
		&session.Notes,
		&session.TotalProducts,
		&session.TotalVariance,
		&session.TotalLossValue,
		&session.TotalGainValue,
		&session.StartedAt,
		&session.CompletedAt,
		&session.CreatedBy,
		&session.CompletedBy,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return session, nil
}

// ListSessions lists all opname sessions with optional filtering
func (r *StockOpnameRepository) ListSessions(ctx context.Context, status *domain.OpnameStatus, limit, offset int) ([]domain.StockOpnameSession, int64, error) {
	countQuery := `SELECT COUNT(*) FROM stock_opname_sessions WHERE ($1::opname_status IS NULL OR status = $1)`
	
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, status).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, session_code, status, notes, total_products, total_variance,
			total_loss_value, total_gain_value, started_at, completed_at,
			created_by, completed_by, created_at, updated_at
		FROM stock_opname_sessions
		WHERE ($1::opname_status IS NULL OR status = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, status, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var sessions []domain.StockOpnameSession
	for rows.Next() {
		var s domain.StockOpnameSession
		if err := rows.Scan(
			&s.ID, &s.SessionCode, &s.Status, &s.Notes,
			&s.TotalProducts, &s.TotalVariance, &s.TotalLossValue, &s.TotalGainValue,
			&s.StartedAt, &s.CompletedAt, &s.CreatedBy, &s.CompletedBy,
			&s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		sessions = append(sessions, s)
	}

	return sessions, total, rows.Err()
}

// UpdateSessionStatus updates the session status
func (r *StockOpnameRepository) UpdateSessionStatus(ctx context.Context, id uuid.UUID, status domain.OpnameStatus, completedBy *string) error {
	var query string
	var args []interface{}

	if status == domain.OpnameStatusCompleted {
		query = `
			UPDATE stock_opname_sessions
			SET status = $2, completed_by = $3, completed_at = NOW()
			WHERE id = $1
		`
		args = []interface{}{id, status, completedBy}
	} else {
		query = `UPDATE stock_opname_sessions SET status = $2 WHERE id = $1`
		args = []interface{}{id, status}
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// UpdateSessionSummary updates the session summary after finalization
func (r *StockOpnameRepository) UpdateSessionSummary(ctx context.Context, id uuid.UUID, totalProducts, totalVariance int, lossValue, gainValue int64) error {
	query := `
		UPDATE stock_opname_sessions
		SET total_products = $2, total_variance = $3, total_loss_value = $4, total_gain_value = $5
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id, totalProducts, totalVariance, lossValue, gainValue)
	return err
}

// RecordCount records or updates a physical count for a product
func (r *StockOpnameRepository) RecordCount(ctx context.Context, item *domain.StockOpnameItem) error {
	query := `
		INSERT INTO stock_opname_items (session_id, product_id, system_stock, physical_stock, cost_per_unit, notes, counted_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (session_id, product_id) DO UPDATE
		SET physical_stock = EXCLUDED.physical_stock, notes = EXCLUDED.notes, 
			counted_by = EXCLUDED.counted_by, counted_at = NOW()
		RETURNING id, variance, variance_value, counted_at
	`

	return r.db.QueryRowContext(ctx, query,
		item.SessionID,
		item.ProductID,
		item.SystemStock,
		item.PhysicalStock,
		item.CostPerUnit,
		item.Notes,
		item.CountedBy,
	).Scan(&item.ID, &item.Variance, &item.VarianceValue, &item.CountedAt)
}

// GetSessionItems gets all items for a session
func (r *StockOpnameRepository) GetSessionItems(ctx context.Context, sessionID uuid.UUID) ([]domain.StockOpnameItem, error) {
	query := `
		SELECT soi.id, soi.session_id, soi.product_id, soi.system_stock, soi.physical_stock,
			soi.variance, soi.cost_per_unit, soi.variance_value, soi.notes, soi.counted_by, soi.counted_at,
			p.name as product_name, p.barcode
		FROM stock_opname_items soi
		JOIN products p ON p.id = soi.product_id
		WHERE soi.session_id = $1
		ORDER BY p.name
	`

	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.StockOpnameItem
	for rows.Next() {
		var item domain.StockOpnameItem
		var productName string
		var barcode *string

		if err := rows.Scan(
			&item.ID, &item.SessionID, &item.ProductID, &item.SystemStock, &item.PhysicalStock,
			&item.Variance, &item.CostPerUnit, &item.VarianceValue, &item.Notes, &item.CountedBy, &item.CountedAt,
			&productName, &barcode,
		); err != nil {
			return nil, err
		}

		item.Product = &domain.Product{
			ID:      item.ProductID,
			Name:    productName,
			Barcode: barcode,
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// GetItemsWithVariance gets items with non-zero variance
func (r *StockOpnameRepository) GetItemsWithVariance(ctx context.Context, sessionID uuid.UUID) ([]domain.StockOpnameItem, error) {
	query := `
		SELECT soi.id, soi.session_id, soi.product_id, soi.system_stock, soi.physical_stock,
			soi.variance, soi.cost_per_unit, soi.variance_value, soi.notes, soi.counted_by, soi.counted_at,
			p.name as product_name, p.barcode
		FROM stock_opname_items soi
		JOIN products p ON p.id = soi.product_id
		WHERE soi.session_id = $1 AND soi.variance != 0
		ORDER BY ABS(soi.variance_value) DESC
	`

	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.StockOpnameItem
	for rows.Next() {
		var item domain.StockOpnameItem
		var productName string
		var barcode *string

		if err := rows.Scan(
			&item.ID, &item.SessionID, &item.ProductID, &item.SystemStock, &item.PhysicalStock,
			&item.Variance, &item.CostPerUnit, &item.VarianceValue, &item.Notes, &item.CountedBy, &item.CountedAt,
			&productName, &barcode,
		); err != nil {
			return nil, err
		}

		item.Product = &domain.Product{
			ID:      item.ProductID,
			Name:    productName,
			Barcode: barcode,
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// GetShoppingList generates a shopping list from low stock products
func (r *StockOpnameRepository) GetShoppingList(ctx context.Context) ([]domain.ShoppingListItem, error) {
	query := `
		SELECT p.id, p.name, p.barcode, p.current_stock, p.min_stock_alert, p.cost_price,
			COALESCE(p.max_stock, p.min_stock_alert * 3) - p.current_stock as suggested_qty
		FROM products p
		WHERE p.is_active = true 
			AND p.is_stock_active = true 
			AND p.current_stock <= p.min_stock_alert
		ORDER BY (p.current_stock::float / NULLIF(p.min_stock_alert, 0)) ASC, p.name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.ShoppingListItem
	for rows.Next() {
		var item domain.ShoppingListItem
		var productName string
		var barcode *string
		var costPrice int64

		if err := rows.Scan(
			&item.ProductID, &productName, &barcode,
			&item.CurrentStock, &item.MinStock, &costPrice, &item.SuggestedQty,
		); err != nil {
			return nil, err
		}

		estimatedCost := int64(item.SuggestedQty) * costPrice
		item.EstimatedCost = &estimatedCost
		item.Product = &domain.Product{
			ID:      item.ProductID,
			Name:    productName,
			Barcode: barcode,
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// GetNearExpiryItems gets stock movements with expiry dates within the given days
func (r *StockOpnameRepository) GetNearExpiryItems(ctx context.Context, daysAhead int) ([]domain.NearExpiryItem, error) {
	query := `
		SELECT sm.product_id, p.name, p.barcode, sm.batch_number, sm.expiry_date,
			(sm.expiry_date - CURRENT_DATE) as days_until_expiry,
			sm.quantity, p.cost_price
		FROM stock_movements sm
		JOIN products p ON p.id = sm.product_id
		WHERE sm.expiry_date IS NOT NULL 
			AND sm.expiry_date <= CURRENT_DATE + ($1 * INTERVAL '1 day')
			AND sm.expiry_date >= CURRENT_DATE
			AND sm.type IN ('purchase', 'initial')
		ORDER BY sm.expiry_date ASC, p.name
	`

	rows, err := r.db.QueryContext(ctx, query, daysAhead)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.NearExpiryItem
	for rows.Next() {
		var item domain.NearExpiryItem
		if err := rows.Scan(
			&item.ProductID, &item.ProductName, &item.Barcode, &item.BatchNumber,
			&item.ExpiryDate, &item.DaysUntilExpiry, &item.Quantity, &item.CostPrice,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
)

type POSRepository struct {
	db *database.PostgresDB
}

func NewPOSRepository(db *database.PostgresDB) *POSRepository {
	return &POSRepository{db: db}
}

// -- Held Carts --

func (r *POSRepository) HoldCart(ctx context.Context, cart *domain.HeldCart) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create Cart
	query := `
		INSERT INTO held_carts (hold_code, customer_id, customer_name, status, subtotal, notes, held_by, held_at, expires_at)
		VALUES (generate_hold_code(), $1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, hold_code, created_at, updated_at
	`
	err = tx.QueryRowContext(ctx, query,
		cart.CustomerID, cart.CustomerName, cart.Status, cart.Subtotal, cart.Notes, cart.HeldBy, cart.HeldAt, cart.ExpiresAt,
	).Scan(&cart.ID, &cart.HoldCode, &cart.CreatedAt, &cart.UpdatedAt)
	if err != nil {
		return err
	}

	// Create Items
	itemQuery := `
		INSERT INTO held_cart_items (cart_id, product_id, product_name, product_barcode, quantity, unit, unit_price, subtotal, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`
	for i := range cart.Items {
		item := &cart.Items[i]
		item.CartID = cart.ID
		err = tx.QueryRowContext(ctx, itemQuery,
			item.CartID, item.ProductID, item.ProductName, item.ProductBarcode,
			item.Quantity, item.Unit, item.UnitPrice, item.Subtotal, item.Notes,
		).Scan(&item.ID, &item.CreatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *POSRepository) GetHeldCart(ctx context.Context, id uuid.UUID) (*domain.HeldCart, error) {
	query := `
		SELECT id, hold_code, customer_id, customer_name, status, subtotal, notes, held_by, resumed_by, held_at, resumed_at, expires_at, created_at, updated_at
		FROM held_carts WHERE id = $1
	`
	var cart domain.HeldCart
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&cart.ID, &cart.HoldCode, &cart.CustomerID, &cart.CustomerName, &cart.Status, &cart.Subtotal, &cart.Notes,
		&cart.HeldBy, &cart.ResumedBy, &cart.HeldAt, &cart.ResumedAt, &cart.ExpiresAt, &cart.CreatedAt, &cart.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Get Items
	itemQuery := `
		SELECT id, cart_id, product_id, product_name, product_barcode, quantity, unit, unit_price, subtotal, notes, created_at
		FROM held_cart_items WHERE cart_id = $1
	`
	rows, err := r.db.QueryContext(ctx, itemQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.HeldCartItem
		if err := rows.Scan(
			&item.ID, &item.CartID, &item.ProductID, &item.ProductName, &item.ProductBarcode,
			&item.Quantity, &item.Unit, &item.UnitPrice, &item.Subtotal, &item.Notes, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		cart.Items = append(cart.Items, item)
	}

	return &cart, nil
}

func (r *POSRepository) ListHeldCarts(ctx context.Context) ([]domain.HeldCart, error) {
	query := `
		SELECT id, hold_code, customer_id, customer_name, status, subtotal, notes, held_by, held_at, expires_at, created_at
		FROM held_carts 
		WHERE status = 'held' 
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var carts []domain.HeldCart
	for rows.Next() {
		var c domain.HeldCart
		if err := rows.Scan(
			&c.ID, &c.HoldCode, &c.CustomerID, &c.CustomerName, &c.Status, &c.Subtotal, &c.Notes,
			&c.HeldBy, &c.HeldAt, &c.ExpiresAt, &c.CreatedAt,
		); err != nil {
			return nil, err
		}
		carts = append(carts, c)
	}
	return carts, nil
}

func (r *POSRepository) UpdateCartStatus(ctx context.Context, id uuid.UUID, status domain.HeldCartStatus, by *string) error {
	var query string
	if status == domain.HeldCartStatusResumed {
		query = `UPDATE held_carts SET status = $2, resumed_by = $3, resumed_at = NOW(), updated_at = NOW() WHERE id = $1`
	} else {
		query = `UPDATE held_carts SET status = $2, updated_at = NOW() WHERE id = $1`
	}
	
	args := []interface{}{id, status}
	if status == domain.HeldCartStatusResumed {
		args = append(args, by)
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// -- Refunds --

func (r *POSRepository) CreateRefund(ctx context.Context, refund *domain.RefundRecord) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO refund_records (refund_number, transaction_id, customer_id, total_refund_amount, refund_method, status, reason, notes, requested_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		RETURNING id, refund_number, created_at, updated_at
	`
	err = tx.QueryRowContext(ctx, query,
		refund.RefundNumber, refund.TransactionID, refund.CustomerID, refund.TotalRefundAmount, refund.RefundMethod, refund.Status,
		refund.Reason, refund.Notes, refund.RequestedBy,
	).Scan(&refund.ID, &refund.RefundNumber, &refund.CreatedAt, &refund.UpdatedAt)
	if err != nil {
		return err
	}

	itemQuery := `
		INSERT INTO refund_items (refund_id, transaction_item_id, product_id, product_name, quantity, unit_price, refund_amount, reason, restock)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`
	for i := range refund.Items {
		item := &refund.Items[i]
		item.RefundID = refund.ID
		err = tx.QueryRowContext(ctx, itemQuery,
			item.RefundID, item.TransactionItemID, item.ProductID, item.ProductName,
			item.Quantity, item.UnitPrice, item.RefundAmount, item.Reason, item.Restock,
		).Scan(&item.ID, &item.CreatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *POSRepository) GetRefund(ctx context.Context, id uuid.UUID) (*domain.RefundRecord, error) {
	query := `
		SELECT id, refund_number, transaction_id, customer_id, total_refund_amount, refund_method, status, reason, notes, requested_by, approved_by, completed_at, created_at, updated_at
		FROM refund_records WHERE id = $1
	`
	var refund domain.RefundRecord
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&refund.ID, &refund.RefundNumber, &refund.TransactionID, &refund.CustomerID, &refund.TotalRefundAmount,
		&refund.RefundMethod, &refund.Status, &refund.Reason, &refund.Notes, &refund.RequestedBy, &refund.ApprovedBy,
		&refund.CompletedAt, &refund.CreatedAt, &refund.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Items
	itemQuery := `
		SELECT id, refund_id, transaction_item_id, product_id, product_name, quantity, unit_price, refund_amount, reason, restock, created_at
		FROM refund_items WHERE refund_id = $1
	`
	rows, err := r.db.QueryContext(ctx, itemQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.RefundItem
		if err := rows.Scan(
			&item.ID, &item.RefundID, &item.TransactionItemID, &item.ProductID, &item.ProductName,
			&item.Quantity, &item.UnitPrice, &item.RefundAmount, &item.Reason, &item.Restock, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		refund.Items = append(refund.Items, item)
	}
	return &refund, nil
}

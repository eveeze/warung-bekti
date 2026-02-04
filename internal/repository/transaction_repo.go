package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
)

// TransactionRepository handles transaction database operations
type TransactionRepository struct {
	db *database.PostgresDB
}

// NewTransactionRepository creates a new TransactionRepository
func NewTransactionRepository(db *database.PostgresDB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Create creates a new transaction with items
func (r *TransactionRepository) Create(ctx context.Context, tx *sql.Tx, transaction *domain.Transaction) error {
	var invoiceNumber string
	err := tx.QueryRowContext(ctx, "SELECT generate_invoice_number()").Scan(&invoiceNumber)
	if err != nil {
		return fmt.Errorf("failed to generate invoice number: %w", err)
	}
	transaction.InvoiceNumber = invoiceNumber

	query := `
		INSERT INTO transactions (
			invoice_number, customer_id, subtotal, discount_amount, tax_amount,
			total_amount, payment_method, amount_paid, change_amount, status, notes, cashier_name
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`

	err = tx.QueryRowContext(ctx, query,
		transaction.InvoiceNumber, transaction.CustomerID, transaction.Subtotal,
		transaction.DiscountAmount, transaction.TaxAmount, transaction.TotalAmount,
		transaction.PaymentMethod, transaction.AmountPaid, transaction.ChangeAmount,
		transaction.Status, transaction.Notes, transaction.CashierName,
	).Scan(&transaction.ID, &transaction.CreatedAt, &transaction.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	for i := range transaction.Items {
		item := &transaction.Items[i]
		item.TransactionID = transaction.ID

		itemQuery := `
			INSERT INTO transaction_items (
				transaction_id, product_id, product_name, product_barcode,
				quantity, unit, unit_price, cost_price, subtotal,
				discount_amount, total_amount, pricing_tier_id, pricing_tier_name, notes
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			RETURNING id, created_at
		`

		err = tx.QueryRowContext(ctx, itemQuery,
			item.TransactionID, item.ProductID, item.ProductName, item.ProductBarcode,
			item.Quantity, item.Unit, item.UnitPrice, item.CostPrice, item.Subtotal,
			item.DiscountAmount, item.TotalAmount, item.PricingTierID, item.PricingTierName, item.Notes,
		).Scan(&item.ID, &item.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to create transaction item: %w", err)
		}
	}

	return nil
}

// GetByID retrieves a transaction by ID with items
func (r *TransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	query := `
		SELECT t.id, t.invoice_number, t.customer_id, t.subtotal, t.discount_amount, t.tax_amount,
			t.total_amount, t.payment_method, t.amount_paid, t.change_amount, t.status, t.notes, t.cashier_name,
			t.created_at, t.updated_at, c.name
		FROM transactions t
		LEFT JOIN customers c ON t.customer_id = c.id
		WHERE t.id = $1
	`

	var t domain.Transaction
	var customerName *string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.InvoiceNumber, &t.CustomerID, &t.Subtotal, &t.DiscountAmount,
		&t.TaxAmount, &t.TotalAmount, &t.PaymentMethod, &t.AmountPaid, &t.ChangeAmount,
		&t.Status, &t.Notes, &t.CashierName, &t.CreatedAt, &t.UpdatedAt, &customerName,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	if t.CustomerID != nil && customerName != nil {
		t.Customer = &domain.Customer{
			ID:   *t.CustomerID,
			Name: *customerName,
		}
	}

	t.Items, _ = r.GetItems(ctx, t.ID)
	return &t, nil
}

// GetItems retrieves transaction items
func (r *TransactionRepository) GetItems(ctx context.Context, transactionID uuid.UUID) ([]domain.TransactionItem, error) {
	query := `
		SELECT id, transaction_id, product_id, product_name, product_barcode,
			quantity, unit, unit_price, cost_price, subtotal, discount_amount,
			total_amount, pricing_tier_id, pricing_tier_name, notes, created_at
		FROM transaction_items WHERE transaction_id = $1 ORDER BY created_at
	`

	rows, err := r.db.QueryContext(ctx, query, transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.TransactionItem
	for rows.Next() {
		var item domain.TransactionItem
		if err := rows.Scan(
			&item.ID, &item.TransactionID, &item.ProductID, &item.ProductName,
			&item.ProductBarcode, &item.Quantity, &item.Unit, &item.UnitPrice,
			&item.CostPrice, &item.Subtotal, &item.DiscountAmount, &item.TotalAmount,
			&item.PricingTierID, &item.PricingTierName, &item.Notes, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// List retrieves transactions with filtering
func (r *TransactionRepository) List(ctx context.Context, filter domain.TransactionFilter) ([]domain.Transaction, int64, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.Search != nil {
		conditions = append(conditions, fmt.Sprintf("(t.invoice_number ILIKE $%d OR c.name ILIKE $%d)", argIndex, argIndex))
		searchPattern := "%" + *filter.Search + "%"
		args = append(args, searchPattern)
		argIndex++
	}

	if filter.CustomerID != nil {
		conditions = append(conditions, fmt.Sprintf("t.customer_id = $%d", argIndex))
		args = append(args, *filter.CustomerID)
		argIndex++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("t.status = $%d", argIndex))
		args = append(args, *filter.Status)
		argIndex++
	}
	if filter.PaymentMethod != nil {
		conditions = append(conditions, fmt.Sprintf("t.payment_method = $%d", argIndex))
		args = append(args, *filter.PaymentMethod)
		argIndex++
	}
	if filter.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("t.created_at >= $%d", argIndex))
		args = append(args, *filter.DateFrom)
		argIndex++
	}
	if filter.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("t.created_at <= $%d", argIndex))
		args = append(args, *filter.DateTo)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM transactions t 
		LEFT JOIN customers c ON t.customer_id = c.id
		%s
	`, whereClause)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
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
		SELECT t.id, t.invoice_number, t.customer_id, t.subtotal, t.discount_amount, t.tax_amount,
			t.total_amount, t.payment_method, t.amount_paid, t.change_amount, t.status, t.notes, t.cashier_name,
			t.created_at, t.updated_at, c.name
		FROM transactions t
		LEFT JOIN customers c ON t.customer_id = c.id
		%s ORDER BY t.created_at DESC LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, perPage, (page-1)*perPage)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []domain.Transaction
	for rows.Next() {
		var t domain.Transaction
		var customerName *string
		if err := rows.Scan(
			&t.ID, &t.InvoiceNumber, &t.CustomerID, &t.Subtotal, &t.DiscountAmount,
			&t.TaxAmount, &t.TotalAmount, &t.PaymentMethod, &t.AmountPaid, &t.ChangeAmount,
			&t.Status, &t.Notes, &t.CashierName, &t.CreatedAt, &t.UpdatedAt, &customerName,
		); err != nil {
			return nil, 0, err
		}
		if t.CustomerID != nil && customerName != nil {
			t.Customer = &domain.Customer{
				ID:   *t.CustomerID,
				Name: *customerName,
			}
		}
		transactions = append(transactions, t)
	}
	return transactions, total, rows.Err()
}

// UpdateStatus updates transaction status
func (r *TransactionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.TransactionStatus) error {
	query := `UPDATE transactions SET status = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return err
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// GetDailySales returns total sales for a date
func (r *TransactionRepository) GetDailySales(ctx context.Context, date string) (int64, int, error) {
	query := `SELECT COALESCE(SUM(total_amount), 0), COUNT(*) FROM transactions 
		WHERE DATE(created_at) = $1 AND status = 'completed'`
	var total int64
	var count int
	err := r.db.QueryRowContext(ctx, query, date).Scan(&total, &count)
	return total, count, err
}

// GetDailyProfit returns profit for a date
func (r *TransactionRepository) GetDailyProfit(ctx context.Context, date string) (int64, error) {
	query := `SELECT COALESCE(SUM(ti.total_amount - (ti.cost_price * ti.quantity)), 0)
		FROM transaction_items ti JOIN transactions t ON t.id = ti.transaction_id
		WHERE DATE(t.created_at) = $1 AND t.status = 'completed'`
	var profit int64
	err := r.db.QueryRowContext(ctx, query, date).Scan(&profit)
	return profit, err
}

// GetHourlySales returns sales grouped by hour for a specific date
func (r *TransactionRepository) GetHourlySales(ctx context.Context, date string) ([]map[string]interface{}, error) {
	query := `
		SELECT EXTRACT(HOUR FROM created_at) as hour, 
		       COALESCE(SUM(total_amount), 0) as sales, 
		       COUNT(*) as transactions
		FROM transactions 
		WHERE DATE(created_at) = $1 AND status = 'completed'
		GROUP BY hour
		ORDER BY hour ASC
	`
	rows, err := r.db.QueryContext(ctx, query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var hour float64
		var sales int64
		var count int
		if err := rows.Scan(&hour, &sales, &count); err != nil {
			continue
		}
		result = append(result, map[string]interface{}{
			"hour":         int(hour),
			"sales":        sales,
			"transactions": count,
		})
	}
	return result, nil
}

// GetTopProducts returns top selling products by sales amount
func (r *TransactionRepository) GetTopProducts(ctx context.Context, date string, limit int) ([]map[string]interface{}, error) {
	query := `
		SELECT ti.product_id, ti.product_name, 
		       SUM(ti.quantity) as total_quantity, 
		       SUM(ti.total_amount) as total_sales
		FROM transaction_items ti
		JOIN transactions t ON t.id = ti.transaction_id
		WHERE DATE(t.created_at) = $1 AND t.status = 'completed'
		GROUP BY ti.product_id, ti.product_name
		ORDER BY total_sales DESC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, date, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var productID uuid.UUID
		var productName string
		var qty int
		var sales int64
		if err := rows.Scan(&productID, &productName, &qty, &sales); err != nil {
			continue
		}
		result = append(result, map[string]interface{}{
			"product_id":     productID,
			"product_name":   productName,
			"total_quantity": qty,
			"total_sales":    sales,
		})
	}
	return result, nil
}

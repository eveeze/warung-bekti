package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
)

// KasbonRepository handles kasbon database operations
type KasbonRepository struct {
	db *database.PostgresDB
}

// NewKasbonRepository creates a new KasbonRepository
func NewKasbonRepository(db *database.PostgresDB) *KasbonRepository {
	return &KasbonRepository{db: db}
}

// CreateDebt creates a new debt record
func (r *KasbonRepository) CreateDebt(ctx context.Context, tx *sql.Tx, customerID uuid.UUID, transactionID *uuid.UUID, amount int64, notes *string, createdBy *string) (*domain.KasbonRecord, error) {
	// Get current balance
	var currentDebt int64
	err := tx.QueryRowContext(ctx, "SELECT current_debt FROM customers WHERE id = $1", customerID).Scan(&currentDebt)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer debt: %w", err)
	}

	newBalance := currentDebt + amount

	query := `
		INSERT INTO kasbon_records (customer_id, transaction_id, type, amount, balance_before, balance_after, notes, created_by)
		VALUES ($1, $2, 'debt', $3, $4, $5, $6, $7)
		RETURNING id, customer_id, transaction_id, type, amount, balance_before, balance_after, notes, created_by, created_at
	`

	var record domain.KasbonRecord
	err = tx.QueryRowContext(ctx, query, customerID, transactionID, amount, currentDebt, newBalance, notes, createdBy).Scan(
		&record.ID, &record.CustomerID, &record.TransactionID, &record.Type,
		&record.Amount, &record.BalanceBefore, &record.BalanceAfter, &record.Notes, &record.CreatedBy, &record.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create kasbon record: %w", err)
	}

	// Update customer debt
	_, err = tx.ExecContext(ctx, "UPDATE customers SET current_debt = $1, updated_at = NOW() WHERE id = $2", newBalance, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to update customer debt: %w", err)
	}

	return &record, nil
}

// CreatePayment creates a new payment record
func (r *KasbonRepository) CreatePayment(ctx context.Context, input domain.KasbonPaymentInput) (*domain.KasbonRecord, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var currentDebt int64
	err = tx.QueryRowContext(ctx, "SELECT current_debt FROM customers WHERE id = $1", input.CustomerID).Scan(&currentDebt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if input.Amount > currentDebt {
		input.Amount = currentDebt // Can't pay more than owed
	}

	newBalance := currentDebt - input.Amount

	query := `
		INSERT INTO kasbon_records (customer_id, type, amount, balance_before, balance_after, notes, created_by)
		VALUES ($1, 'payment', $2, $3, $4, $5, $6)
		RETURNING id, customer_id, transaction_id, type, amount, balance_before, balance_after, notes, created_by, created_at
	`

	var record domain.KasbonRecord
	err = tx.QueryRowContext(ctx, query, input.CustomerID, input.Amount, currentDebt, newBalance, input.Notes, input.CreatedBy).Scan(
		&record.ID, &record.CustomerID, &record.TransactionID, &record.Type,
		&record.Amount, &record.BalanceBefore, &record.BalanceAfter, &record.Notes, &record.CreatedBy, &record.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, "UPDATE customers SET current_debt = $1, updated_at = NOW() WHERE id = $2", newBalance, input.CustomerID)
	if err != nil {
		return nil, err
	}

	return &record, tx.Commit()
}

// GetByCustomer retrieves kasbon records for a customer
func (r *KasbonRepository) GetByCustomer(ctx context.Context, customerID uuid.UUID, filter domain.KasbonFilter) ([]domain.KasbonRecord, int64, error) {
	args := []interface{}{customerID}
	whereClause := "WHERE customer_id = $1"
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM kasbon_records %s", whereClause)
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
		SELECT id, customer_id, transaction_id, type, amount, balance_before, balance_after, notes, created_by, created_at
		FROM kasbon_records %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, perPage, (page-1)*perPage)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var records []domain.KasbonRecord
	for rows.Next() {
		var rec domain.KasbonRecord
		if err := rows.Scan(
			&rec.ID, &rec.CustomerID, &rec.TransactionID, &rec.Type,
			&rec.Amount, &rec.BalanceBefore, &rec.BalanceAfter, &rec.Notes, &rec.CreatedBy, &rec.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		records = append(records, rec)
	}
	return records, total, rows.Err()
}

// GetSummary returns kasbon summary for a customer
func (r *KasbonRepository) GetSummary(ctx context.Context, customerID uuid.UUID) (*domain.KasbonSummary, error) {
	query := `
		SELECT c.id, c.name, c.current_debt, c.credit_limit,
			COALESCE((SELECT SUM(amount) FROM kasbon_records WHERE customer_id = c.id AND type = 'debt'), 0),
			COALESCE((SELECT SUM(amount) FROM kasbon_records WHERE customer_id = c.id AND type = 'payment'), 0),
			(SELECT MAX(created_at) FROM kasbon_records WHERE customer_id = c.id)
		FROM customers c WHERE c.id = $1
	`

	var summary domain.KasbonSummary
	err := r.db.QueryRowContext(ctx, query, customerID).Scan(
		&summary.CustomerID, &summary.CustomerName, &summary.CurrentBalance, &summary.CreditLimit,
		&summary.TotalDebt, &summary.TotalPayment, &summary.LastTransactionAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if summary.CreditLimit > 0 {
		summary.RemainingCredit = summary.CreditLimit - summary.CurrentBalance
		if summary.RemainingCredit < 0 {
			summary.RemainingCredit = 0
		}
	} else {
		summary.RemainingCredit = -1 // unlimited
	}

	return &summary, nil
}

// GetReport returns kasbon report for all customers
func (r *KasbonRepository) GetReport(ctx context.Context) (*domain.KasbonReport, error) {
	query := `
		SELECT COUNT(*), COALESCE(SUM(current_debt), 0), COUNT(*) FILTER (WHERE current_debt > 0)
		FROM customers WHERE is_active = true
	`

	var report domain.KasbonReport
	err := r.db.QueryRowContext(ctx, query).Scan(&report.TotalCustomers, &report.TotalOutstanding, &report.CustomersWithDebt)
	if err != nil {
		return nil, err
	}

	// Get top debtors
	summaryQuery := `
		SELECT id, name, current_debt, credit_limit FROM customers
		WHERE is_active = true AND current_debt > 0
		ORDER BY current_debt DESC LIMIT 10
	`
	rows, err := r.db.QueryContext(ctx, summaryQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s domain.KasbonSummary
		if err := rows.Scan(&s.CustomerID, &s.CustomerName, &s.CurrentBalance, &s.CreditLimit); err != nil {
			return nil, err
		}
		report.Summaries = append(report.Summaries, s)
	}

	return &report, rows.Err()
}

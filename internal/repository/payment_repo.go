package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
)

// PaymentRepository handles payment data access
type PaymentRepository struct {
	db *database.PostgresDB
}

// NewPaymentRepository creates a new PaymentRepository
func NewPaymentRepository(db *database.PostgresDB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// Create creates a new payment record
func (r *PaymentRepository) Create(ctx context.Context, tx *sql.Tx, record *domain.PaymentRecord) error {
	query := `
		INSERT INTO payment_records (
			transaction_id, order_id, snap_token, redirect_url,
			payment_type, gross_amount, currency, status, fraud_status,
			midtrans_response, paid_at, expired_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`

	var execer interface {
		QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	}
	if tx != nil {
		execer = tx
	} else {
		execer = r.db
	}

	respStr := "{}"
	if len(record.MidtransResp) > 0 {
		respStr = string(record.MidtransResp)
	}

	return execer.QueryRowContext(ctx, query,
		record.TransactionID,
		record.OrderID,
		record.SnapToken,
		record.RedirectURL,
		record.PaymentType,
		record.GrossAmount,
		record.Currency,
		record.Status,
		record.FraudStatus,
		respStr,
		record.PaidAt,
		record.ExpiredAt,
	).Scan(&record.ID, &record.CreatedAt, &record.UpdatedAt)
}

// GetByID retrieves a payment record by ID
func (r *PaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PaymentRecord, error) {
	query := `
		SELECT id, transaction_id, order_id, snap_token, redirect_url,
			payment_type, gross_amount, currency, status, fraud_status,
			midtrans_response, paid_at, expired_at, created_at, updated_at
		FROM payment_records
		WHERE id = $1
	`

	record := &domain.PaymentRecord{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&record.ID,
		&record.TransactionID,
		&record.OrderID,
		&record.SnapToken,
		&record.RedirectURL,
		&record.PaymentType,
		&record.GrossAmount,
		&record.Currency,
		&record.Status,
		&record.FraudStatus,
		&record.MidtransResp,
		&record.PaidAt,
		&record.ExpiredAt,
		&record.CreatedAt,
		&record.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return record, nil
}

// GetByOrderID retrieves a payment record by Midtrans order ID
func (r *PaymentRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.PaymentRecord, error) {
	query := `
		SELECT id, transaction_id, order_id, snap_token, redirect_url,
			payment_type, gross_amount, currency, status, fraud_status,
			midtrans_response, paid_at, expired_at, created_at, updated_at
		FROM payment_records
		WHERE order_id = $1
	`

	record := &domain.PaymentRecord{}
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&record.ID,
		&record.TransactionID,
		&record.OrderID,
		&record.SnapToken,
		&record.RedirectURL,
		&record.PaymentType,
		&record.GrossAmount,
		&record.Currency,
		&record.Status,
		&record.FraudStatus,
		&record.MidtransResp,
		&record.PaidAt,
		&record.ExpiredAt,
		&record.CreatedAt,
		&record.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return record, nil
}

// GetByTransactionID retrieves a payment record by transaction ID
func (r *PaymentRepository) GetByTransactionID(ctx context.Context, transactionID uuid.UUID) (*domain.PaymentRecord, error) {
	query := `
		SELECT id, transaction_id, order_id, snap_token, redirect_url,
			payment_type, gross_amount, currency, status, fraud_status,
			midtrans_response, paid_at, expired_at, created_at, updated_at
		FROM payment_records
		WHERE transaction_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	record := &domain.PaymentRecord{}
	err := r.db.QueryRowContext(ctx, query, transactionID).Scan(
		&record.ID,
		&record.TransactionID,
		&record.OrderID,
		&record.SnapToken,
		&record.RedirectURL,
		&record.PaymentType,
		&record.GrossAmount,
		&record.Currency,
		&record.Status,
		&record.FraudStatus,
		&record.MidtransResp,
		&record.PaidAt,
		&record.ExpiredAt,
		&record.CreatedAt,
		&record.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return record, nil
}

// UpdateStatus updates the payment status and stores the Midtrans response
func (r *PaymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PaymentStatus, response json.RawMessage) error {
	var paidAt *time.Time
	if status.IsSuccess() {
		now := time.Now()
		paidAt = &now
	}

	query := `
		UPDATE payment_records
		SET status = $2, midtrans_response = $3, paid_at = COALESCE($4, paid_at), updated_at = NOW()
		WHERE id = $1
	`

	// Cast response (json.RawMessage / []byte) to string for JSONB driver compatibility
	result, err := r.db.ExecContext(ctx, query, id, status, string(response), paidAt)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// UpdateSnapToken updates the snap token and redirect URL
func (r *PaymentRepository) UpdateSnapToken(ctx context.Context, id uuid.UUID, snapToken, redirectURL string) error {
	query := `
		UPDATE payment_records
		SET snap_token = $2, redirect_url = $3, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, snapToken, redirectURL)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

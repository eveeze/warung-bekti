package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
)

type ConsignmentRepository struct {
	db *database.PostgresDB
}

func NewConsignmentRepository(db *database.PostgresDB) *ConsignmentRepository {
	return &ConsignmentRepository{db: db}
}

func (r *ConsignmentRepository) CreateConsignor(ctx context.Context, input domain.CreateConsignorInput) (*domain.Consignor, error) {
	query := `
		INSERT INTO consignors (name, phone, address, bank_account, bank_name, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	c := &domain.Consignor{
		Name:        input.Name,
		Phone:       input.Phone,
		Address:     input.Address,
		BankAccount: input.BankAccount,
		BankName:    input.BankName,
		Notes:       input.Notes,
		IsActive:    true,
	}
	err := r.db.QueryRowContext(ctx, query,
		c.Name, c.Phone, c.Address, c.BankAccount, c.BankName, c.Notes,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *ConsignmentRepository) UpdateConsignor(ctx context.Context, id uuid.UUID, input domain.UpdateConsignorInput) (*domain.Consignor, error) {
	// Build dynamic query
	// Simplified for now: just fetch then update? Or dynamic SQL building.
	// Since minimal fields, let's use COALESCE in SQL or standard update.
	
	query := `
		UPDATE consignors
		SET name = COALESCE($2, name), phone = COALESCE($3, phone), address = COALESCE($4, address),
		    bank_account = COALESCE($5, bank_account), bank_name = COALESCE($6, bank_name),
		    notes = COALESCE($7, notes), is_active = COALESCE($8, is_active), updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, phone, address, bank_account, bank_name, notes, is_active, created_at, updated_at
	`
	var c domain.Consignor
	err := r.db.QueryRowContext(ctx, query,
		id, input.Name, input.Phone, input.Address, input.BankAccount, input.BankName, input.Notes, input.IsActive,
	).Scan(
		&c.ID, &c.Name, &c.Phone, &c.Address, &c.BankAccount, &c.BankName, &c.Notes, &c.IsActive, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ConsignmentRepository) ListConsignors(ctx context.Context) ([]domain.Consignor, error) {
	query := `SELECT id, name, phone, address, bank_account, bank_name, notes, is_active, created_at, updated_at FROM consignors ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var consignors []domain.Consignor
	for rows.Next() {
		var c domain.Consignor
		if err := rows.Scan(
			&c.ID, &c.Name, &c.Phone, &c.Address, &c.BankAccount, &c.BankName, &c.Notes, &c.IsActive, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		consignors = append(consignors, c)
	}
	return consignors, nil
}

// Settlements
func (r *ConsignmentRepository) CreateSettlement(ctx context.Context, s *domain.ConsignmentSettlement) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO consignment_settlements (settlement_number, consignor_id, period_start, period_end, total_sales, commission_amount, consignor_amount, status, created_by)
		VALUES (generate_settlement_number(), $1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, settlement_number, created_at, updated_at
	`
	err = tx.QueryRowContext(ctx, query,
		s.ConsignorID, s.PeriodStart, s.PeriodEnd, s.TotalSales, s.CommissionAmount, s.ConsignorAmount, s.Status, s.CreatedBy,
	).Scan(&s.ID, &s.SettlementNumber, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return err
	}

	itemQuery := `
		INSERT INTO consignment_settlement_items (settlement_id, product_id, product_name, quantity_sold, unit_price, total_sales, commission_rate, commission_amount, consignor_amount)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	for _, item := range s.Items {
		_, err = tx.ExecContext(ctx, itemQuery,
			s.ID, item.ProductID, item.ProductName, item.QuantitySold, item.UnitPrice,
			item.TotalSales, item.CommissionRate, item.CommissionAmount, item.ConsignorAmount,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

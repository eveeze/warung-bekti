package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
)

// CashFlowRepository handles cash flow database operations
type CashFlowRepository struct {
	db *database.PostgresDB
}

// NewCashFlowRepository creates a new CashFlowRepository
func NewCashFlowRepository(db *database.PostgresDB) *CashFlowRepository {
	return &CashFlowRepository{db: db}
}

// -- Categories --

func (r *CashFlowRepository) GetCategories(ctx context.Context) ([]domain.CashFlowCategory, error) {
	query := `SELECT id, name, type, description, is_active, created_at, updated_at FROM cash_flow_categories WHERE is_active = true ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.CashFlowCategory
	for rows.Next() {
		var c domain.CashFlowCategory
		if err := rows.Scan(&c.ID, &c.Name, &c.Type, &c.Description, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

// -- Drawer Sessions --

func (r *CashFlowRepository) OpenDrawer(ctx context.Context, input domain.OpenDrawerInput) (*domain.CashDrawerSession, error) {
	// Check if there is already an open session
	var count int
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM cash_drawer_sessions WHERE status = 'open'").Scan(&count)
	if count > 0 {
		return nil, fmt.Errorf("there is already an open drawer session")
	}

	session := &domain.CashDrawerSession{
		SessionDate:    time.Now().Truncate(24 * time.Hour),
		OpeningBalance: input.OpeningBalance,
		Status:         domain.DrawerSessionStatusOpen,
		OpenedBy:       &input.OpenedBy,
		Notes:          input.Notes,
		OpenedAt:       time.Now(),
	}

	query := `
		INSERT INTO cash_drawer_sessions (session_date, opening_balance, status, opened_by, notes, opened_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query, session.SessionDate, session.OpeningBalance, session.Status, session.OpenedBy, session.Notes, session.OpenedAt).Scan(&session.ID, &session.CreatedAt, &session.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (r *CashFlowRepository) CloseDrawer(ctx context.Context, input domain.CloseDrawerInput) (*domain.CashDrawerSession, error) {
	// Calculate expected closing balance
	// expected = opening + income - expense
	var income, expense int64
	summaryQuery := `
		SELECT 
			COALESCE(SUM(amount) FILTER (WHERE type = 'income'), 0),
			COALESCE(SUM(amount) FILTER (WHERE type = 'expense'), 0)
		FROM cash_flow_records WHERE drawer_session_id = $1
	`
	r.db.QueryRowContext(ctx, summaryQuery, input.SessionID).Scan(&income, &expense)

	var openingBalance int64
	err := r.db.QueryRowContext(ctx, "SELECT opening_balance FROM cash_drawer_sessions WHERE id = $1", input.SessionID).Scan(&openingBalance)
	if err != nil {
		return nil, err
	}

	expected := openingBalance + income - expense
	difference := input.ClosingBalance - expected
	now := time.Now()

	query := `
		UPDATE cash_drawer_sessions
		SET closing_balance = $2, expected_closing = $3, difference = $4, status = 'closed', closed_by = $5, notes = $6, closed_at = $7, updated_at = NOW()
		WHERE id = $1
		RETURNING id, session_date, opening_balance, status, opened_by,  opened_at, created_at, updated_at
	`
	
	session := &domain.CashDrawerSession{
		ID:              input.SessionID,
		ClosingBalance:  &input.ClosingBalance,
		ExpectedClosing: &expected,
		Difference:      &difference,
		ClosedBy:        &input.ClosedBy,
		Notes:           input.Notes,
		ClosedAt:        &now,
		Status:          domain.DrawerSessionStatusClosed,
	}

	err = r.db.QueryRowContext(ctx, query, input.SessionID, input.ClosingBalance, expected, difference, input.ClosedBy, input.Notes, now).Scan(
		&session.ID, &session.SessionDate, &session.OpeningBalance, &session.Status, &session.OpenedBy, &session.OpenedAt, &session.CreatedAt, &session.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	session.TotalIncome = income
	session.TotalExpense = expense
	
	return session, nil
}

func (r *CashFlowRepository) GetCurrentSession(ctx context.Context) (*domain.CashDrawerSession, error) {
	query := `
		SELECT id, session_date, opening_balance, status, opened_by, notes, opened_at, created_at, updated_at
		FROM cash_drawer_sessions WHERE status = 'open' LIMIT 1
	`
	var session domain.CashDrawerSession
	err := r.db.QueryRowContext(ctx, query).Scan(
		&session.ID, &session.SessionDate, &session.OpeningBalance, &session.Status, &session.OpenedBy, &session.Notes, &session.OpenedAt, &session.CreatedAt, &session.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// -- Cash Flow Records --

func (r *CashFlowRepository) RecordCashFlow(ctx context.Context, tx *sql.Tx, input domain.CashFlowInput, sessionID *uuid.UUID, refType *string, refID *uuid.UUID) (*domain.CashFlowRecord, error) {
	record := &domain.CashFlowRecord{
		DrawerSessionID: sessionID,
		CategoryID:      input.CategoryID,
		Type:            input.Type,
		Amount:          input.Amount,
		Description:     input.Description,
		ReferenceType:   refType,
		ReferenceID:     refID,
		CreatedBy:       &input.CreatedBy,
	}

	query := `
		INSERT INTO cash_flow_records (drawer_session_id, category_id, type, amount, description, reference_type, reference_id, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`

	var err error
	if tx != nil {
		err = tx.QueryRowContext(ctx, query, sessionID, input.CategoryID, input.Type, input.Amount, input.Description, refType, refID, input.CreatedBy).Scan(&record.ID, &record.CreatedAt)
	} else {
		err = r.db.QueryRowContext(ctx, query, sessionID, input.CategoryID, input.Type, input.Amount, input.Description, refType, refID, input.CreatedBy).Scan(&record.ID, &record.CreatedAt)
	}

	if err != nil {
		return nil, err
	}
	return record, nil
}

func (r *CashFlowRepository) ListCashFlows(ctx context.Context, filter domain.CashFlowFilter) ([]domain.CashFlowRecord, int64, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if filter.SessionID != nil {
		whereClause += fmt.Sprintf(" AND drawer_session_id = $%d", argIndex)
		args = append(args, *filter.SessionID)
		argIndex++
	}
	if filter.CategoryID != nil {
		whereClause += fmt.Sprintf(" AND category_id = $%d", argIndex)
		args = append(args, *filter.CategoryID)
		argIndex++
	}
	if filter.Type != nil {
		whereClause += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, *filter.Type)
		argIndex++
	}

	var total int64
	r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM cash_flow_records %s", whereClause), args...).Scan(&total)

	query := fmt.Sprintf(`
		SELECT cfp.id, cfp.drawer_session_id, cfp.category_id, cfp.type, cfp.amount, cfp.description, cfp.reference_type, cfp.reference_id, cfp.created_by, cfp.created_at,
		       c.name
		FROM cash_flow_records cfp
		LEFT JOIN cash_flow_categories c ON cfp.category_id = c.id
		%s
		ORDER BY cfp.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
	
	page := filter.Page
	if page < 1 { page = 1 }
	perPage := filter.PerPage
	if perPage < 1 { perPage = 20 }

	args = append(args, perPage, (page-1)*perPage)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var records []domain.CashFlowRecord
	for rows.Next() {
		var rec domain.CashFlowRecord
		var catName *string
		if err := rows.Scan(
			&rec.ID, &rec.DrawerSessionID, &rec.CategoryID, &rec.Type, &rec.Amount, &rec.Description, &rec.ReferenceType, &rec.ReferenceID, &rec.CreatedBy, &rec.CreatedAt,
			&catName,
		); err != nil {
			return nil, 0, err
		}
		if rec.CategoryID != nil && catName != nil {
			rec.Category = &domain.CashFlowCategory{ID: *rec.CategoryID, Name: *catName}
		}
		records = append(records, rec)
	}

	return records, total, nil
}

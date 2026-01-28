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

// CustomerRepository handles customer database operations
type CustomerRepository struct {
	db *database.PostgresDB
}

// NewCustomerRepository creates a new CustomerRepository
func NewCustomerRepository(db *database.PostgresDB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

// Create creates a new customer
func (r *CustomerRepository) Create(ctx context.Context, input domain.CustomerCreateInput) (*domain.Customer, error) {
	creditLimit := int64(0)
	if input.CreditLimit != nil {
		creditLimit = *input.CreditLimit
	}

	query := `
		INSERT INTO customers (name, phone, address, notes, credit_limit)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, phone, address, notes, credit_limit, current_debt, is_active, created_at, updated_at
	`

	var customer domain.Customer
	err := r.db.QueryRowContext(ctx, query,
		input.Name, input.Phone, input.Address, input.Notes, creditLimit,
	).Scan(
		&customer.ID, &customer.Name, &customer.Phone, &customer.Address,
		&customer.Notes, &customer.CreditLimit, &customer.CurrentDebt,
		&customer.IsActive, &customer.CreatedAt, &customer.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return &customer, nil
}

// GetByID retrieves a customer by ID
func (r *CustomerRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	query := `
		SELECT id, name, phone, address, notes, credit_limit, current_debt, is_active, created_at, updated_at
		FROM customers
		WHERE id = $1
	`

	var customer domain.Customer
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&customer.ID, &customer.Name, &customer.Phone, &customer.Address,
		&customer.Notes, &customer.CreditLimit, &customer.CurrentDebt,
		&customer.IsActive, &customer.CreatedAt, &customer.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return &customer, nil
}

// List retrieves customers with filtering and pagination
func (r *CustomerRepository) List(ctx context.Context, filter domain.CustomerFilter) ([]domain.Customer, int64, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(name ILIKE $%d OR phone ILIKE $%d)",
			argIndex, argIndex,
		))
		args = append(args, "%"+*filter.Search+"%")
		argIndex++
	}

	if filter.HasDebt != nil && *filter.HasDebt {
		conditions = append(conditions, "current_debt > 0")
	}

	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *filter.IsActive)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM customers %s", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count customers: %w", err)
	}

	// Sort & pagination
	validSortFields := map[string]bool{
		"name": true, "created_at": true, "current_debt": true,
	}
	sortBy := "name"
	if filter.SortBy != "" && validSortFields[filter.SortBy] {
		sortBy = filter.SortBy
	}
	sortOrder := "ASC"
	if strings.ToUpper(filter.SortOrder) == "DESC" {
		sortOrder = "DESC"
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	query := fmt.Sprintf(`
		SELECT id, name, phone, address, notes, credit_limit, current_debt, is_active, created_at, updated_at
		FROM customers
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortBy, sortOrder, argIndex, argIndex+1)

	args = append(args, perPage, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list customers: %w", err)
	}
	defer rows.Close()

	var customers []domain.Customer
	for rows.Next() {
		var c domain.Customer
		if err := rows.Scan(
			&c.ID, &c.Name, &c.Phone, &c.Address, &c.Notes,
			&c.CreditLimit, &c.CurrentDebt, &c.IsActive, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, c)
	}

	return customers, total, rows.Err()
}

// Update updates a customer
func (r *CustomerRepository) Update(ctx context.Context, id uuid.UUID, input domain.CustomerUpdateInput) (*domain.Customer, error) {
	var setClauses []string
	var args []interface{}
	argIndex := 1

	if input.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *input.Name)
		argIndex++
	}
	if input.Phone != nil {
		setClauses = append(setClauses, fmt.Sprintf("phone = $%d", argIndex))
		args = append(args, input.Phone)
		argIndex++
	}
	if input.Address != nil {
		setClauses = append(setClauses, fmt.Sprintf("address = $%d", argIndex))
		args = append(args, input.Address)
		argIndex++
	}
	if input.Notes != nil {
		setClauses = append(setClauses, fmt.Sprintf("notes = $%d", argIndex))
		args = append(args, input.Notes)
		argIndex++
	}
	if input.CreditLimit != nil {
		setClauses = append(setClauses, fmt.Sprintf("credit_limit = $%d", argIndex))
		args = append(args, *input.CreditLimit)
		argIndex++
	}
	if input.IsActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *input.IsActive)
		argIndex++
	}

	if len(setClauses) == 0 {
		return r.GetByID(ctx, id)
	}

	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE customers SET %s, updated_at = NOW()
		WHERE id = $%d
	`, strings.Join(setClauses, ", "), argIndex)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetByID(ctx, id)
}

// UpdateDebt updates the customer's current debt
func (r *CustomerRepository) UpdateDebt(ctx context.Context, id uuid.UUID, newDebt int64) error {
	query := `UPDATE customers SET current_debt = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, newDebt, id)
	if err != nil {
		return fmt.Errorf("failed to update debt: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// AddDebt adds to the customer's current debt
func (r *CustomerRepository) AddDebt(ctx context.Context, id uuid.UUID, amount int64) error {
	query := `UPDATE customers SET current_debt = current_debt + $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, amount, id)
	if err != nil {
		return fmt.Errorf("failed to add debt: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// SubtractDebt subtracts from the customer's current debt
func (r *CustomerRepository) SubtractDebt(ctx context.Context, id uuid.UUID, amount int64) error {
	query := `
		UPDATE customers 
		SET current_debt = GREATEST(0, current_debt - $1), updated_at = NOW() 
		WHERE id = $2
	`
	result, err := r.db.ExecContext(ctx, query, amount, id)
	if err != nil {
		return fmt.Errorf("failed to subtract debt: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete soft deletes a customer
func (r *CustomerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE customers SET is_active = false, updated_at = NOW() WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// GetCustomersWithDebt retrieves all customers with outstanding debt
func (r *CustomerRepository) GetCustomersWithDebt(ctx context.Context) ([]domain.Customer, error) {
	query := `
		SELECT id, name, phone, address, notes, credit_limit, current_debt, is_active, created_at, updated_at
		FROM customers
		WHERE current_debt > 0 AND is_active = true
		ORDER BY current_debt DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get customers with debt: %w", err)
	}
	defer rows.Close()

	var customers []domain.Customer
	for rows.Next() {
		var c domain.Customer
		if err := rows.Scan(
			&c.ID, &c.Name, &c.Phone, &c.Address, &c.Notes,
			&c.CreditLimit, &c.CurrentDebt, &c.IsActive, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, c)
	}

	return customers, rows.Err()
}

// GetTotalOutstandingDebt returns the total outstanding debt across all customers
func (r *CustomerRepository) GetTotalOutstandingDebt(ctx context.Context) (int64, error) {
	query := `SELECT COALESCE(SUM(current_debt), 0) FROM customers WHERE is_active = true`
	var total int64
	err := r.db.QueryRowContext(ctx, query).Scan(&total)
	return total, err
}

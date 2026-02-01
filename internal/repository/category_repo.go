package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/google/uuid"
)

type CategoryRepository struct {
	db *database.PostgresDB
}

func NewCategoryRepository(db *database.PostgresDB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// Create inserts a new category
func (r *CategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	query := `
		INSERT INTO categories (id, name, description, parent_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	category.ID = uuid.New()
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()
	category.IsActive = true

	_, err := r.db.ExecContext(ctx, query,
		category.ID,
		category.Name,
		category.Description,
		category.ParentID,
		category.IsActive,
		category.CreatedAt,
		category.UpdatedAt,
	)
	return err
}

// FindAll retrieves categories with product count
func (r *CategoryRepository) FindAll(ctx context.Context, filter map[string]interface{}) ([]domain.CategoryResponse, error) {
	// Critical: Join with products to get count
	query := `
		SELECT 
			c.id, c.name, c.description, c.parent_id, c.is_active, c.created_at, c.updated_at,
			COUNT(p.id) as product_count
		FROM categories c
		LEFT JOIN products p ON c.id = p.category_id AND p.is_active = true
		WHERE c.is_active = true
	`
	
	// Add filters if needed (e.g. search)
	args := []interface{}{}
	if search, ok := filter["search"].(string); ok && search != "" {
		query += " AND c.name ILIKE $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, "%"+search+"%")
	}

	query += " GROUP BY c.id ORDER BY c.name ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.CategoryResponse
	for rows.Next() {
		var c domain.CategoryResponse
		if err := rows.Scan(
			&c.ID, &c.Name, &c.Description, &c.ParentID, &c.IsActive, &c.CreatedAt, &c.UpdatedAt,
			&c.ProductCount,
		); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

// FindByID retrieves a single category
func (r *CategoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	query := `
		SELECT id, name, description, parent_id, is_active, created_at, updated_at
		FROM categories
		WHERE id = $1 AND is_active = true
	`
	var c domain.Category
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.Name, &c.Description, &c.ParentID, &c.IsActive, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Update modifies a category
func (r *CategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	query := `
		UPDATE categories
		SET name = $1, description = $2, parent_id = $3, is_active = $4, updated_at = $5
		WHERE id = $6
	`
	category.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query,
		category.Name,
		category.Description,
		category.ParentID,
		category.IsActive,
		category.UpdatedAt,
		category.ID,
	)
	return err
}

// Delete soft deletes a category
// Strictly, user requested validation against existing products.
// That validation check should happen at Service layer or via FK constraint check.
// Here we just perform delete/soft-delete.
func (r *CategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE categories SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// HasProducts checks if a category is used by any products
func (r *CategoryRepository) HasProducts(ctx context.Context, categoryID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE category_id = $1 AND is_active = true)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, categoryID).Scan(&exists)
	return exists, err
}

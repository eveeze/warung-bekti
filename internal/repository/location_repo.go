package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
)

// LocationRepository handles location database operations
type LocationRepository struct {
	db *database.PostgresDB
}

// NewLocationRepository creates a new LocationRepository
func NewLocationRepository(db *database.PostgresDB) *LocationRepository {
	return &LocationRepository{db: db}
}

// Create creates a new location
func (r *LocationRepository) Create(ctx context.Context, input domain.LocationCreateInput) (*domain.Location, error) {
	query := `
		INSERT INTO locations (
			name, category, x_coordinate, y_coordinate, z_coordinate,
			width, depth, height, description
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING 
			id, name, category, x_coordinate, y_coordinate, z_coordinate,
			width, depth, height, description, is_active, created_at, updated_at
	`

	var loc domain.Location
	err := r.db.QueryRowContext(ctx, query,
		input.Name,
		input.Category,
		input.XCoordinate,
		input.YCoordinate,
		input.ZCoordinate,
		input.Width,
		input.Depth,
		input.Height,
		input.Description,
	).Scan(
		&loc.ID, &loc.Name, &loc.Category,
		&loc.XCoordinate, &loc.YCoordinate, &loc.ZCoordinate,
		&loc.Width, &loc.Depth, &loc.Height,
		&loc.Description, &loc.IsActive, &loc.CreatedAt, &loc.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create location: %w", err)
	}

	return &loc, nil
}

// GetByID retrieves a location by ID
func (r *LocationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Location, error) {
	query := `
		SELECT 
			id, name, category, x_coordinate, y_coordinate, z_coordinate,
			width, depth, height, description, is_active, created_at, updated_at
		FROM locations
		WHERE id = $1
	`

	var loc domain.Location
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&loc.ID, &loc.Name, &loc.Category,
		&loc.XCoordinate, &loc.YCoordinate, &loc.ZCoordinate,
		&loc.Width, &loc.Depth, &loc.Height,
		&loc.Description, &loc.IsActive, &loc.CreatedAt, &loc.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get location by ID: %w", err)
	}

	return &loc, nil
}

// ListAll retrieves all locations (typically for rendering map)
func (r *LocationRepository) ListAll(ctx context.Context) ([]domain.Location, error) {
	query := `
		SELECT 
			id, name, category, x_coordinate, y_coordinate, z_coordinate,
			width, depth, height, description, is_active, created_at, updated_at
		FROM locations
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list locations: %w", err)
	}
	defer rows.Close()

	var locations []domain.Location
	for rows.Next() {
		var loc domain.Location
		err := rows.Scan(
			&loc.ID, &loc.Name, &loc.Category,
			&loc.XCoordinate, &loc.YCoordinate, &loc.ZCoordinate,
			&loc.Width, &loc.Depth, &loc.Height,
			&loc.Description, &loc.IsActive, &loc.CreatedAt, &loc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan location: %w", err)
		}
		locations = append(locations, loc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over locations: %w", err)
	}

	return locations, nil
}

// Update updates a location
func (r *LocationRepository) Update(ctx context.Context, id uuid.UUID, input domain.LocationUpdateInput) (*domain.Location, error) {
	var setClauses []string
	var args []interface{}
	argCount := 1

	if input.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argCount))
		args = append(args, *input.Name)
		argCount++
	}

	if input.Category != nil {
		setClauses = append(setClauses, fmt.Sprintf("category = $%d", argCount))
		args = append(args, *input.Category)
		argCount++
	}

	if input.XCoordinate != nil {
		setClauses = append(setClauses, fmt.Sprintf("x_coordinate = $%d", argCount))
		args = append(args, *input.XCoordinate)
		argCount++
	}

	if input.YCoordinate != nil {
		setClauses = append(setClauses, fmt.Sprintf("y_coordinate = $%d", argCount))
		args = append(args, *input.YCoordinate)
		argCount++
	}
	
	if input.ZCoordinate != nil {
		setClauses = append(setClauses, fmt.Sprintf("z_coordinate = $%d", argCount))
		args = append(args, *input.ZCoordinate)
		argCount++
	}

	if input.Width != nil {
		setClauses = append(setClauses, fmt.Sprintf("width = $%d", argCount))
		args = append(args, *input.Width)
		argCount++
	}
	
	if input.Depth != nil {
		setClauses = append(setClauses, fmt.Sprintf("depth = $%d", argCount))
		args = append(args, *input.Depth)
		argCount++
	}
	
	if input.Height != nil {
		setClauses = append(setClauses, fmt.Sprintf("height = $%d", argCount))
		args = append(args, *input.Height)
		argCount++
	}

	if input.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argCount))
		args = append(args, *input.Description)
		argCount++
	}

	if input.IsActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argCount))
		args = append(args, *input.IsActive)
		argCount++
	}

	if len(setClauses) == 0 {
		return r.GetByID(ctx, id)
	}

	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE locations
		SET %s
		WHERE id = $%d
		RETURNING 
			id, name, category, x_coordinate, y_coordinate, z_coordinate,
			width, depth, height, description, is_active, created_at, updated_at
	`, strings.Join(setClauses, ", "), argCount)

	var loc domain.Location
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&loc.ID, &loc.Name, &loc.Category,
		&loc.XCoordinate, &loc.YCoordinate, &loc.ZCoordinate,
		&loc.Width, &loc.Depth, &loc.Height,
		&loc.Description, &loc.IsActive, &loc.CreatedAt, &loc.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to update location: %w", err)
	}

	return &loc, nil
}

// Delete deletes a location
func (r *LocationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM locations WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete location: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

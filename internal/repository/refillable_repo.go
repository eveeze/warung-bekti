package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
)

type RefillableRepository struct {
	db *database.PostgresDB
}

func NewRefillableRepository(db *database.PostgresDB) *RefillableRepository {
	return &RefillableRepository{db: db}
}

func (r *RefillableRepository) GetContainers(ctx context.Context) ([]domain.RefillableContainer, error) {
	query := `
		SELECT rc.id, rc.product_id, rc.container_type, rc.empty_count, rc.full_count, rc.notes, rc.created_at, rc.updated_at,
		       p.name, p.barcode
		FROM refillable_containers rc
		JOIN products p ON rc.product_id = p.id
		ORDER BY rc.container_type
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var containers []domain.RefillableContainer
	for rows.Next() {
		var c domain.RefillableContainer
		var pName string
		var pBarcode *string
		if err := rows.Scan(
			&c.ID, &c.ProductID, &c.ContainerType, &c.EmptyCount, &c.FullCount, &c.Notes, &c.CreatedAt, &c.UpdatedAt,
			&pName, &pBarcode,
		); err != nil {
			return nil, err
		}
		c.Product = &domain.Product{ID: c.ProductID, Name: pName, Barcode: pBarcode}
		containers = append(containers, c)
	}
	return containers, nil
	return containers, nil
}

func (r *RefillableRepository) Create(ctx context.Context, container *domain.RefillableContainer) error {
	query := `
		INSERT INTO refillable_containers (product_id, container_type, empty_count, full_count, notes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		container.ProductID, container.ContainerType, container.EmptyCount, container.FullCount, container.Notes,
	).Scan(&container.ID, &container.CreatedAt, &container.UpdatedAt)
}

func (r *RefillableRepository) UpdateContainerStock(ctx context.Context, tx *sql.Tx, containerID uuid.UUID, emptyChange, fullChange int) error {
	var execer interface {
		ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	}
	if tx != nil {
		execer = tx
	} else {
		execer = r.db
	}

	query := `
		UPDATE refillable_containers
		SET empty_count = empty_count + $2, full_count = full_count + $3, updated_at = NOW()
		WHERE id = $1
	`
	_, err := execer.ExecContext(ctx, query, containerID, emptyChange, fullChange)
	return err
}

func (r *RefillableRepository) RecordMovement(ctx context.Context, tx *sql.Tx, m *domain.ContainerMovement) error {
	var execer interface {
		QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	}
	if tx != nil {
		execer = tx
	} else {
		execer = r.db
	}

	// Get current balance/snapshot for history
	var empty, full int
	err := execer.QueryRowContext(ctx, "SELECT empty_count, full_count FROM refillable_containers WHERE id = $1", m.ContainerID).Scan(&empty, &full)
	if err != nil {
		return err
	}

	m.EmptyBefore = empty
	m.FullBefore = full
	m.EmptyAfter = empty + m.EmptyChange
	m.FullAfter = full + m.FullChange

	query := `
		INSERT INTO container_movements (
			container_id, type, empty_change, full_change, empty_before, empty_after, full_before, full_after,
			reference_type, reference_id, notes, created_by, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
		RETURNING id, created_at
	`
	return execer.QueryRowContext(ctx, query,
		m.ContainerID, m.Type, m.EmptyChange, m.FullChange, m.EmptyBefore, m.EmptyAfter, m.FullBefore, m.FullAfter,
		m.ReferenceType, m.ReferenceID, m.Notes, m.CreatedBy,
	).Scan(&m.ID, &m.CreatedAt)
}

func (r *RefillableRepository) GetMovements(ctx context.Context, containerID uuid.UUID, limit, offset int) ([]domain.ContainerMovement, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM container_movements WHERE container_id = $1`
	if err := r.db.QueryRowContext(ctx, countQuery, containerID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, container_id, type, empty_change, full_change, empty_before, empty_after,
		       full_before, full_after, reference_type, reference_id, notes, created_by, created_at
		FROM container_movements
		WHERE container_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, containerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var movements []domain.ContainerMovement
	for rows.Next() {
		var m domain.ContainerMovement
		if err := rows.Scan(
			&m.ID, &m.ContainerID, &m.Type, &m.EmptyChange, &m.FullChange,
			&m.EmptyBefore, &m.EmptyAfter, &m.FullBefore, &m.FullAfter,
			&m.ReferenceType, &m.ReferenceID, &m.Notes, &m.CreatedBy, &m.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		movements = append(movements, m)
	}
	return movements, total, nil
}

func (r *RefillableRepository) GetByProductID(ctx context.Context, productID uuid.UUID) (*domain.RefillableContainer, error) {
	query := `
		SELECT id, product_id, container_type, empty_count, full_count, notes, created_at, updated_at
		FROM refillable_containers WHERE product_id = $1
	`
	var c domain.RefillableContainer
	err := r.db.QueryRowContext(ctx, query, productID).Scan(
		&c.ID, &c.ProductID, &c.ContainerType, &c.EmptyCount, &c.FullCount, &c.Notes, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // Not found is okay, just means not refillable
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

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

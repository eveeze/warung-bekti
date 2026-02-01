package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/google/uuid"
)

type UserRepository struct {
	db *database.PostgresDB
}

func NewUserRepository(db *database.PostgresDB) domain.UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, name, email, password_hash, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	if user.ID == uuid.Nil { user.ID = uuid.New() }
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	return r.db.QueryRowContext(ctx, query,
		user.ID, user.Name, user.Email, user.PasswordHash, user.Role, user.IsActive, user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID)
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, name, email, password_hash, role, is_active, last_login_at, created_at, updated_at FROM users WHERE id = $1 AND deleted_at IS NULL`
	var user domain.User
	var roleStr string
	err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &roleStr, &user.IsActive, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { return nil, domain.ErrNotFound }
		return nil, err
	}
	user.Role = domain.UserRole(roleStr)
	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, name, email, password_hash, role, is_active, last_login_at, created_at, updated_at FROM users WHERE email = $1 AND is_active = true AND deleted_at IS NULL`
	var user domain.User
	var roleStr string
	err := r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &roleStr, &user.IsActive, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { return nil, domain.ErrNotFound }
		return nil, err
	}
	user.Role = domain.UserRole(roleStr)
	return &user, nil
}

func (r *UserRepository) List(ctx context.Context, p domain.UserListParams) ([]domain.User, int64, error) {
	var users []domain.User
	var total int64
	offset := (p.Page - 1) * p.PerPage

	countQuery := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`
	if p.Search != "" {
		countQuery += fmt.Sprintf(" AND (name ILIKE '%%%s%%' OR email ILIKE '%%%s%%')", p.Search, p.Search)
	}
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, name, email, role, is_active, last_login_at, created_at, updated_at FROM users WHERE deleted_at IS NULL`
	if p.Search != "" {
		query += fmt.Sprintf(" AND (name ILIKE '%%%s%%' OR email ILIKE '%%%s%%')", p.Search, p.Search)
	}
	query += " ORDER BY created_at DESC LIMIT $1 OFFSET $2"

	rows, err := r.db.QueryContext(ctx, query, p.PerPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var u domain.User
		var roleStr string
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &roleStr, &u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		u.Role = domain.UserRole(roleStr)
		users = append(users, u)
	}
	return users, total, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `UPDATE users SET name=$1, email=$2, password_hash=$3, role=$4, is_active=$5, updated_at=NOW() WHERE id=$6 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, user.Name, user.Email, user.PasswordHash, user.Role, user.IsActive, user.ID)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET deleted_at = NOW(), is_active = false WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET last_login_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
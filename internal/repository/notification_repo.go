package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/database"
)

type Notification struct {
	ID        uuid.UUID       `json:"id"`
	UserID    *uuid.UUID      `json:"user_id"`
	Title     string          `json:"title"`
	Message   string          `json:"message"`
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	IsRead    bool            `json:"is_read"`
	CreatedAt time.Time       `json:"created_at"`
}

type NotificationRepository struct {
	db *database.PostgresDB
}

func NewNotificationRepository(db *database.PostgresDB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, n *Notification) error {
	query := `
		INSERT INTO notifications (user_id, title, message, type, data, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`
	// Handle nil UserID
	var userID sql.NullString
	if n.UserID != nil {
		userID.String = n.UserID.String()
		userID.Valid = true
	}

	return r.db.QueryRowContext(ctx, query,
		userID,
		n.Title,
		n.Message,
		n.Type,
		n.Data,
		n.IsRead,
		time.Now(),
	).Scan(&n.ID, &n.CreatedAt)
}

func (r *NotificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Notification, error) {
	query := `
		SELECT id, user_id, title, message, type, data, is_read, created_at
		FROM notifications
		WHERE user_id = $1 OR user_id IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		var uid sql.NullString
		if err := rows.Scan(
			&n.ID,
			&uid,
			&n.Title,
			&n.Message,
			&n.Type,
			&n.Data,
			&n.IsRead,
			&n.CreatedAt,
		); err != nil {
			return nil, err
		}
		if uid.Valid {
			id, _ := uuid.Parse(uid.String)
			n.UserID = &id
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE notifications SET is_read = TRUE WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE notifications SET is_read = TRUE WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

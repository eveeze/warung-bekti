package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/database"
)

type AuditAction string

const (
	AuditActionCreate  AuditAction = "create"
	AuditActionUpdate  AuditAction = "update"
	AuditActionDelete  AuditAction = "delete"
	AuditActionLogin   AuditAction = "login"
	AuditActionLogout  AuditAction = "logout"
	AuditActionView    AuditAction = "view"
	AuditActionExport  AuditAction = "export"
	AuditActionImport  AuditAction = "import"
	AuditActionApprove AuditAction = "approve"
	AuditActionReject  AuditAction = "reject"
)

type AuditLog struct {
	ID         uuid.UUID   `json:"id"`
	UserID     *uuid.UUID  `json:"user_id,omitempty"`
	Username   *string     `json:"username,omitempty"`
	UserRole   *string     `json:"user_role,omitempty"`
	Action     AuditAction `json:"action"`
	EntityType string      `json:"entity_type"`
	EntityID   *uuid.UUID  `json:"entity_id,omitempty"`
	EntityName *string     `json:"entity_name,omitempty"`
	OldValues  interface{} `json:"old_values,omitempty"`
	NewValues  interface{} `json:"new_values,omitempty"`
	IPAddress  *string     `json:"ip_address,omitempty"`
	UserAgent  *string     `json:"user_agent,omitempty"`
	RequestID  *string     `json:"request_id,omitempty"`
	Notes      *string     `json:"notes,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
}

type AuditRepository struct {
	db *database.PostgresDB
}

func NewAuditRepository(db *database.PostgresDB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Log(ctx context.Context, log *AuditLog) error {
	query := `
		INSERT INTO audit_logs (
			user_id, username, user_role, action, entity_type, entity_id, entity_name,
			old_values, new_values, ip_address, user_agent, request_id, notes, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	
	oldJSON, _ := json.Marshal(log.OldValues)
	newJSON, _ := json.Marshal(log.NewValues)
	
	_, err := r.db.ExecContext(ctx, query,
		log.UserID, log.Username, log.UserRole, log.Action, log.EntityType, log.EntityID, log.EntityName,
		oldJSON, newJSON, log.IPAddress, log.UserAgent, log.RequestID, log.Notes, time.Now(),
	)
	return err
}

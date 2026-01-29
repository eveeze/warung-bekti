package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/repository"
)

func Audit(auditRepo *repository.AuditRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For now, we manually log important actions in Handlers.
			// This middleware can be used for general "Access" logging or capturing RequestID.
			
			// Generate RequestID
			requestID := uuid.New().String()
			ctx := context.WithValue(r.Context(), "request_id", requestID)
			w.Header().Set("X-Request-ID", requestID)
			
			next.ServeHTTP(w, r.WithContext(ctx))
			
			// We could log every request here, but that might be noisy.
			// "Who modified what" is better handled explicitly or via a smarter wrapper.
			// Let's keep it simple for now: RequestID generation.
		})
	}
}

// Helper to log audit from handlers
func LogAudit(ctx context.Context, auditRepo *repository.AuditRepository, action repository.AuditAction, entityType string, entityID *uuid.UUID, notes string) {
	// Extract user from context
	claims := GetUserFromContext(ctx)
	var userID *uuid.UUID
	var username, role *string
	
	if claims != nil {
		uid, _ := uuid.Parse(claims.UserID)
		userID = &uid
		username = &claims.Username
		role = &claims.Role
	}
	
	reqID, _ := ctx.Value("request_id").(string)

	log := &repository.AuditLog{
		UserID:     userID,
		Username:   username,
		UserRole:   role,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		RequestID:  &reqID,
		Notes:      &notes,
		CreatedAt:  time.Now(),
	}
	
	// Fire and forget (async) to not block response?
	// or synchronous for strict audit? Synchronous is safer for money.
	go auditRepo.Log(context.Background(), log)
}

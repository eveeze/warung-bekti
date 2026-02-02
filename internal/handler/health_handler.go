package handler

import (
	"net/http"

	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/pkg/response"
	"github.com/eveeze/warung-backend/internal/storage"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db    *database.PostgresDB
	redis *database.RedisClient
	r2    *storage.R2Client
}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler(db *database.PostgresDB, redis *database.RedisClient, r2 *storage.R2Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redis, r2: r2}
}

// Health returns the health status of all services
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	status := map[string]interface{}{
		"status": "healthy",
		"services": map[string]string{
			"api": "up",
		},
	}

	// Check database
	if h.db != nil {
		if err := h.db.Health(ctx); err != nil {
			status["status"] = "degraded"
			status["services"].(map[string]string)["database"] = "down"
		} else {
			status["services"].(map[string]string)["database"] = "up"
		}
	}

	// Check Redis
	if h.redis != nil {
		if err := h.redis.Health(ctx); err != nil {
			status["status"] = "degraded"
			status["services"].(map[string]string)["redis"] = "down"
		} else {
			status["services"].(map[string]string)["redis"] = "up"
		}
	}

	// Check R2
	if h.r2 != nil {
		if err := h.r2.Health(ctx); err != nil {
			status["status"] = "degraded"
			status["services"].(map[string]string)["r2"] = "down"
		} else {
			status["services"].(map[string]string)["r2"] = "up"
		}
	}

	statusCode := http.StatusOK
	if status["status"] == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	response.JSON(w, statusCode, status)
}

// Ready returns whether the service is ready to accept requests
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check critical dependencies
	if h.db != nil {
		if err := h.db.Health(ctx); err != nil {
			response.ServiceUnavailable(w, "Database not ready")
			return
		}
	}

	response.OK(w, "Service is ready", nil)
}

// Live returns whether the service is alive
func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	response.OK(w, "Service is alive", nil)
}

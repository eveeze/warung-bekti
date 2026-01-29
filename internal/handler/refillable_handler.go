package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/middleware"
	"github.com/eveeze/warung-backend/internal/pkg/response"
	"github.com/eveeze/warung-backend/internal/service"
)

type RefillableHandler struct {
	refillableSvc *service.RefillableService
}

func NewRefillableHandler(refillableSvc *service.RefillableService) *RefillableHandler {
	return &RefillableHandler{refillableSvc: refillableSvc}
}

func (h *RefillableHandler) GetContainers(w http.ResponseWriter, r *http.Request) {
	containers, err := h.refillableSvc.GetContainers(r.Context())
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}
	response.OK(w, "Containers retrieved", containers)
}

func (h *RefillableHandler) AdjustStock(w http.ResponseWriter, r *http.Request) {
	var input domain.ContainerMovement
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "Invalid body")
		return
	}
	
	claims := middleware.GetUserFromContext(r.Context())
	username := "system"
	if claims != nil {
		username = claims.Username
	}
	input.CreatedBy = &username

	result, err := h.refillableSvc.AdjustStock(r.Context(), input)
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}
	response.Created(w, "Stock adjusted", result)
}

// Helper to parse UUID
func parseUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}

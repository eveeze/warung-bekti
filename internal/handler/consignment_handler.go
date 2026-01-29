package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/pkg/response"
	"github.com/eveeze/warung-backend/internal/service"
)

type ConsignmentHandler struct {
	consignmentSvc *service.ConsignmentService
}

func NewConsignmentHandler(consignmentSvc *service.ConsignmentService) *ConsignmentHandler {
	return &ConsignmentHandler{consignmentSvc: consignmentSvc}
}

func (h *ConsignmentHandler) CreateConsignor(w http.ResponseWriter, r *http.Request) {
	var input domain.CreateConsignorInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "Invalid body")
		return
	}
	
	consignor, err := h.consignmentSvc.CreateConsignor(r.Context(), input)
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}
	response.Created(w, "Consignor created", consignor)
}

func (h *ConsignmentHandler) ListConsignors(w http.ResponseWriter, r *http.Request) {
	list, err := h.consignmentSvc.ListConsignors(r.Context())
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}
	response.OK(w, "Consignors retrieved", list)
}

func (h *ConsignmentHandler) UpdateConsignor(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid ID")
		return
	}

	var input domain.UpdateConsignorInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "Invalid body")
		return
	}

	consignor, err := h.consignmentSvc.UpdateConsignor(r.Context(), id, input)
	if err == domain.ErrNotFound {
		response.NotFound(w, "Consignor not found")
		return
	}
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}
	response.OK(w, "Consignor updated", consignor)
}

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

type POSHandler struct {
	posSvc *service.POSService
}

func NewPOSHandler(posSvc *service.POSService) *POSHandler {
	return &POSHandler{posSvc: posSvc}
}

// -- Held Carts --

func (h *POSHandler) HoldCart(w http.ResponseWriter, r *http.Request) {
	var input domain.HoldCartInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "Invalid body")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	username := "system"
	if claims != nil {
		username = claims.Username
	}
	input.HeldBy = username

	cart, err := h.posSvc.HoldCart(r.Context(), input)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Created(w, "Cart held", cart)
}

func (h *POSHandler) ListHeldCarts(w http.ResponseWriter, r *http.Request) {
	carts, err := h.posSvc.ListHeldCarts(r.Context())
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}
	response.OK(w, "Held carts retrieved", carts)
}

func (h *POSHandler) GetHeldCart(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid ID")
		return
	}

	cart, err := h.posSvc.GetHeldCart(r.Context(), id)
	if err == domain.ErrNotFound {
		response.NotFound(w, "Cart not found")
		return
	}
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}
	response.OK(w, "Cart retrieved", cart)
}

func (h *POSHandler) ResumeCart(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid ID")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	username := "system"
	if claims != nil {
		username = claims.Username
	}

	cart, err := h.posSvc.ResumeCart(r.Context(), id, username)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.OK(w, "Cart resumed", cart)
}

func (h *POSHandler) DiscardCart(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid ID")
		return
	}

	if err := h.posSvc.DiscardCart(r.Context(), id); err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.OK(w, "Cart discarded", nil)
}

// -- Refunds --

func (h *POSHandler) CreateRefund(w http.ResponseWriter, r *http.Request) {
	var input domain.CreateRefundInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "Invalid body")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	username := "system"
	if claims != nil {
		username = claims.Username
	}
	input.RequestedBy = username

	refund, err := h.posSvc.CreateRefund(r.Context(), input)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Created(w, "Refund request created", refund)
}

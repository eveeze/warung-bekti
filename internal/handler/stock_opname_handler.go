package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/middleware"
	"github.com/eveeze/warung-backend/internal/pkg/response"
	"github.com/eveeze/warung-backend/internal/pkg/validator"
	"github.com/eveeze/warung-backend/internal/service"
)

// StockOpnameHandler handles stock opname HTTP requests
type StockOpnameHandler struct {
	opnameSvc *service.StockOpnameService
}

// NewStockOpnameHandler creates a new StockOpnameHandler
func NewStockOpnameHandler(opnameSvc *service.StockOpnameService) *StockOpnameHandler {
	return &StockOpnameHandler{opnameSvc: opnameSvc}
}

// StartSession starts a new opname session
// POST /stock-opname/sessions
func (h *StockOpnameHandler) StartSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Notes *string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
		response.BadRequest(w, "Invalid request body")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	createdBy := "system"
	if claims != nil {
		createdBy = claims.Username
	}

	input := domain.StartOpnameInput{
		Notes:     req.Notes,
		CreatedBy: createdBy,
	}

	session, err := h.opnameSvc.StartSession(r.Context(), input)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Created(w, "Stock opname session started", session)
}

// GetSession gets a session by ID
// GET /stock-opname/sessions/{id}
func (h *StockOpnameHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid session ID")
		return
	}

	session, err := h.opnameSvc.GetSession(r.Context(), id)
	if err != nil {
		if err == domain.ErrNotFound {
			response.NotFound(w, "Session not found")
			return
		}
		response.BadRequest(w, err.Error())
		return
	}

	response.OK(w, "Session retrieved", session)
}

// ListSessions lists all sessions
// GET /stock-opname/sessions
func (h *StockOpnameHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	statusStr := r.URL.Query().Get("status")

	var status *domain.OpnameStatus
	if statusStr != "" {
		s := domain.OpnameStatus(statusStr)
		status = &s
	}

	sessions, total, err := h.opnameSvc.ListSessions(r.Context(), status, page, perPage)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	meta := response.NewMeta(page, perPage, total)
	response.SuccessWithMeta(w, http.StatusOK, "Sessions retrieved", sessions, meta)
}

// RecordCount records a physical count
// POST /stock-opname/sessions/{id}/items
func (h *StockOpnameHandler) RecordCount(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := r.PathValue("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid session ID")
		return
	}

	var req struct {
		ProductID     string  `json:"product_id"`
		PhysicalStock int     `json:"physical_stock"`
		Notes         *string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	v := validator.New()
	v.Required("product_id", req.ProductID, "Product ID is required")
	v.Min("physical_stock", req.PhysicalStock, 0, "Physical stock cannot be negative")

	if v.HasErrors() {
		response.ValidationError(w, v.Errors())
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		response.BadRequest(w, "Invalid product ID format")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	countedBy := "system"
	if claims != nil {
		countedBy = claims.Username
	}

	input := domain.RecordCountInput{
		SessionID:     sessionID,
		ProductID:     productID,
		PhysicalStock: req.PhysicalStock,
		Notes:         req.Notes,
		CountedBy:     countedBy,
	}

	item, err := h.opnameSvc.RecordCount(r.Context(), input)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.OK(w, "Count recorded", item)
}

// FinalizeSession finalizes a session
// POST /stock-opname/sessions/{id}/finalize
func (h *StockOpnameHandler) FinalizeSession(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := r.PathValue("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid session ID")
		return
	}

	var req struct {
		ApplyAdjustments bool `json:"apply_adjustments"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
		response.BadRequest(w, "Invalid request body")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	completedBy := "system"
	if claims != nil {
		completedBy = claims.Username
	}

	input := domain.FinalizeOpnameInput{
		SessionID:        sessionID,
		CompletedBy:      completedBy,
		ApplyAdjustments: req.ApplyAdjustments,
	}

	report, err := h.opnameSvc.FinalizeSession(r.Context(), input)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.OK(w, "Session finalized", report)
}

// GetVarianceReport gets the variance report for a session
// GET /stock-opname/sessions/{id}/variance
func (h *StockOpnameHandler) GetVarianceReport(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := r.PathValue("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid session ID")
		return
	}

	report, err := h.opnameSvc.GetVarianceReport(r.Context(), sessionID)
	if err != nil {
		if err == domain.ErrNotFound {
			response.NotFound(w, "Session not found")
			return
		}
		response.BadRequest(w, err.Error())
		return
	}

	response.OK(w, "Variance report generated", report)
}

// CancelSession cancels a session
// POST /stock-opname/sessions/{id}/cancel
func (h *StockOpnameHandler) CancelSession(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := r.PathValue("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid session ID")
		return
	}

	if err := h.opnameSvc.CancelSession(r.Context(), sessionID); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.OK(w, "Session cancelled", nil)
}

// GetShoppingList gets the auto-generated shopping list
// GET /stock-opname/shopping-list
func (h *StockOpnameHandler) GetShoppingList(w http.ResponseWriter, r *http.Request) {
	list, err := h.opnameSvc.GetShoppingList(r.Context())
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.OK(w, "Shopping list generated", list)
}

// GetNearExpiryReport gets items nearing expiry
// GET /stock-opname/near-expiry
func (h *StockOpnameHandler) GetNearExpiryReport(w http.ResponseWriter, r *http.Request) {
	daysAhead, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if daysAhead <= 0 {
		daysAhead = 30
	}

	report, err := h.opnameSvc.GetNearExpiryReport(r.Context(), daysAhead)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.OK(w, "Near expiry report generated", report)
}

package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/middleware"
	"github.com/eveeze/warung-backend/internal/pkg/response"
	"github.com/eveeze/warung-backend/internal/pkg/validator"
	"github.com/eveeze/warung-backend/internal/service"
)

type CashFlowHandler struct {
	cashFlowSvc *service.CashFlowService
}

func NewCashFlowHandler(cashFlowSvc *service.CashFlowService) *CashFlowHandler {
	return &CashFlowHandler{cashFlowSvc: cashFlowSvc}
}

// GetCategories lists all categories
func (h *CashFlowHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.cashFlowSvc.GetCategories(r.Context())
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}
	response.OK(w, "Categories retrieved", categories)
}

// OpenDrawer opens a new session
func (h *CashFlowHandler) OpenDrawer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OpeningBalance int64   `json:"opening_balance"`
		Notes          *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid body")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	username := "system"
	if claims != nil {
		username = claims.Username
	}

	input := domain.OpenDrawerInput{
		OpeningBalance: req.OpeningBalance,
		OpenedBy:       username,
		Notes:          req.Notes,
	}

	session, err := h.cashFlowSvc.OpenDrawer(r.Context(), input)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Created(w, "Drawer opened", session)
}

// CloseDrawer closes an open session
func (h *CashFlowHandler) CloseDrawer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID      string  `json:"session_id"`
		ClosingBalance int64   `json:"closing_balance"`
		Notes          *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid body")
		return
	}

	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		response.BadRequest(w, "Invalid session ID")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	username := "system"
	if claims != nil {
		username = claims.Username
	}

	input := domain.CloseDrawerInput{
		SessionID:      sessionID,
		ClosingBalance: req.ClosingBalance,
		ClosedBy:       username,
		Notes:          req.Notes,
	}

	session, err := h.cashFlowSvc.CloseDrawer(r.Context(), input)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.OK(w, "Drawer closed", session)
}

// GetCurrentSession info
func (h *CashFlowHandler) GetCurrentSession(w http.ResponseWriter, r *http.Request) {
	session, err := h.cashFlowSvc.GetCurrentSession(r.Context())
	if err == domain.ErrNotFound {
		response.NotFound(w, "No open session")
		return
	}
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}
	response.OK(w, "Current session retrieved", session)
}

// RecordCashFlow adds new record
func (h *CashFlowHandler) RecordCashFlow(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CategoryID  string  `json:"category_id"`
		Type        string  `json:"type"`
		Amount      int64   `json:"amount"`
		Description *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid body")
		return
	}

	v := validator.New()
	v.Required("type", req.Type, "Type is required")
	v.Positive("amount", req.Amount, "Amount must be positive")
	if v.HasErrors() {
		response.ValidationError(w, v.Errors())
		return
	}

	var categoryID *uuid.UUID
	if req.CategoryID != "" {
		cid, err := uuid.Parse(req.CategoryID)
		if err == nil {
			categoryID = &cid
		}
	}

	claims := middleware.GetUserFromContext(r.Context())
	username := "system"
	if claims != nil {
		username = claims.Username
	}

	input := domain.CashFlowInput{
		CategoryID:  categoryID,
		Type:        domain.CashFlowType(req.Type),
		Amount:      req.Amount,
		Description: req.Description,
		CreatedBy:   username,
	}

	record, err := h.cashFlowSvc.RecordCashFlow(r.Context(), input)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Created(w, "Cash flow recorded", record)
}

// ListCashFlows gets filtered list
func (h *CashFlowHandler) ListCashFlows(w http.ResponseWriter, r *http.Request) {
	filter := domain.CashFlowFilter{
		Page:    1,
		PerPage: 20,
	}
	
	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil { filter.Page = p }
	if pp, err := strconv.Atoi(r.URL.Query().Get("per_page")); err == nil { filter.PerPage = pp }
	
	sessionIDStr := r.URL.Query().Get("session_id")
	if sessionIDStr != "" {
		if sid, err := uuid.Parse(sessionIDStr); err == nil {
			filter.SessionID = &sid
		}
	}
	
	typeStr := r.URL.Query().Get("type")
	if typeStr != "" {
		t := domain.CashFlowType(typeStr)
		filter.Type = &t
	}
	
	dateFromStr := r.URL.Query().Get("date_from")
	if dateFromStr != "" {
		if t, err := time.Parse(time.RFC3339, dateFromStr); err == nil {
			filter.DateFrom = &t
		}
	}

	records, total, err := h.cashFlowSvc.ListCashFlows(r.Context(), filter)
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}

	meta := response.NewMeta(filter.Page, filter.PerPage, total)
	response.SuccessWithMeta(w, http.StatusOK, "Cash flows retrieved", records, meta)
}

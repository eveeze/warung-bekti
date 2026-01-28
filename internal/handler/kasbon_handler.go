package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/pkg/response"
	"github.com/eveeze/warung-backend/internal/pkg/validator"
	"github.com/eveeze/warung-backend/internal/repository"
)

// KasbonHandler handles kasbon endpoints
type KasbonHandler struct {
	kasbonRepo   *repository.KasbonRepository
	customerRepo *repository.CustomerRepository
}

// NewKasbonHandler creates a new KasbonHandler
func NewKasbonHandler(kasbonRepo *repository.KasbonRepository, customerRepo *repository.CustomerRepository) *KasbonHandler {
	return &KasbonHandler{kasbonRepo: kasbonRepo, customerRepo: customerRepo}
}

// GetHistory retrieves kasbon history for a customer
func (h *KasbonHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	customerID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid customer ID")
		return
	}

	query := r.URL.Query()
	filter := domain.KasbonFilter{CustomerID: &customerID, Page: 1, PerPage: 20}

	if page := query.Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			filter.Page = p
		}
	}
	if perPage := query.Get("per_page"); perPage != "" {
		if pp, err := strconv.Atoi(perPage); err == nil {
			filter.PerPage = pp
		}
	}
	if typeStr := query.Get("type"); typeStr != "" {
		t := domain.KasbonType(typeStr)
		filter.Type = &t
	}

	records, total, err := h.kasbonRepo.GetByCustomer(r.Context(), customerID, filter)
	if err != nil {
		response.InternalServerError(w, "Failed to get kasbon history")
		return
	}

	meta := response.NewMeta(filter.Page, filter.PerPage, total)
	response.SuccessWithMeta(w, http.StatusOK, "Kasbon history retrieved", records, meta)
}

// RecordPayment records a kasbon payment
func (h *KasbonHandler) RecordPayment(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	customerID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid customer ID")
		return
	}

	var input struct {
		Amount    int64   `json:"amount"`
		Notes     *string `json:"notes,omitempty"`
		CreatedBy *string `json:"created_by,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	v := validator.New()
	v.Positive("amount", input.Amount, "Amount must be positive")
	if v.HasErrors() {
		response.ValidationError(w, v.Errors())
		return
	}

	// Check if customer exists
	customer, err := h.customerRepo.GetByID(r.Context(), customerID)
	if err == domain.ErrNotFound {
		response.NotFound(w, "Customer not found")
		return
	}
	if err != nil {
		response.InternalServerError(w, "Failed to get customer")
		return
	}

	if customer.CurrentDebt == 0 {
		response.BadRequest(w, "Customer has no debt")
		return
	}

	paymentInput := domain.KasbonPaymentInput{
		CustomerID: customerID,
		Amount:     input.Amount,
		Notes:      input.Notes,
		CreatedBy:  input.CreatedBy,
	}

	record, err := h.kasbonRepo.CreatePayment(r.Context(), paymentInput)
	if err != nil {
		response.InternalServerError(w, "Failed to record payment")
		return
	}

	response.Created(w, "Payment recorded successfully", record)
}

// GetSummary retrieves kasbon summary for a customer
func (h *KasbonHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	customerID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid customer ID")
		return
	}

	summary, err := h.kasbonRepo.GetSummary(r.Context(), customerID)
	if err == domain.ErrNotFound {
		response.NotFound(w, "Customer not found")
		return
	}
	if err != nil {
		response.InternalServerError(w, "Failed to get kasbon summary")
		return
	}

	response.OK(w, "Kasbon summary retrieved", summary)
}

// GetReport retrieves overall kasbon report
func (h *KasbonHandler) GetReport(w http.ResponseWriter, r *http.Request) {
	report, err := h.kasbonRepo.GetReport(r.Context())
	if err != nil {
		response.InternalServerError(w, "Failed to get kasbon report")
		return
	}

	response.OK(w, "Kasbon report retrieved", report)
}

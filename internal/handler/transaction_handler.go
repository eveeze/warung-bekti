package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/pkg/response"
	"github.com/eveeze/warung-backend/internal/pkg/validator"
	"github.com/eveeze/warung-backend/internal/repository"
	"github.com/eveeze/warung-backend/internal/service"
)

// TransactionHandler handles transaction endpoints
type TransactionHandler struct {
	svc  *service.TransactionService
	repo *repository.TransactionRepository
}

// NewTransactionHandler creates a new TransactionHandler
func NewTransactionHandler(svc *service.TransactionService, repo *repository.TransactionRepository) *TransactionHandler {
	return &TransactionHandler{svc: svc, repo: repo}
}

// Create creates a new transaction (checkout)
func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input domain.TransactionCreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	v := validator.New()
	v.Custom("items", len(input.Items) > 0, "Cart cannot be empty")
	v.InSlice("payment_method", string(input.PaymentMethod), 
		[]string{"cash", "kasbon", "transfer", "qris"}, "Invalid payment method")
	
	for i, item := range input.Items {
		v.Custom("items.product_id", item.ProductID != uuid.Nil, "Product ID is required")
		v.Min("items.quantity", item.Quantity, 1, "Quantity must be at least 1")
		if v.HasErrors() {
			v.Errors().Add("item_index", strconv.Itoa(i))
			break
		}
	}

	if v.HasErrors() {
		response.ValidationError(w, v.Errors())
		return
	}

	transaction, err := h.svc.CreateTransaction(r.Context(), input)
	if err != nil {
		switch err {
		case domain.ErrEmptyCart:
			response.BadRequest(w, "Cart is empty")
		case domain.ErrInsufficientStock:
			response.BadRequest(w, err.Error())
		case domain.ErrCreditLimitExceeded:
			response.BadRequest(w, "Customer credit limit exceeded")
		case domain.ErrInvalidPaymentAmount:
			response.BadRequest(w, "Payment amount is less than total")
		case domain.ErrCustomerInactive:
			response.BadRequest(w, "Customer is inactive")
		default:
			response.InternalServerError(w, err.Error())
		}
		return
	}

	response.Created(w, "Transaction created successfully", transaction)
}

// GetByID retrieves a transaction by ID
func (h *TransactionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid transaction ID")
		return
	}

	transaction, err := h.repo.GetByID(r.Context(), id)
	if err == domain.ErrNotFound {
		response.NotFound(w, "Transaction not found")
		return
	}
	if err != nil {
		response.InternalServerError(w, "Failed to get transaction")
		return
	}

	response.OK(w, "Transaction retrieved", transaction)
}

// List retrieves transactions with filtering
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	filter := domain.TransactionFilter{Page: 1, PerPage: 20, SortBy: "created_at", SortOrder: "desc"}

	if customerID := query.Get("customer_id"); customerID != "" {
		if id, err := uuid.Parse(customerID); err == nil {
			filter.CustomerID = &id
		}
	}
	if status := query.Get("status"); status != "" {
		s := domain.TransactionStatus(status)
		filter.Status = &s
	}
	if method := query.Get("payment_method"); method != "" {
		m := domain.PaymentMethod(method)
		filter.PaymentMethod = &m
	}
	if dateFrom := query.Get("date_from"); dateFrom != "" {
		if t, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filter.DateFrom = &t
		}
	}
	if dateTo := query.Get("date_to"); dateTo != "" {
		if t, err := time.Parse("2006-01-02", dateTo); err == nil {
			endOfDay := t.Add(24*time.Hour - time.Second)
			filter.DateTo = &endOfDay
		}
	}
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

	transactions, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		response.InternalServerError(w, "Failed to list transactions")
		return
	}

	meta := response.NewMeta(filter.Page, filter.PerPage, total)
	response.SuccessWithMeta(w, http.StatusOK, "Transactions retrieved", transactions, meta)
}

// Cancel cancels a transaction
func (h *TransactionHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid transaction ID")
		return
	}

	if err := h.svc.CancelTransaction(r.Context(), id); err != nil {
		switch err {
		case domain.ErrNotFound:
			response.NotFound(w, "Transaction not found")
		case domain.ErrTransactionCancelled:
			response.BadRequest(w, "Transaction is already cancelled")
		default:
			response.InternalServerError(w, "Failed to cancel transaction")
		}
		return
	}

	response.OK(w, "Transaction cancelled", nil)
}

// Calculate calculates cart totals without creating transaction
func (h *TransactionHandler) Calculate(w http.ResponseWriter, r *http.Request) {
	var input domain.CartCalculateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if len(input.Items) == 0 {
		response.BadRequest(w, "Cart is empty")
		return
	}

	result, err := h.svc.CalculateCart(r.Context(), input)
	if err != nil {
		response.InternalServerError(w, "Failed to calculate cart")
		return
	}

	response.OK(w, "Cart calculated", result)
}

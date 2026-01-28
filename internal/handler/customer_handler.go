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

// CustomerHandler handles customer endpoints
type CustomerHandler struct {
	repo *repository.CustomerRepository
}

// NewCustomerHandler creates a new CustomerHandler
func NewCustomerHandler(repo *repository.CustomerRepository) *CustomerHandler {
	return &CustomerHandler{repo: repo}
}

// Create creates a new customer
func (h *CustomerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input domain.CustomerCreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	v := validator.New()
	v.Required("name", input.Name, "Name is required")
	v.MinLength("name", input.Name, 2, "Name must be at least 2 characters")
	if input.Phone != nil {
		v.Phone("phone", *input.Phone, "Invalid phone number format")
	}
	if v.HasErrors() {
		response.ValidationError(w, v.Errors())
		return
	}

	customer, err := h.repo.Create(r.Context(), input)
	if err != nil {
		response.InternalServerError(w, "Failed to create customer")
		return
	}

	response.Created(w, "Customer created successfully", customer)
}

// GetByID retrieves a customer by ID
func (h *CustomerHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid customer ID")
		return
	}

	customer, err := h.repo.GetByID(r.Context(), id)
	if err == domain.ErrNotFound {
		response.NotFound(w, "Customer not found")
		return
	}
	if err != nil {
		response.InternalServerError(w, "Failed to get customer")
		return
	}

	response.OK(w, "Customer retrieved", customer)
}

// List retrieves customers with filtering
func (h *CustomerHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	filter := domain.CustomerFilter{Page: 1, PerPage: 20, SortBy: "name", SortOrder: "asc"}
	if search := query.Get("search"); search != "" {
		filter.Search = &search
	}
	if query.Get("has_debt") == "true" {
		hasDebt := true
		filter.HasDebt = &hasDebt
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

	customers, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		response.InternalServerError(w, "Failed to list customers")
		return
	}

	meta := response.NewMeta(filter.Page, filter.PerPage, total)
	response.SuccessWithMeta(w, http.StatusOK, "Customers retrieved", customers, meta)
}

// Update updates a customer
func (h *CustomerHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid customer ID")
		return
	}

	var input domain.CustomerUpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	customer, err := h.repo.Update(r.Context(), id, input)
	if err == domain.ErrNotFound {
		response.NotFound(w, "Customer not found")
		return
	}
	if err != nil {
		response.InternalServerError(w, "Failed to update customer")
		return
	}

	response.OK(w, "Customer updated", customer)
}

// Delete soft deletes a customer
func (h *CustomerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid customer ID")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err == domain.ErrNotFound {
		response.NotFound(w, "Customer not found")
		return
	} else if err != nil {
		response.InternalServerError(w, "Failed to delete customer")
		return
	}

	response.NoContent(w)
}

// GetWithDebt retrieves all customers with outstanding debt
func (h *CustomerHandler) GetWithDebt(w http.ResponseWriter, r *http.Request) {
	customers, err := h.repo.GetCustomersWithDebt(r.Context())
	if err != nil {
		response.InternalServerError(w, "Failed to get customers with debt")
		return
	}

	response.OK(w, "Customers with debt retrieved", customers)
}

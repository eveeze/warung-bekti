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

// ProductHandler handles product endpoints
type ProductHandler struct {
	repo *repository.ProductRepository
}

// NewProductHandler creates a new ProductHandler
func NewProductHandler(repo *repository.ProductRepository) *ProductHandler {
	return &ProductHandler{repo: repo}
}

// Create creates a new product
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input domain.ProductCreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Validate
	v := validator.New()
	v.Required("name", input.Name, "Name is required")
	v.MinLength("name", input.Name, 2, "Name must be at least 2 characters")
	v.Required("unit", input.Unit, "Unit is required")
	v.Positive("base_price", input.BasePrice, "Base price must be positive")
	v.NonNegative("cost_price", input.CostPrice, "Cost price cannot be negative")

	if v.HasErrors() {
		response.ValidationError(w, v.Errors())
		return
	}

	product, err := h.repo.Create(r.Context(), input)
	if err != nil {
		response.InternalServerError(w, "Failed to create product")
		return
	}

	response.Created(w, "Product created successfully", product)
}

// GetByID retrieves a product by ID
func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid product ID")
		return
	}

	product, err := h.repo.GetByID(r.Context(), id)
	if err == domain.ErrNotFound {
		response.NotFound(w, "Product not found")
		return
	}
	if err != nil {
		response.InternalServerError(w, "Failed to get product")
		return
	}

	response.OK(w, "Product retrieved", product)
}

// GetByBarcode retrieves a product by barcode
func (h *ProductHandler) GetByBarcode(w http.ResponseWriter, r *http.Request) {
	barcode := r.URL.Query().Get("barcode")
	if barcode == "" {
		response.BadRequest(w, "Barcode is required")
		return
	}

	product, err := h.repo.GetByBarcode(r.Context(), barcode)
	if err == domain.ErrNotFound {
		response.NotFound(w, "Product not found")
		return
	}
	if err != nil {
		response.InternalServerError(w, "Failed to get product")
		return
	}

	response.OK(w, "Product retrieved", product)
}

// List retrieves products with filtering
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	filter := domain.DefaultFilter()
	if search := query.Get("search"); search != "" {
		filter.Search = &search
	}
	if catID := query.Get("category_id"); catID != "" {
		if id, err := uuid.Parse(catID); err == nil {
			filter.CategoryID = &id
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
	if sortBy := query.Get("sort_by"); sortBy != "" {
		filter.SortBy = sortBy
	}
	if sortOrder := query.Get("sort_order"); sortOrder != "" {
		filter.SortOrder = sortOrder
	}
	if query.Get("low_stock") == "true" {
		filter.LowStockOnly = true
	}

	products, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		response.InternalServerError(w, "Failed to list products")
		return
	}

	meta := response.NewMeta(filter.Page, filter.PerPage, total)
	response.SuccessWithMeta(w, http.StatusOK, "Products retrieved", products, meta)
}

// Update updates a product
func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid product ID")
		return
	}

	var input domain.ProductUpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	product, err := h.repo.Update(r.Context(), id, input)
	if err == domain.ErrNotFound {
		response.NotFound(w, "Product not found")
		return
	}
	if err != nil {
		response.InternalServerError(w, "Failed to update product")
		return
	}

	response.OK(w, "Product updated", product)
}

// Delete soft deletes a product
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid product ID")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err == domain.ErrNotFound {
		response.NotFound(w, "Product not found")
		return
	} else if err != nil {
		response.InternalServerError(w, "Failed to delete product")
		return
	}

	response.NoContent(w)
}

// AddPricingTier adds a pricing tier to a product
func (h *ProductHandler) AddPricingTier(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	productID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid product ID")
		return
	}

	var input domain.PricingTierInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	v := validator.New()
	v.Min("min_quantity", input.MinQuantity, 1, "Min quantity must be at least 1")
	v.Positive("price", input.Price, "Price must be positive")
	if v.HasErrors() {
		response.ValidationError(w, v.Errors())
		return
	}

	tier, err := h.repo.CreatePricingTier(r.Context(), productID, input)
	if err != nil {
		response.InternalServerError(w, "Failed to create pricing tier")
		return
	}

	response.Created(w, "Pricing tier created", tier)
}

// UpdatePricingTier updates a pricing tier
func (h *ProductHandler) UpdatePricingTier(w http.ResponseWriter, r *http.Request) {
	tierIDStr := r.PathValue("tierId")
	tierID, err := uuid.Parse(tierIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid tier ID")
		return
	}

	var input domain.PricingTierInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	tier, err := h.repo.UpdatePricingTier(r.Context(), tierID, input)
	if err == domain.ErrNotFound {
		response.NotFound(w, "Pricing tier not found")
		return
	}
	if err != nil {
		response.InternalServerError(w, "Failed to update pricing tier")
		return
	}

	response.OK(w, "Pricing tier updated", tier)
}

// DeletePricingTier deletes a pricing tier
func (h *ProductHandler) DeletePricingTier(w http.ResponseWriter, r *http.Request) {
	tierIDStr := r.PathValue("tierId")
	tierID, err := uuid.Parse(tierIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid tier ID")
		return
	}

	if err := h.repo.DeletePricingTier(r.Context(), tierID); err == domain.ErrNotFound {
		response.NotFound(w, "Pricing tier not found")
		return
	} else if err != nil {
		response.InternalServerError(w, "Failed to delete pricing tier")
		return
	}

	response.NoContent(w)
}

// GetLowStock retrieves products with low stock
func (h *ProductHandler) GetLowStock(w http.ResponseWriter, r *http.Request) {
	products, err := h.repo.GetLowStockProducts(r.Context())
	if err != nil {
		response.InternalServerError(w, "Failed to get low stock products")
		return
	}

	response.OK(w, "Low stock products retrieved", products)
}

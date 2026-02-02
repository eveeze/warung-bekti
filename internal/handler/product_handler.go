package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/pkg/response"
	"github.com/eveeze/warung-backend/internal/pkg/validator"
	"github.com/eveeze/warung-backend/internal/repository"
	"github.com/eveeze/warung-backend/internal/service"
	"github.com/eveeze/warung-backend/internal/storage"
)

// ProductHandler handles product endpoints
type ProductHandler struct {
	repo  *repository.ProductRepository
	minio *storage.MinioClient
	cache *service.CacheService
}

// NewProductHandler creates a new ProductHandler
func NewProductHandler(repo *repository.ProductRepository, minio *storage.MinioClient, cache *service.CacheService) *ProductHandler {
	return &ProductHandler{
		repo:  repo,
		minio: minio,
		cache: cache,
	}
}

// Create creates a new product
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input domain.ProductCreateInput

	// Check if multipart/form-data
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		// Limit to 10MB
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			response.BadRequest(w, "File too large or invalid multipart data")
			return
		}

		// Decode JSON from "data" field
		jsonData := r.FormValue("data")
		if jsonData == "" {
			response.BadRequest(w, "Missing 'data' field containing JSON payload")
			return
		}
		if err := json.Unmarshal([]byte(jsonData), &input); err != nil {
			response.BadRequest(w, "Invalid JSON in 'data' field")
			return
		}

		// Handle Image Upload
		file, header, err := r.FormFile("image")
		if err == nil {
			defer file.Close()

			// Validate
			if err := h.validateImage(header); err != nil {
				response.BadRequest(w, err.Error())
				return
			}

			// Process (Resize/Compress)
			reader, size, contentType, filename, err := h.processImage(file, header)
			if err != nil {
				response.InternalServerError(w, "Failed to process image")
				return
			}

			// Upload to MinIO
			// Use a simple path structure: products/{uuid-filename}
			objectName := fmt.Sprintf("products/%s", filename)
			url, err := h.minio.UploadFile(r.Context(), objectName, reader, size, contentType)
			if err != nil {
				response.InternalServerError(w, "Failed to upload image")
				return
			}

			input.ImageURL = &url
		} else if err != http.ErrMissingFile {
			response.BadRequest(w, "Error retrieving file")
			return
		}

	} else {
		// Standard JSON body
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			response.BadRequest(w, "Invalid request body")
			return
		}
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

	// Invalidate products cache
	_ = h.cache.InvalidatePattern(r.Context(), "products:list:*")

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
	if isActive := query.Get("is_active"); isActive != "" {
		if strings.ToLower(isActive) == "all" {
			filter.IsActive = nil
		} else if val, err := strconv.ParseBool(isActive); err == nil {
			filter.IsActive = &val
		}
	}
	if isStockActive := query.Get("is_stock_active"); isStockActive != "" {
		if val, err := strconv.ParseBool(isStockActive); err == nil {
			filter.IsStockActive = &val
		}
	}

	// Generate cache key based on filters
	cacheKey := fmt.Sprintf("products:list:%s:%s:%d:%d:%s:%s:%v:%v:%v",
		query.Get("search"),
		query.Get("category_id"),
		filter.Page,
		filter.PerPage,
		filter.SortBy,
		filter.SortOrder,
		filter.LowStockOnly,
		query.Get("is_active"),
		query.Get("is_stock_active"),
	)

	// Try cache first
	type CachedResponse struct {
		Products []domain.Product `json:"products"`
		Total    int64            `json:"total"`
	}
	
	var cached CachedResponse
	err := h.cache.Get(r.Context(), cacheKey, &cached)
	if err == nil {
		// Cache hit
		meta := response.NewMeta(filter.Page, filter.PerPage, cached.Total)
		response.SuccessWithMeta(w, http.StatusOK, "Products retrieved (cached)", cached.Products, meta)
		return
	}

	// Cache miss - fetch from DB
	products, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		response.InternalServerError(w, "Failed to list products")
		return
	}

	// Cache for 5 minutes
	cached = CachedResponse{Products: products, Total: total}
	_ = h.cache.Set(r.Context(), cacheKey, cached, 5*time.Minute)

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

	// Check if multipart/form-data
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			response.BadRequest(w, "File too large or invalid multipart data")
			return
		}

		jsonData := r.FormValue("data")
		if jsonData == "" {
			response.BadRequest(w, "Missing 'data' field containing JSON payload")
			return
		}
		if err := json.Unmarshal([]byte(jsonData), &input); err != nil {
			response.BadRequest(w, "Invalid JSON in 'data' field")
			return
		}

		// Handle Image Upload
		file, header, err := r.FormFile("image")
		if err == nil {
			defer file.Close()

			if err := h.validateImage(header); err != nil {
				response.BadRequest(w, err.Error())
				return
			}
			reader, size, contentType, filename, err := h.processImage(file, header)
			if err != nil {
				response.InternalServerError(w, "Failed to process image")
				return
			}

			// Get existing product to delete old image later
			existingProduct, err := h.repo.GetByID(r.Context(), id)
			if err != nil && err != domain.ErrNotFound {
				response.InternalServerError(w, "Failed to check existing product")
				return
			}

			// Upload new image
			objectName := fmt.Sprintf("products/%s", filename)
			url, err := h.minio.UploadFile(r.Context(), objectName, reader, size, contentType)
			if err != nil {
				response.InternalServerError(w, "Failed to upload image")
				return
			}

			input.ImageURL = &url

			// Delete old image if exists
			if existingProduct != nil && existingProduct.ImageURL != nil {
				// Extract object name from URL or store just the key? 
				// Our URL is typically http://.../bucket/products/xyz.jpg
				// MinIO client helper only knows how to upload/delete by object name.
				// We need to parse the object name from the URL.
				// Assuming standard structure: .../bucketName/objectName
				// Simple hack: Take substring after bucket name?
				// Better: Since we know the prefix "products/", we can just take the last part? No, nested paths.
				// Let's assume we can tolerate not deleting for now or just log errors if parsing fails.
				// Actually, `storage` package has `GetPublicURL`. We can't reverse it easily without robust parsing.
				// But we stored the Full URL.
				// Correct approach: Store OBJECT KEY in DB, or parse it here.
				// Let's defer "Delete old image" logic until later or try a simple suffix match.
				parts := strings.Split(*existingProduct.ImageURL, "/warung-assets/") // WARUNG_BUCKET_NAME default
				if len(parts) > 1 {
					key := parts[1]
					// Run delete in background or ignore error
					_ = h.minio.DeleteFile(r.Context(), key) 
				}
			}
		}

	} else {
		// Standard JSON body
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			response.BadRequest(w, "Invalid request body")
			return
		}
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

	// Invalidate products cache
	_ = h.cache.InvalidatePattern(r.Context(), "products:list:*")

	response.OK(w, "Product updated", product)
}

// ToggleActive toggles a product's active status
func (h *ProductHandler) ToggleActive(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid product ID")
		return
	}

	product, err := h.repo.ToggleActive(r.Context(), id)
	if err == domain.ErrNotFound {
		response.NotFound(w, "Product not found")
		return
	}
	if err != nil {
		response.InternalServerError(w, "Failed to toggle product status")
		return
	}

	// Invalidate products cache
	_ = h.cache.InvalidatePattern(r.Context(), "products:list:*")

	response.OK(w, "Product status toggled", product)
}

// Delete soft deletes a product
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid product ID")
		return
	}

	// Fetch product first to get image URL
	product, err := h.repo.GetByID(r.Context(), id)
	if err == domain.ErrNotFound {
		response.NotFound(w, "Product not found")
		return
	}
	if err != nil {
		response.InternalServerError(w, "Failed to fetch product")
		return
	}

	// Delete image if exists
	if product.ImageURL != nil {
		parts := strings.Split(*product.ImageURL, "/warung-assets/")
		if len(parts) > 1 {
			key := parts[1]
			_ = h.minio.DeleteFile(r.Context(), key)
		}
	}

	// Delete product
	if err := h.repo.Delete(r.Context(), id); err != nil {
		response.InternalServerError(w, "Failed to delete product")
		return
	}

	// Invalidate products cache
	_ = h.cache.InvalidatePattern(r.Context(), "products:list:*")

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

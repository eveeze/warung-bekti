package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/pkg/pdf"
	"github.com/eveeze/warung-backend/internal/pkg/response"
	"github.com/eveeze/warung-backend/internal/pkg/validator"
	"github.com/eveeze/warung-backend/internal/repository"
	"github.com/eveeze/warung-backend/internal/service"
)

// InventoryHandler handles inventory endpoints
type InventoryHandler struct {
	inventoryRepo *repository.InventoryRepository
	productRepo   *repository.ProductRepository
	cache         *service.CacheService
	event         *service.EventService
}

// NewInventoryHandler creates a new InventoryHandler
func NewInventoryHandler(
	inventoryRepo *repository.InventoryRepository,
	productRepo *repository.ProductRepository,
	cache *service.CacheService,
	event *service.EventService,
) *InventoryHandler {
	return &InventoryHandler{
		inventoryRepo: inventoryRepo,
		productRepo:   productRepo,
		cache:         cache,
		event:         event,
	}
}

// Restock adds stock to a product
func (h *InventoryHandler) Restock(w http.ResponseWriter, r *http.Request) {
	var input domain.RestockInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	v := validator.New()
	v.Custom("product_id", input.ProductID != uuid.Nil, "Product ID is required")
	v.Min("quantity", input.Quantity, 1, "Quantity must be at least 1")
	v.NonNegative("cost_per_unit", input.CostPerUnit, "Cost per unit cannot be negative")
	if v.HasErrors() {
		response.ValidationError(w, v.Errors())
		return
	}

	// Check if product exists
	_, err := h.productRepo.GetByID(r.Context(), input.ProductID)
	if err == domain.ErrNotFound {
		response.NotFound(w, "Product not found")
		return
	}
	if err != nil {
		response.InternalServerError(w, "Failed to get product")
		return
	}

	movement, err := h.inventoryRepo.Restock(r.Context(), input)
	if err != nil {
		response.InternalServerError(w, "Failed to restock product")
		return
	}

	// Invalidate cache
	// InvalidatePattern("products:list:*") - clears all product lists
	// InvalidatePattern("reports:*") - clears reports
	// Delete("product:{id}") - specific product cache if we have one (we might add it later)
	_ = h.cache.InvalidatePattern(r.Context(), "products:list:*")
	_ = h.cache.InvalidatePattern(r.Context(), "reports:*")

	// Publish real-time event
	// Data: { "product_id": ..., "new_stock": ... }
	// We need new stock. Movement usually returns it, or we can fetch/guess.
	// Movement struct usually has 'CurrentStock'. Let's assume so or fetch product.
	// For efficiency, let's just send the movement which usually contains relevant data or just fetch.
	// Actually, movement might be just the log. Let's fetch the product to be safe and accurate for UI updates.
	product, _ := h.productRepo.GetByID(r.Context(), input.ProductID)
	if product != nil {
		h.event.Publish(service.EventStockUpdate, map[string]interface{}{
			"product_id":    product.ID,
			"current_stock": product.CurrentStock,
			"is_low_stock":  product.CurrentStock <= product.MinStockAlert,
		})
	}

	response.Created(w, "Stock restocked successfully", movement)
}

// Adjust adjusts stock manually
func (h *InventoryHandler) Adjust(w http.ResponseWriter, r *http.Request) {
	var input domain.StockAdjustmentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	v := validator.New()
	v.Custom("product_id", input.ProductID != uuid.Nil, "Product ID is required")
	v.Required("reason", input.Reason, "Reason is required")
	if v.HasErrors() {
		response.ValidationError(w, v.Errors())
		return
	}

	// Check if product exists
	_, err := h.productRepo.GetByID(r.Context(), input.ProductID)
	if err == domain.ErrNotFound {
		response.NotFound(w, "Product not found")
		return
	}
	if err != nil {
		response.InternalServerError(w, "Failed to get product")
		return
	}

	movement, err := h.inventoryRepo.Adjust(r.Context(), input)
	if err != nil {
		response.InternalServerError(w, "Failed to adjust stock")
		return
	}

	// Invalidate cache
	_ = h.cache.InvalidatePattern(r.Context(), "products:list:*")
	_ = h.cache.InvalidatePattern(r.Context(), "reports:*")

	// Publish event
	product, _ := h.productRepo.GetByID(r.Context(), input.ProductID)
	if product != nil {
		h.event.Publish(service.EventStockUpdate, map[string]interface{}{
			"product_id":    product.ID,
			"current_stock": product.CurrentStock,
			"is_low_stock":  product.CurrentStock <= product.MinStockAlert,
		})
	}

	response.Created(w, "Stock adjusted successfully", movement)
}

// GetMovements retrieves stock movements for a product
func (h *InventoryHandler) GetMovements(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("productId")
	productID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid product ID")
		return
	}

	query := r.URL.Query()
	filter := domain.StockMovementFilter{ProductID: &productID, Page: 1, PerPage: 20}

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
		t := domain.StockMovementType(typeStr)
		filter.Type = &t
	}

	movements, total, err := h.inventoryRepo.GetByProduct(r.Context(), productID, filter)
	if err != nil {
		response.InternalServerError(w, "Failed to get stock movements")
		return
	}

	meta := response.NewMeta(filter.Page, filter.PerPage, total)
	response.SuccessWithMeta(w, http.StatusOK, "Stock movements retrieved", movements, meta)
}

// GetLowStock retrieves products with low stock
func (h *InventoryHandler) GetLowStock(w http.ResponseWriter, r *http.Request) {
	products, err := h.productRepo.GetLowStockProducts(r.Context())
	if err != nil {
		response.InternalServerError(w, "Failed to get low stock products")
		return
	}

	response.OK(w, "Low stock products retrieved", products)
}

// GetReport retrieves stock inventory report
func (h *InventoryHandler) GetReport(w http.ResponseWriter, r *http.Request) {
	report, err := h.inventoryRepo.GetStockReport(r.Context())
	if err != nil {
		response.InternalServerError(w, "Failed to get stock report")
		return
	}

	// Get low stock products for the report
	lowStock, _ := h.productRepo.GetLowStockProducts(r.Context())
	report.LowStockProducts = lowStock

	response.OK(w, "Stock report retrieved", report)
}

// DownloadRestockPDF generates and downloads PDF for low-stock products
func (h *InventoryHandler) DownloadRestockPDF(w http.ResponseWriter, r *http.Request) {
	// Get low stock products
	lowStockProducts, err := h.productRepo.GetLowStockProducts(r.Context())
	if err != nil {
		response.InternalServerError(w, "Failed to get low stock products")
		return
	}

	if len(lowStockProducts) == 0 {
		response.BadRequest(w, "No low stock products found")
		return
	}

	// Convert to RestockItems
	items := make([]pdf.RestockItem, len(lowStockProducts))
	for i, p := range lowStockProducts {
		items[i] = pdf.RestockItem{
			ProductName:  p.Product.Name,
			CurrentStock: p.Product.CurrentStock,
			MinStock:     p.Product.MinStockAlert,
			Deficit:      p.DeficitAmount,
			Unit:         p.Product.Unit,
			CostPrice:    p.Product.CostPrice,
		}
	}

	// Generate PDF
	pdfData := pdf.RestockData{
		StoreName:    "Warung Kelontong",
		StoreAddress: "Jl. Contoh No. 123",
		GeneratedAt:  time.Now(),
		Items:        items,
	}

	pdfDoc, err := pdf.GenerateRestockPDF(pdfData)
	if err != nil {
		response.InternalServerError(w, "Failed to generate PDF")
		return
	}

	// Set headers for PDF download
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=restock_list.pdf")

	// Output PDF to response
	if err := pdfDoc.Output(w); err != nil {
		response.InternalServerError(w, "Failed to output PDF")
		return
	}
}

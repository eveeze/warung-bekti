package handler

import (
	"net/http"
	"strconv"

	"github.com/eveeze/warung-backend/internal/pkg/response"
	"github.com/eveeze/warung-backend/internal/repository"
	"github.com/eveeze/warung-backend/internal/service"

	"github.com/google/uuid"
)

// PublicHandler handles public (unauthenticated) endpoints for the landing page
type PublicHandler struct {
	productRepo  *repository.ProductRepository
	categoryRepo *repository.CategoryRepository
	cache        *service.CacheService
}

// NewPublicHandler creates a new PublicHandler
func NewPublicHandler(
	productRepo *repository.ProductRepository,
	categoryRepo *repository.CategoryRepository,
	cache *service.CacheService,
) *PublicHandler {
	return &PublicHandler{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		cache:        cache,
	}
}

// PublicProduct is a sanitized product for public consumption (no cost_price, stock info, etc.)
type PublicProduct struct {
	ID           uuid.UUID          `json:"id"`
	Name         string             `json:"name"`
	Description  *string            `json:"description,omitempty"`
	Unit         string             `json:"unit"`
	BasePrice    int64              `json:"base_price"`
	ImageURL     *string            `json:"image_url,omitempty"`
	Category     *PublicCategory    `json:"category,omitempty"`
	PricingTiers []PublicPricingTier `json:"pricing_tiers,omitempty"`
}

// PublicPricingTier is a sanitized pricing tier
type PublicPricingTier struct {
	Name        *string `json:"name,omitempty"`
	MinQuantity int     `json:"min_quantity"`
	MaxQuantity *int    `json:"max_quantity,omitempty"`
	Price       int64   `json:"price"`
}

// PublicCategory is a sanitized category
type PublicCategory struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
}

// ListProducts returns active products for the public landing page
func (h *PublicHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	isActive := true
	filter := repository.PublicProductFilter{
		IsActive: &isActive,
		Page:     1,
		PerPage:  20,
		SortBy:   "name",
		SortOrder: "asc",
	}

	// Parse query parameters
	if search := r.URL.Query().Get("search"); search != "" {
		filter.Search = &search
	}
	if catID := r.URL.Query().Get("category_id"); catID != "" {
		if id, err := uuid.Parse(catID); err == nil {
			filter.CategoryID = &id
		}
	}
	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			filter.Page = p
		}
	}
	if perPage := r.URL.Query().Get("per_page"); perPage != "" {
		if pp, err := strconv.Atoi(perPage); err == nil && pp > 0 && pp <= 100 {
			filter.PerPage = pp
		}
	}

	products, total, err := h.productRepo.ListPublic(r.Context(), filter)
	if err != nil {
		response.InternalServerError(w, "Failed to get products")
		return
	}

	// Convert to public format (strip sensitive fields)
	publicProducts := make([]PublicProduct, len(products))
	for i, p := range products {
		pp := PublicProduct{
			ID:        p.ID,
			Name:      p.Name,
			Description: p.Description,
			Unit:      p.Unit,
			BasePrice: p.BasePrice,
			ImageURL:  p.ImageURL,
		}

		if p.Category != nil {
			pp.Category = &PublicCategory{
				ID:          p.Category.ID,
				Name:        p.Category.Name,
				Description: p.Category.Description,
			}
		}

		if len(p.PricingTiers) > 0 {
			pp.PricingTiers = make([]PublicPricingTier, 0)
			for _, tier := range p.PricingTiers {
				if tier.IsActive {
					pp.PricingTiers = append(pp.PricingTiers, PublicPricingTier{
						Name:        tier.Name,
						MinQuantity: tier.MinQuantity,
						MaxQuantity: tier.MaxQuantity,
						Price:       tier.Price,
					})
				}
			}
		}

		publicProducts[i] = pp
	}

	response.OK(w, "Products retrieved", map[string]interface{}{
		"products": publicProducts,
		"total":    total,
		"page":     filter.Page,
		"per_page": filter.PerPage,
	})
}

// ListCategories returns active categories for the public landing page
func (h *PublicHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categoryRepo.ListActive(r.Context())
	if err != nil {
		response.InternalServerError(w, "Failed to get categories")
		return
	}

	publicCategories := make([]PublicCategory, len(categories))
	for i, c := range categories {
		publicCategories[i] = PublicCategory{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
		}
	}

	response.OK(w, "Categories retrieved", publicCategories)
}
